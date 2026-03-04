package auth

import (
	"context"
	"time"

	"github.com/vuduongtp/go-core/internal/model"
	"github.com/vuduongtp/go-core/pkg/database"

	"gorm.io/gorm"
)

// New creates new auth service
func New(db *gorm.DB, udb UserDB, jwt JWT, cr Crypter) *Auth {
	return &Auth{
		db:  db,
		udb: udb,
		jwt: jwt,
		cr:  cr,
	}
}

// Auth represents auth application service
type Auth struct {
	db  *gorm.DB
	udb UserDB
	jwt JWT
	cr  Crypter
}

// UserDB represents user repository interface
type UserDB interface {
	database.Intf
	FindByUsername(context.Context, *gorm.DB, string) (*model.User, error)
	FindByRefreshToken(context.Context, *gorm.DB, string) (*model.User, error)
}

// JWT represents token generator (jwt) interface
type JWT interface {
	GenerateToken(map[string]interface{}, *time.Time) (string, int, error)
}

// Crypter represents security interface
type Crypter interface {
	CompareHashAndPassword(string, string) bool
	UID() string
}
