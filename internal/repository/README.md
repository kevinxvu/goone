# Repository Layer

This package contains all database repository implementations for the application's data access layer.

## Structure

All repositories are in a flat structure within this package:
- `user_repository.go` - User entity repository
- `country_repository.go` - Country entity repository

## Pattern

Each repository:
1. Uses `package repository` 
2. Embeds `*database.DB` to inherit common CRUD operations
3. Named with entity prefix: `UserRepository`, `CountryRepository`
4. Created via constructor: `NewUserRepository()`, `NewCountryRepository()`
5. Implements custom query methods as needed

## Usage

```go
import "github.com/kevinxvu/goone/internal/repository"

// In DI/Wire
userRepo := repository.NewUserRepository()
countryRepo := repository.NewCountryRepository()
```

All repositories implement the `database.Intf` interface for standard operations.

## Adding New Repositories

When adding a new entity repository:
1. Create `{entity}_repository.go` in this package
2. Use `package repository`
3. Name the struct `{Entity}Repository` (e.g., `ProductRepository`)
4. Create constructor `New{Entity}Repository()`
5. Embed `*database.DB` with the model
6. Add custom query methods with receiver `(*{Entity}Repository)`

Example:
```go
package repository

import (
    "github.com/kevinxvu/goone/internal/model"
    "github.com/kevinxvu/goone/pkg/database"
)

func NewProductRepository() *ProductRepository {
    return &ProductRepository{database.NewDB(model.Product{})}
}

type ProductRepository struct {
    *database.DB
}

func (d *ProductRepository) FindBySKU(ctx context.Context, db *gorm.DB, sku string) (*model.Product, error) {
    // custom query implementation
}
```
