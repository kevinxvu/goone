# GoCore Development Guide

## Architecture Overview

This is a clean architecture Go API starter kit using **Echo v4** and **GORM**. The service supports PostgreSQL, MySQL/MariaDB, and SQLite.

**Current Setup**: Uses MariaDB 12 via Docker Compose on port 3306.

**Layer Flow**: `cmd/` (entry points) → `internal/` (business logic) → `pkg/` (reusable utilities)
- **cmd/api**: HTTP server initialization and dependency injection
- **internal/api**: Domain modules (user, auth, country) - each has http.go, service.go, model
- **internal/repository**: Repository implementations with custom queries (flat structure)
- **pkg/database**: Database connection, base repository with CRUD operations
- **pkg/server**: Echo server setup, middleware, validators, error handling
- **pkg/util**: Shared utilities (crypter, cfg, swagger, etc.)

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

### 2. Dependency Injection with Wire

This project uses **Google Wire** for compile-time dependency injection. All dependencies are wired in [internal/di/wire.go](internal/di/wire.go).

**Wire Workflow:**

1. **Define Provider Functions** in `internal/di/wire.go`:
```go
//go:build wireinject
// +build wireinject

package di

import "github.com/google/wire"

// Provider function - creates and returns a service
func ProvideUserService(db *gorm.DB, userDB *repository.UserRepository, crypterSvc *crypter.Service) user.Service {
    return user.New(db, userDB, crypterSvc)
}

// Injector function - tells Wire what to build
func InitializeApplication() (*Application, error) {
    wire.Build(
        ProvideConfig,
        ProvideDB,
        ProvideUserDB,
        ProvideCrypter,
        ProvideUserService,
        // ... other providers
        wire.Struct(new(Application), "*"),
    )
    return nil, nil  // Wire replaces this
}
```

2. **Generate Wire Code**:
```bash
# Using Makefile (recommended):
make wire

# Or run directly from project root:
cd internal/di && wire

# Or use GOFLAGS if using vendor:
cd internal/di && GOFLAGS=-mod=mod wire
```

This creates `wire_gen.go` with the actual `InitializeApplication()` implementation.

3. **Use in main.go**:
```go
import "github.com/vuduongtp/go-core/internal/di"

func main() {
    app, err := di.InitializeApplication()  // Wire-generated function
    checkErr(err)
    
    // Use app.Config, app.DB, app.Server, etc.
}
```

**Adding New Dependencies:**
1. Create a `ProvideXXX()` function in `internal/di/wire.go`
2. Add it to `wire.Build()` list
3. Add field to `Application` struct if needed
4. Run `make wire` to regenerate `wire_gen.go`

**Important:** 
- Provider functions are in `internal/di/wire.go` (build tag: `wireinject`)
- Generated code is in `internal/di/wire_gen.go` (build tag: `!wireinject`)
- Never edit `wire_gen.go` manually
- Use concrete types in provider signatures for better type safety
- See [internal/di/README.md](internal/di/README.md) for detailed Wire documentation

### 3. Database Repository Pattern

All repositories are in `internal/repository/` with a **flat structure**. Each repository uses `package repository` and embeds `*database.DB` for CRUD operations.

```go
// internal/repository/user_repository.go
package repository

import (
    "github.com/vuduongtp/go-core/internal/model"
    "github.com/vuduongtp/go-core/pkg/database"
)

func NewUserRepository() *UserRepository {
    return &UserRepository{database.NewDB(model.User{})}
}

type UserRepository struct {
    *database.DB
}

// Custom query method
func (d *UserRepository) FindByUsername(ctx context.Context, db *gorm.DB, uname string) (*model.User, error) {
    rec := new(model.User)
    if err := d.View(ctx, db, rec, "username = ?", uname); err != nil {
        return nil, err
    }
    return rec, nil
}
```

**Key Points:**
- All repositories use `package repository` (flat structure in one folder)
- Repository types are named with entity prefix: `UserRepository`, `CountryRepository`, `ProductRepository`
- Factory methods: `NewUserRepository()`, `NewCountryRepository()`, etc.
- Embed `*database.DB` which provides: `Create`, `View`, `List`, `Update`, `Delete`, `Exist`, `CreateInBatches`
- Import from `pkg/database` (not `pkg/util/db`)

**Adding New Repository:**
1. Create `{entity}_repository.go` in `internal/repository/`
2. Use `package repository`
3. Name struct `{Entity}Repository` and factory `New{Entity}Repository()`
4. Add to DI providers in `internal/di/wire.go`
5. Run `make wire`

### 4. Models MUST Use Custom Base

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

### 5. Authentication Flow

JWT middleware extracts `model.AuthUser` from token. Access in handlers:
```go
func (h *HTTP) me(c echo.Context) error {
    au := h.auth.User(c) // Returns *model.AuthUser (ID, Username, Email, Role)
    // ... use au
}
```

Auth endpoints (`/login`, `/refresh-token`) are in [internal/api/auth](internal/api/auth) - no JWT required.

Protected routes use `v1Router.Use(jwtSvc.MWFunc())` - see [cmd/api/main.go](cmd/api/main.go#L90-L95).

### 6. Custom Validators

Register custom validators in [pkg/server/validator.go](pkg/server/validator.go#L15-L20). Available:
- `validate:"mobile"` - Phone number format `^(\+\d{1,3})?\s?\d{5,15}$`
- `validate:"date"` - Date format `YYYY-MM-DD` or with `T00:00:00Z`

### 7. Migrations

Migration files use **gormigrate** in [internal/functions/migration/main.go](internal/functions/migration/main.go#L67-L101). 

**Critical**: Copy model structs inside migration functions to prevent schema drift. Migration IDs are timestamps (e.g., `201905051012`).

For MySQL tables, use `tx.Set("gorm:table_options", defaultTableOpts)` to set `ENGINE=InnoDB ROW_FORMAT=DYNAMIC`.

### 8. Logging

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

# Generate Wire DI code (after modifying internal/di/wire.go)
make wire

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

- **Package imports**: `database` (from pkg/database), `logging` (from pkg/logging), `repository` (from internal/repository)
- **Error handling**: Use `server.NewHTTPError(statusCode, code, message)` for API errors
- **Logging**: Use `logging.FromContext(ctx)` with zap fields, not printf-style logging in production code
- **Context propagation**: Always pass `context.Context` through service and DB layers to preserve request tracking
- **Repository access**: Import `"github.com/vuduongtp/go-core/internal/repository"` and use `repository.UserRepository`, `repository.CountryRepository`

## AWS Lambda Support

The codebase supports AWS Lambda deployment (see `functions/` and `internal/functions/`). Migrations can run as Lambda functions using apex deployment scripts in [Makefile](Makefile#L68-L88).
