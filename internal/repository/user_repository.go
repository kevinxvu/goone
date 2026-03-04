package repository

import (
	"context"

	"github.com/vuduongtp/go-core/internal/model"
	"github.com/vuduongtp/go-core/pkg/database"

	"gorm.io/gorm"
)

// NewUserRepository returns a new user database instance
func NewUserRepository() *UserRepository {
	return &UserRepository{database.NewDB(model.User{})}
}

// UserRepository represents the client for user table
type UserRepository struct {
	*database.DB
}

// FindByUsername queries for single user by username
func (d *UserRepository) FindByUsername(ctx context.Context, db *gorm.DB, uname string) (*model.User, error) {
	rec := new(model.User)
	if err := d.View(ctx, db, rec, "username = ?", uname); err != nil {
		return nil, err
	}
	return rec, nil
}

// FindByRefreshToken queries for single user by refresh token
func (d *UserRepository) FindByRefreshToken(ctx context.Context, db *gorm.DB, token string) (*model.User, error) {
	rec := new(model.User)
	if err := d.View(ctx, db, rec, "refresh_token = ?", token); err != nil {
		return nil, err
	}
	return rec, nil
}
