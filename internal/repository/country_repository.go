package repository

import (
	"context"

	"github.com/vuduongtp/go-core/internal/model"
	"github.com/vuduongtp/go-core/pkg/database"

	"gorm.io/gorm"
)

// NewCountryRepository returns a new country database instance
func NewCountryRepository() *CountryRepository {
	return &CountryRepository{database.NewDB(model.Country{})}
}

// CountryRepository represents the client for country table
type CountryRepository struct {
	*database.DB
}

// FindByCode queries for single country by code
func (d *CountryRepository) FindByCode(ctx context.Context, db *gorm.DB, code string) (*model.Country, error) {
	rec := new(model.Country)
	if err := d.View(ctx, db, rec, "code = ?", code); err != nil {
		return nil, err
	}
	return rec, nil
}
