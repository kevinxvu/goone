# Wire Dependency Injection Guide

## Tổng quan

Dự án này sử dụng [Google Wire](https://github.com/google/wire) để quản lý Dependency Injection (DI) một cách tự động và type-safe tại compile-time.

Tất cả code liên quan đến DI được tổ chức trong thư mục `internal/di` để tách biệt khỏi business logic và entry points.

## Cách thức hoạt động

Wire hoạt động bằng cách:
1. Đọc các **provider functions** trong `wire.go` (có build tag `wireinject`)
2. Phân tích dependency graph và tự động sắp xếp thứ tự khởi tạo
3. Generate code trong `wire_gen.go` (có build tag `!wireinject`)
4. Khi build application, chỉ `wire_gen.go` được compile, `wire.go` bị skip

## Cấu trúc files

```
internal/di/
├── wire.go          # Khai báo providers (chỉ dùng cho wire tool)
├── wire_gen.go      # Generated code (được wire tạo ra)
└── README.md        # Tài liệu này

cmd/api/
└── main.go          # Entry point (import internal/di)
```

## Provider Functions trong wire.go

**Location:** `internal/di/wire.go`

### 1. Provider cơ bản

```go
// ProvideConfig loads configuration
func ProvideConfig() (*config.Configuration, error) {
    return config.Load()
}
```

Provider này:
- Không có dependencies (không nhận tham số)
- Trả về `*config.Configuration` và `error`
- Wire sẽ tự động check error

### 2. Provider có dependencies

```go
// ProvideDB initializes database connection
func ProvideDB(cfg *config.Configuration) (*gorm.DB, error) {
    return dbutil.New(cfg.DbType, cfg.DbDsn, cfg.DbLog)
}
```

Provider này:
- Cần `cfg *config.Configuration` (Wire tự động inject từ ProvideConfig)
- Trả về `*gorm.DB` và `error`

### 3. Provider với concrete types

```go
// ProvideUserDB creates user database repository
func ProvideUserDB() *userdb.DB {
    return userdb.NewDB()
}

// ProvideCrypter creates crypter service
func ProvideCrypter() *crypter.Service {
    return crypter.New()
}
```

**Quan trọng:** Sử dụng **concrete types** (`*userdb.DB`, `*crypter.Service`) thay vì interfaces để Wire có thể check type chính xác hơn.

### 4. Provider cho services

```go
// ProvideUserService creates user service
func ProvideUserService(db *gorm.DB, userDB *userdb.DB, crypterSvc *crypter.Service) user.Service {
    return user.New(db, userDB, crypterSvc)
}
```

Provider này:
- Nhận 3 dependencies (Wire tự động inject)
- Trả về `user.Service` interface

## Application Struct

**Location:** `internal/di/wire.go`

```go
package di

type Application struct {
    Config      *config.Configuration
    DB          *gorm.DB
    Server      *echo.Echo
    AuthSvc     auth.Service
    UserSvc     user.Service
    CountrySvc  country.Service
    JWTSvc      auth.JWT
}
```

Struct này chứa tất cả dependencies cần thiết cho application. Wire sẽ tự động fill tất cả các fields.

## Injector Function

```go
func InitializeApplication() (*Application, error) {
    wire.Build(
        ProvideConfig,
        ProvideDB,
        ProvideUserDB,
        ProvideCountryDB,
        ProvideCrypter,
        ProvideJWT,
        ProvideAuthService,
        ProvideUserService,
        ProvideCountryService,
        ProvideServer,
        wire.Struct(new(Application), "*"),
    )
    return nil, nil  // Wire sẽ thay thế implementation này
}
```

- `wire.Build()` liệt kê tất cả providers
- `wire.Struct(new(Application), "*")` yêu cầu Wire fill tất cả fields của Application
- Body function (`return nil, nil`) sẽ bị Wire thay thế hoàn toàn

## Generate Wire Code

### Sử dụng Makefile (Khuyến nghị):

```bash
# Generate wire code
make wire

# Check nếu wire code cần regenerate
make wire.check
```

### Hoặc chạy trực tiếp từ project root:

```bash
cd internal/di && wire
```

### Nếu dùng vendor:

```bash
cd internal/di && GOFLAGS=-mod=mod wire
```

### Output:

```
wire: github.com/kevinxvu/goone/internal/di: wrote /home/duongvu/github/go-core/internal/di/wire_gen.go
```

## Sử dụng trong main.go

```go
package main

import (
    "github.com/kevinxvu/goone/internal/di"
)

func main() {
    // Initialize application với Wire DI
    app, err := di.InitializeApplication()
    checkErr(err)

    // Configure logging (không thể Wire vì ProvideLogger không return value)
    logLevel := zapcore.InfoLevel
    if app.Config.Debug {
        logLevel = zapcore.DebugLevel
    }
    logging.SetConfig(&logging.Config{
        Level:      logLevel,
        FilePath:   "logs/app.log",
        TimeFormat: "2006-01-02 15:04:05",
    })

    // Sử dụng các dependencies từ app
    sqlDB, err := app.DB.DB()
    checkErr(err)
    defer sqlDB.Close()

    // Setup routes
    auth.NewHTTP(app.AuthSvc, app.Server)
    v1Router := app.Server.Group("/v1")
    v1Router.Use(app.JWTSvc.MWFunc())
    user.NewHTTP(app.UserSvc, app.AuthSvc, v1Router.Group("/users"))
    country.NewHTTP(app.CountrySvc, app.AuthSvc, v1Router.Group("/countries"))

    // Start server
    server.Start(app.Server, app.Config.Stage == "development")
}
```

## Thêm Dependency Mới

### Bước 1: Tạo Provider Function

**File:** `internal/di/wire.go`

Ví dụ thêm Email service:

```go
// ProvideEmailService creates email service
func ProvideEmailService(cfg *config.Configuration) *email.Service {
    return email.New(cfg.AwsRegion, cfg.EmailFrom)
}
```

### Bước 2: Thêm vào wire.Build()

```go
func InitializeApplication() (*Application, error) {
    wire.Build(
        ProvideConfig,
        ProvideDB,
        // ... existing providers
        ProvideEmailService,  // ← Thêm vào đây
        wire.Struct(new(Application), "*"),
    )
    return nil, nil
}
```

### Bước 3: Thêm field vào Application (nếu cần)

```go
type Application struct {
    Config      *config.Configuration
    DB          *gorm.DB
    Server      *echo.Echo
    EmailSvc    *email.Service  // ← Thêm field mới
    // ... other fields
}
```

### Bước 4: Regenerate Wire

```bash
# Sử dụng Makefile
make wire

# Hoặc chạy trực tiếp
cd internal/di && GOFLAGS=-mod=mod wire
```

### Bước 5: Sử dụng trong main.go

```go
package main

import (
    "github.com/kevinxvu/goone/internal/di"
)

func main() {
    app, err := di.InitializeApplication()
    checkErr(err)
    
    // Sử dụng email service
    err = app.EmailSvc.Send("user@example.com", "Subject", "Body")
    checkErr(err)
}
```

## Debug Wire Errors

### Error: "unused provider"

```
wire: inject InitializeApplication: unused provider "di.ProvideLogger"
```

**Nguyên nhân:** Provider không được sử dụng bởi bất kỳ field nào của Application hoặc provider khác.

**Giải pháp:** 
- Xóa provider khỏi `wire.Build()`
- Hoặc thêm field vào Application struct
- Hoặc dùng provider đó như dependency của provider khác

### Error: "undefined type"

```
wire: wire.go:53:33: undefined: dbutil.DB
```

**Nguyên nhân:** Import path không đúng hoặc thiếu import.

**Giải pháp:** Kiểm tra import:
```go
import (
    pkgdb "github.com/kevinxvu/goone/pkg/util/db"
)

func ProvideCountryDB() *pkgdb.DB {  // Dùng alias đúng
    return country.NewDB()
}
```

### Error: "could not import"

```
wire: wire.go:7:2: could not import github.com/google/wire (invalid package name: "")
```

**Nguyên nhân:** Wire tool không tìm thấy dependencies trong vendor.

**Giải pháp:** Chạy với `GOFLAGS=-mod=mod`:
```bash
# Sử dụng Makefile (đã set GOFLAGS tự động)
make wire

# Hoặc chạy trực tiếp
cd internal/di && GOFLAGS=-mod=mod wire
```

## Best Practices

### 1. Sử dụng Concrete Types cho Providers

✅ **Tốt:**
```go
func ProvideUserDB() *userdb.DB {
    return userdb.NewDB()
}
```

❌ **Tránh:**
```go
func ProvideUserDB() user.MyDB {  // Interface
    return userdb.NewDB()
}
```

**Lý do:** Concrete types giúp Wire check dependencies chính xác hơn.

### 2. Provider trả về Interface cho Services

✅ **Tốt:**
```go
func ProvideUserService(db *gorm.DB, userDB *userdb.DB, crypterSvc *crypter.Service) user.Service {
    return user.New(db, userDB, crypterSvc)
}
```

**Lý do:** Services nên return interface để dễ test và mock.

### 3. Không dùng Wire cho Side Effects

❌ **Tránh trong `internal/di/wire.go`:**
```go
func ProvideLogger(cfg *config.Configuration) error {
    logging.SetConfig(&logging.Config{...})
    return nil  // Chỉ có side effect, không return value hữu ích
}
```

✅ **Tốt hơn:** Gọi trực tiếp trong `cmd/api/main.go`:
```go
func main() {
    app, err := InitializeApplication()
    checkErr(err)
    
    logging.SetConfig(&logging.Config{...})  // Setup logging trực tiếp
}
```

### 4. Group Providers theo Module

```go
wire.Build(
    // Config & Infrastructure
    ProvideConfig,
    ProvideDB,
    ProvideServer,
    
    // Repositories
    ProvideUserDB,
    ProvideCountryDB,
    
    // Utilities
    ProvideCrypter,
    ProvideJWT,
    
    // Services
    ProvideAuthService,
    ProvideUserService,
    ProvideCountryService,
    
    // Application
    wire.Struct(new(Application), "*"),
)
```

## Testing với Wire

Wire không support testing directly. Để test, có 2 cách:

### Cách 1: Tạo test providers

Tạo `wire_test.go` với providers riêng cho testing:

```go
//go:build wireinject
// +build wireinject

package main

func InitializeTestApplication() (*Application, error) {
    wire.Build(
        ProvideTestConfig,  // Mock config
        ProvideDB,
        // ... other providers
    )
    return nil, nil
}
```

### Cách 2: Manual DI trong tests

```go
func TestUserService(t *testing.T) {
    db := setupTestDB()
    userDB := userdb.NewDB()
    crypterSvc := crypter.New()
    
    userSvc := user.New(db, userDB, crypterSvc)
    
    // Test userSvc...
}
```

## Tham khảo

- [Wire User Guide](https://github.com/google/wire/blob/main/docs/guide.md)
- [Wire Best Practices](https://github.com/google/wire/blob/main/docs/best-practices.md)
- [Wire FAQ](https://github.com/google/wire/blob/main/docs/faq.md)
