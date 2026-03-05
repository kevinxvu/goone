# GoCore Development Guide

## Architecture Overview

This is a clean architecture Go API starter kit using **Echo v4** and **GORM**. The service supports PostgreSQL, MySQL/MariaDB, and SQLite.

**Current Setup**: Uses MariaDB 12 via Docker Compose on port 3306.

**Layer Flow**: `cmd/` (entry points) → `internal/` (business logic) → `pkg/` (reusable utilities)
- **cmd/api**: HTTP server initialization and dependency injection
- **cmd/migration**: Database migration tool using Goose
- **internal/api**: Layered API structure with clear separation of concerns:
  - **docs**: Swagger/OpenAPI generated documentation (docs.go, swagger.json, swagger.yaml)
  - **service/{module}**: Business logic, DTOs, interfaces (e.g., service/user/service.go)
  - **handler/{module}**: HTTP handlers, route registration (e.g., handler/user/handler.go)
  - **router**: Centralized route configuration (router/router.go)
- **internal/migrations**: SQL migration files managed by Goose (YYYYMMDDHHMMSS_description.sql)
- **internal/repository**: Repository implementations with custom queries (flat structure)
- **pkg/database**: Database connection, base repository with CRUD operations
- **pkg/server**: Echo server setup with subpackages:
  - **apperr**: Application error types and HTTP error handler
  - **binder**: Custom request binder and validators
  - **middleware**: JWT, logger, secure middlewares
- **pkg/util**: Shared utilities:
  - **request**: HTTP request helpers (ReqID, ReqListQuery, etc.)
  - **crypter**: Password hashing, UID generation (static functions)
  - **config**: Configuration loader (OS env vars have priority over .env files)
  - **swagger**: Swagger types and utilities
  - **migration**: Database migration utilities using Goose (SQL-based migrations)
- **pkg/aws**: AWS service wrappers:
  - **email**: SES email service (send email, raw email with attachments)
  - **s3**: S3 service (presigned URLs, file operations)
  - **sns**: SNS push notification service (iOS APNS, Android FCM)
  - **sqs**: SQS message queue service

## Code Patterns

### 1. Creating New API Modules

Follow the **Service → Handler → Router** pattern. The architecture uses a layered approach with clear separation:

**Service Layer** (`internal/api/service/{module}/service.go`):
```go
package user

import (
    "context"
    "github.com/vuduongtp/go-core/internal/model"
    "github.com/vuduongtp/go-core/pkg/database"
    "gorm.io/gorm"
)

// Factory function
func New(db *gorm.DB, udb MyDB) *User {
    return &User{db: db, udb: udb}
}

// Service struct
type User struct {
    db  *gorm.DB
    udb MyDB
}

// Repository interface (consumed by service)
type MyDB interface {
    database.Intf
    FindByUsername(context.Context, *gorm.DB, string) (*model.User, error)
}

// Service interface (exposed to handler)
type Service interface {
    Create(context.Context, *model.AuthUser, CreationData) (*model.User, error)
    View(context.Context, *model.AuthUser, int) (*model.User, error)
    // ... more methods
}

// DTOs
type CreationData struct {
    Username string `json:"username" validate:"required,min=3"`
    Password string `json:"password" validate:"required,min=8"`
    // ... more fields
}

type UpdateData struct {
    Username *string `json:"username,omitempty"`
    // ... more fields
}

// Custom errors
var (
    ErrUserNotFound = apperr.NewHTTPError(http.StatusBadRequest, "USER_NOTFOUND", "User not found")
)

// Business logic methods
func (s *User) Create(ctx context.Context, authUsr *model.AuthUser, data CreationData) (*model.User, error) {
    hashedPassword := crypter.HashPassword(data.Password)
    rec := &model.User{
        Username: data.Username,
        Password: hashedPassword,
    }
    if err := s.udb.Create(ctx, s.db, rec); err != nil {
        return nil, apperr.NewHTTPInternalError("Error creating user").SetInternal(err)
    }
    return rec, nil
}
```

**Handler Layer** (`internal/api/handler/{module}/handler.go`):
```go
package user

import (
    "net/http"
    "github.com/labstack/echo/v4"
    "github.com/vuduongtp/go-core/internal/api/service/user"
    "github.com/vuduongtp/go-core/internal/model"
)

// HTTP struct
type HTTP struct {
    svc  user.Service
    auth model.Auth
}

// Route registration
func NewHTTP(svc user.Service, auth model.Auth, eg *echo.Group) {
    h := HTTP{svc, auth}
    eg.POST("", h.create)
    eg.GET("/:id", h.view)
    // ... more routes
}

// Handler methods
func (h *HTTP) create(c echo.Context) error {
    r := user.CreationData{}
    if err := c.Bind(&r); err != nil {
        return err
    }
    resp, err := h.svc.Create(c.Request().Context(), h.auth.User(c), r)
    if err != nil {
        return err
    }
    return c.JSON(http.StatusOK, resp)
}
```

**Router Layer** (`internal/api/router/router.go`):
```go
package router

import (
    "github.com/vuduongtp/go-core/internal/api/handler/auth"
    "github.com/vuduongtp/go-core/internal/api/handler/user"
    "github.com/vuduongtp/go-core/internal/di"
)

func RegisterRoutes(app *di.Application) {
    // Auth routes (no JWT middleware)
    auth.NewHTTP(app.AuthSvc, app.Server)
    
    // Protected v1 routes with JWT middleware
    v1Router := app.Server.Group("/v1")
    v1Router.Use(app.JWT.MWFunc())
    
    // Register module routes
    user.NewHTTP(app.UserSvc, app.Auth, v1Router.Group("/users"))
}
```

**Adding a New Module:**
1. Create `internal/api/service/{module}/service.go` with DTOs, interfaces, and business logic
2. Create `internal/api/handler/{module}/handler.go` with HTTP handlers
3. Add provider functions to `internal/di/wire.go`
4. Register routes in `internal/api/router/router.go`
5. Run `make wire` to regenerate DI code

### 2. Dependency Injection with Wire

This project uses **Google Wire** for compile-time dependency injection. All dependencies are wired in [internal/di/wire.go](internal/di/wire.go).

**Wire Workflow:**

1. **Define Provider Functions** in `internal/di/wire.go`:
```go
//go:build wireinject
// +build wireinject

package di

import (
    "github.com/google/wire"
    userSvc "github.com/vuduongtp/go-core/internal/api/service/user"
    authSvc "github.com/vuduongtp/go-core/internal/api/service/auth"
    "github.com/vuduongtp/go-core/internal/repository"
)

// Provider function - creates and returns a service
func ProvideUserService(db *gorm.DB, userDB *repository.UserRepository) userSvc.Service {
    return userSvc.New(db, userDB)
}

// Application struct
type Application struct {
    Config     *config.Configuration
    DB         *gorm.DB
    Server     *echo.Echo
    JWT        *jwt.Service
    Auth       model.Auth
    AuthSvc    authSvc.Service
    UserSvc    userSvc.Service
    CountrySvc countrySvc.Service
}

// Injector function - tells Wire what to build
func InitializeApplication() (*Application, error) {
    wire.Build(
        ProvideConfig,
        ProvideDB,
        ProvideUserDB,
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

Auth endpoints (`/login`, `/refresh-token`) are in [internal/api/service/auth](internal/api/service/auth) - no JWT required.

Protected routes use `v1Router.Use(app.JWT.MWFunc())` - see [internal/api/router/router.go](internal/api/router/router.go).

### 6. Custom Validators

Register custom validators in `pkg/server/binder/validator.go`. Available:
- `validate:"mobile"` - Phone number format `^(\+\d{1,3})?\s?\d{5,15}$`
- `validate:"date"` - Date format `YYYY-MM-DD` or with `T00:00:00Z`

**Usage in structs:**
```go
type UserData struct {
    Mobile string `json:"mobile" validate:"required,mobile"`
    DOB    string `json:"dob" validate:"required,date"`
}
```

### 7. Migrations

Database migrations use **Goose** (https://github.com/pressly/goose) with SQL migration files. Migration files are stored in [internal/migrations/](internal/migrations/) and executed via [cmd/migration/main.go](cmd/migration/main.go).

> **Note**: This project uses Goose for SQL-based migrations. The migration history is tracked in the `goose_db_version` table.

**Migration File Format:**
- Naming: `YYYYMMDDHHMMSS_description.sql` (e.g., `20240101000001_create_users_table.sql`)
- Each file must have `-- +goose Up` and `-- +goose Down` sections
- Wrap DDL statements in `-- +goose StatementBegin` and `-- +goose StatementEnd`

**Example Migration:**
```sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL DEFAULT NULL,
    INDEX idx_users_deleted_at (deleted_at)
) ENGINE=InnoDB ROW_FORMAT=DYNAMIC DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
```

**Common Commands:**
```bash
# Run all pending migrations
make migrate

# Check migration status
make migrate.status

# Rollback last migration
make migrate.undo

# Create new migration
make migrate.create name=add_users_table

# Show current version
make migrate.version

# Advanced commands
go run cmd/migration/main.go up-to 20240101000001  # Migrate to specific version
go run cmd/migration/main.go reset                 # Rollback all migrations
go run cmd/migration/main.go redo                  # Redo last migration
```

**Best Practices:**
- Use `IF NOT EXISTS` and `IF EXISTS` for idempotent migrations
- Always test both Up and Down migrations
- Keep migrations small and focused (one logical change per file)
- Never modify existing migration files - create new ones instead
- For MySQL: Always use `ENGINE=InnoDB ROW_FORMAT=DYNAMIC` and `DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`
- Track migration history in `goose_db_version` table

**Programmatic Usage:**
```go
import (
    "github.com/vuduongtp/go-core/pkg/util/migration"
)

cfg := &migration.GooseConfig{
    Dir:       "internal/migrations",
    TableName: "goose_db_version",
    Dialect:   "mysql",
    Verbose:   true,
}
err := migration.RunGoose(db, cfg)
```

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

# Migration commands
make migrate                          # Run all pending migrations
make migrate.status                   # Check migration status
make migrate.undo                     # Rollback last migration
make migrate.version                  # Show current version
make migrate.create name=xxx          # Create new migration
make migrate.reset                    # Rollback all migrations
make migrate.redo                     # Redo last migration

# Run tests with coverage
make test.cover
```

The API runs on `http://localhost:8080`. Swagger docs at `/docs/index.html`.

**Database**: MariaDB 12 runs in Docker on `localhost:3306` (credentials in `.env`).

Default credentials: username `superadmin`, password `superadmin123!@#`

## Configuration

Environment variables are loaded via [pkg/util/config/config.go](pkg/util/config/config.go) with the following priority order:
1. **OS environment variables** (highest priority - never overwritten)
2. `.env.local` (gitignored, for local overrides)
3. `.env` (committed defaults, lowest priority)

**Important**: OS environment variables always take precedence. This allows deployment platforms (Docker, Kubernetes, cloud services) to override local .env files without conflicts.

Key variables in [config/config.go](config/config.go#L11-L29): `STAGE`, `PORT`, `DB_TYPE`, `DB_DSN`, `JWT_SECRET`, `JWT_DURATION`, `JWT_ALGORITHM`, `ALLOW_ORIGINS`.

**Database Configuration:**
- **Type**: `mysql` (for MariaDB/MySQL)
- **DSN Format**: `username:password@tcp(host:port)/database?charset=utf8mb4&parseTime=True&loc=Local`
- **Default**: `goone:goone123@tcp(localhost:3306)/goone?charset=utf8mb4&parseTime=True&loc=Local`
- **Docker**: MariaDB 12 container on port 3306, database `goone`, user `goone`, password `goone123`

For PostgreSQL: Use `postgres://user:pass@host:port/db?sslmode=disable`  
For SQLite: Use file path like `./test.db`

## Common Utilities

### Package Structure & Imports

**Core Packages:**
- `database` - Import from `pkg/database` for base repository operations
- `logging` - Import from `pkg/logging` for structured logging
- `repository` - Import from `internal/repository` and use `repository.UserRepository`, etc.
- `apperr` - Import from `pkg/server/apperr` for HTTP error handling
- `request` - Import from `pkg/util/request` for HTTP request utilities
- `binder` - Import from `pkg/server/binder` for validators and binders
- `crypter` - Import from `pkg/util/crypter` for password hashing (static functions)
- `config` - Import from `pkg/util/config` as `cfgutil` for configuration loading
- `migration` - Import from `pkg/util/migration` for database migrations

**API Layer Packages:**
- `service` - Import service packages with aliases: `userSvc "github.com/vuduongtp/go-core/internal/api/service/user"`
- `handler` - Import handler packages: `"github.com/vuduongtp/go-core/internal/api/handler/user"`
- `router` - Import router: `"github.com/vuduongtp/go-core/internal/api/router"`
- DTOs are in service packages: Use `user.CreationData`, `user.UpdateData`, etc.
- Service interfaces: Use `user.Service`, `auth.Service`, etc.

**AWS Service Packages:**
- `email` - Import from `pkg/aws/email` for SES email service
- `s3` - Import from `pkg/aws/s3` for S3 file storage
- `sns` - Import from `pkg/aws/sns` as `snsutil` for push notifications
- `sqs` - Import from `pkg/aws/sqs` as `sqsutil` for message queues

**Error Handling:**
```go
import "github.com/vuduongtp/go-core/pkg/server/apperr"

// Create custom errors
var ErrUserNotFound = apperr.NewHTTPError(http.StatusBadRequest, "USER_NOTFOUND", "User not found")
var ErrInvalidInput = apperr.NewHTTPValidationError("Invalid input data")
var ErrInternal = apperr.NewHTTPInternalError("Internal server error")

// Use in handlers
if err != nil {
    return ErrUserNotFound.SetInternal(err)
}
```

**Request Utilities:**
```go
import "github.com/vuduongtp/go-core/pkg/util/request"

// Get ID from URL parameter
id, err := request.ReqID(c)

// Parse list query parameters (pagination, filter, sort)
lq, err := request.ReqListQuery(c)

// String helpers
email := request.TrimSpacePointer(data.Email)
mobile := request.RemoveSpacePointer(data.Mobile)
```

**Logging:**
- Use `logging.FromContext(ctx)` to get logger with request context (includes correlation ID)
- Use Zap fields for structured logging, not printf-style in production code
- Example: `logging.FromContext(ctx).Info("user created", zap.Int("user_id", user.ID))`

**Context Propagation:**
- Always pass `context.Context` through service and DB layers to preserve request tracking
- Use `c.Request().Context()` to get context from Echo handler

**Crypter (Static Package):**
```go
import "github.com/vuduongtp/go-core/pkg/util/crypter"

// All functions are static - no instantiation required
hashedPassword := crypter.HashPassword(plainPassword)
isValid := crypter.CompareHashAndPassword(hashedPassword, plainPassword)
uid := crypter.UID()
```

**Configuration Loading:**
```go
import cfgutil "github.com/vuduongtp/go-core/pkg/util/config"

cfg := new(Configuration)
if err := cfgutil.LoadConfig(cfg, appName, stage); err != nil {
    return nil, err
}
// OS environment variables override .env files automatically
```

**Swagger Annotations:**
When using `swagger.ListRequest` in Swagger comments, reference it as `ListRequest` (not `swagger.ListRequest`):
```go
// @Param q query ListRequest false "QueryListRequest"
```

For model types with `@name` annotations (User, AuthToken, Country), use the short name without the `model.` prefix:
```go
// Correct:
// @Success 200 {object} User
// @Success 200 {object} AuthToken

// Incorrect:
// @Success 200 {object} model.User  
// @Success 200 {object} model.AuthToken
```

For service DTOs, use the qualified package name:
```go
// @Param request body user.CreationData true "CreationData"
// @Success 200 {object} user.ListResp
```
