package user

import (
	"context"

	"github.com/vuduongtp/go-core/internal/model"
	"github.com/vuduongtp/go-core/pkg/database"

	"gorm.io/gorm"
)

// New creates new user application service
func New(db *gorm.DB, udb MyDB) *User {
	return &User{db: db, udb: udb}
}

// User represents user application service
type User struct {
	db  *gorm.DB
	udb MyDB
}

// MyDB represents user repository interface
type MyDB interface {
	database.Intf
	FindByUsername(context.Context, *gorm.DB, string) (*model.User, error)
}
