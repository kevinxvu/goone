# GoCore Development Guide

## Architecture Overview

This is a clean architecture Go API starter kit using **Echo v4** and **GORM**. The service supports PostgreSQL, MySQL/MariaDB, and SQLite.

**Current Setup**: Uses MariaDB 12 via Docker Compose on port 3306.

**Layer Flow**: `cmd/` (entry points) → `internal/` (business logic) → `pkg/` (reusable utilities)
- **cmd/api**: HTTP server initialization and dependency injection
- **internal/api**: Domain modules (user, auth, country) - each has http.go, service.go, model
- **internal/db**: Repository implementations with custom queries
- **pkg/server**: Echo server setup, middleware, validators, error handling
- **pkg/util**: Shared utilities (db, crypter, cfg, swagger, etc.)

## Code Patterns

### 1. Creating New API Modules

Follow the **HTTP → Service → DB** pattern. Example from `internal/api/user/`:

```go
// http.go - Define HTTP layer and routes
type HTTP struct {
    svc  Service
    auth model.Auth
}

func NewHTTP(svc Service, auth model.Auth, eg *echo.Group) {
    h := HTTP{svc, auth}
    eg.POST("", h.create)
    eg.GET("/:id", h.view)
    // ... more routes
}

// service.go - Business logic
type User struct {
    db   *gorm.DB
    udb  MyDB
    cr   Crypter
}

func New(db *gorm.DB, udb MyDB, cr Crypter) *User {
    return &User{db: db, udb: udb, cr: cr}
}
```

Wire dependencies in [cmd/api/main.go](cmd/api/main.go#L75-L95).

### 2. Database Repository Pattern

All repositories embed `*dbutil.DB` for CRUD operations. Add custom queries as methods:

```go
// internal/db/user/db.go
type DB struct {
    *dbutil.DB
}

func NewDB() *DB {
    return &DB{dbutil.NewDB(model.User{})}
}

func (d *DB) FindByUsername(ctx context.Context, db *gorm.DB, uname string) (*model.User, error) {
    rec := new(model.User)
    if err := d.View(ctx, db, rec, "username = ?", uname); err != nil {
        return nil, err
    }
    return rec, nil
}
```

The `dbutil.DB` provides: `Create`, `View`, `List`, `Update`, `Delete`, `Exist`, `CreateInBatches`.

### 3. Models MUST Use Custom Base

**Never use `gorm.Model`**. All models must embed [internal/model/base.go](internal/model/base.go#L11-L20):

```go
type Base struct {
    ID        int            `json:"id" gorm:"primary_key"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}
```

Why? This project uses `int` IDs instead of `uint` (gorm.Model default).

### 4. Authentication Flow

JWT middleware extracts `model.AuthUser` from token. Access in handlers:
```go
func (h *HTTP) me(c echo.Context) error {
    au := h.auth.User(c) // Returns *model.AuthUser (ID, Username, Email, Role)
    // ... use au
}
```

Auth endpoints (`/login`, `/refresh-token`) are in [internal/api/auth](internal/api/auth) - no JWT required.

Protected routes use `v1Router.Use(jwtSvc.MWFunc())` - see [cmd/api/main.go](cmd/api/main.go#L90-L95).

### 5. Custom Validators

Register custom validators in [pkg/server/validator.go](pkg/server/validator.go#L15-L20). Available:
- `validate:"mobile"` - Phone number format `^(\+\d{1,3})?\s?\d{5,15}$`
- `validate:"date"` - Date format `YYYY-MM-DD` or with `T00:00:00Z`

### 6. Migrations

Migration files use **gormigrate** in [internal/functions/migration/main.go](internal/functions/migration/main.go#L67-L101). 

**Critical**: Copy model structs inside migration functions to prevent schema drift. Migration IDs are timestamps (e.g., `201905051012`).

For MySQL tables, use `tx.Set("gorm:table_options", defaultTableOpts)` to set `ENGINE=InnoDB ROW_FORMAT=DYNAMIC`.

### 7. Logging

The project uses **Uber Zap** for structured logging with JSON format output.

**Setup in main.go:**
```go
logging.SetConfig(&logging.Config{
    Level:      zapcore.InfoLevel,  // or DebugLevel in development
    FilePath:   "logs/app.log",     // enables file logging with rotation
    TimeFormat: "2006-01-02 15:04:05",
})
```

**Usage patterns:**
```go
// Get logger from context (includes request ID, correlation ID)
logger := logging.FromContext(ctx)
logger.Info("user created", zap.Int("user_id", user.ID))

// Component-specific logger
logger := logging.Component("auth")
logger.Warn("invalid token", zap.String("token", token))

// Default logger (not recommended in handlers)
logging.DefaultLogger().Error("critical error", zap.Error(err))

// Sugar API for printf-style logging
logging.DefaultLogger().Sugar().Infof("Processing %d items", count)
```

**HTTP Request Logging:**
The `pkg/server/middleware/logger` middleware automatically:
- Generates `X-Correlation-ID` and `X-Request-ID` for each request
- Logs request/response details (IP, method, path, status, latency, size)
- Stores logger in context with tracking IDs
- Outputs structured JSON format

**Available helpers:**
- `logging.FromContext(ctx)` - Get logger from context
- `logging.Component("name")` - Create logger with component field
- `logging.ErrField(err)` - Helper for error fields
- `logging.NewEchoLogger()` - Echo framework logger adapter
- `logging.NewGormLogger()` - GORM logger adapter

**Output format:** JSON to both console (stdout) and file with rotation (MaxSize: 10MB, MaxBackups: 3, MaxAge: 15 days, Compress: true).

## Development Workflow

```bash
# First time setup (docker + db migration)
make provision

# Run with hot reload (uses air)
make dev

# Generate Swagger docs (required after API changes)
make specs

# Run migrations manually
make migrate

# Undo last migration
make migrate.undo

# Run tests with coverage
make test.cover
```

The API runs on `http://localhost:8080`. Swagger docs at `/docs/index.html`.

**Database**: MariaDB 12 runs in Docker on `localhost:3306` (credentials in `.env`).

Default credentials: username `superadmin`, password `superadmin123!@#`

## Configuration

Environment variables are loaded via [pkg/util/cfg/cfg.go](pkg/util/cfg/cfg.go#L30-L50):
1. `.env.local` (gitignored, for local overrides)
2. `.env` (committed defaults)

Key variables in [config/config.go](config/config.go#L11-L29): `STAGE`, `PORT`, `DB_TYPE`, `DB_DSN`, `JWT_SECRET`, `JWT_DURATION`, `JWT_ALGORITHM`, `ALLOW_ORIGINS`.

**Database Configuration:**
- **Type**: `mysql` (for MariaDB/MySQL)
- **DSN Format**: `username:password@tcp(host:port)/database?charset=utf8mb4&parseTime=True&loc=Local`
- **Default**: `goone:goone123@tcp(localhost:3306)/goone?charset=utf8mb4&parseTime=True&loc=Local`
- **Docker**: MariaDB 12 container on port 3306, database `goone`, user `goone`, password `goone123`

For PostgreSQL: Use `postgres://user:pass@host:port/db?sslmode=disable`  
For SQLite: Use file path like `./test.db`

## Common Utilities

- **Package aliases**: Import as `dbutil`, `cfgutil`, `swaggerutil`, `httputil`, and `logging` (see existing imports)
- **Error handling**: Use `server.NewHTTPError(statusCode, code, message)` for API errors
- **Logging**: Use `logging.FromContext(ctx)` with zap fields, not printf-style logging in production code
- **Context propagation**: Always pass `context.Context` through service and DB layers to preserve request tracking

## AWS Lambda Support

The codebase supports AWS Lambda deployment (see `functions/` and `internal/functions/`). Migrations can run as Lambda functions using apex deployment scripts in [Makefile](Makefile#L68-L88).
