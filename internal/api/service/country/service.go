package country

import (
	"context"
	"net/http"

	"github.com/vuduongtp/go-core/internal/model"
	"github.com/vuduongtp/go-core/pkg/database"
	"github.com/vuduongtp/go-core/pkg/server/apperr"
	structutil "github.com/vuduongtp/go-core/pkg/util/struct"

	"gorm.io/gorm"
)

// New creates new country application service
func New(db *gorm.DB, cdb database.Intf) *Country {
	return &Country{
		db:  db,
		cdb: cdb,
	}
}

// Country represents country application service
type Country struct {
	db  *gorm.DB
	cdb database.Intf
}

// Service represents country application interface
type Service interface {
	Create(context.Context, *model.AuthUser, CreationData) (*model.Country, error)
	View(context.Context, *model.AuthUser, int) (*model.Country, error)
	List(context.Context, *model.AuthUser, *database.ListQueryCondition, *int64) ([]*model.Country, error)
	Update(context.Context, *model.AuthUser, int, UpdateData) (*model.Country, error)
	Delete(context.Context, *model.AuthUser, int) error
}

// CreationData contains country data from json request
type CreationData struct {
	// example: Vietnam
	Name string `json:"name" validate:"required,min=3"`
	// example: vn
	Code string `json:"code" validate:"required,min=2,max=10"`
	// example: +84
	PhoneCode string `json:"phone_code" validate:"required,min=2,max=10"`
}

// UpdateData contains country data from json request
type UpdateData struct {
	// example: Vietnam
	Name *string `json:"name,omitempty" validate:"omitempty,min=3"`
	// example: vn
	Code *string `json:"code,omitempty" validate:"omitempty,min=2,max=10"`
	// example: +84
	PhoneCode *string `json:"phone_code,omitempty" validate:"omitempty,min=2,max=10"`
}

// ListResp contains list of paginated countries and total numbers of countries
type ListResp struct {
	// example: [{"id": 1, "created_at": "2020-01-14T10:03:41Z", "updated_at": "2020-01-14T10:03:41Z", "name": "Singapore", "code": "SG", "phone_code": "+65"}]
	Data []*model.Country `json:"data"`
	// example: 1
	TotalCount int64 `json:"total_count"`
}

// Custom errors
var (
	ErrCountryNotFound    = apperr.NewHTTPError(http.StatusBadRequest, "COUNTRY_NOTFOUND", "Country not found")
	ErrCountryNameExisted = apperr.NewHTTPValidationError("Country name already exists")
)

// Create creates a new country
func (s *Country) Create(ctx context.Context, authUsr *model.AuthUser, data CreationData) (*model.Country, error) {
	if existed, err := s.cdb.Exist(ctx, s.db, map[string]interface{}{"name": data.Name}); err != nil || existed {
		return nil, ErrCountryNameExisted.SetInternal(err)
	}

	rec := &model.Country{
		Name:      data.Name,
		Code:      data.Code,
		PhoneCode: data.PhoneCode,
	}
	if err := s.cdb.Create(ctx, s.db, rec); err != nil {
		return nil, apperr.NewHTTPInternalError("Error creating country").SetInternal(err)
	}

	return rec, nil
}

// View returns single country
func (s *Country) View(ctx context.Context, authUsr *model.AuthUser, id int) (*model.Country, error) {
	rec := new(model.Country)
	if err := s.cdb.View(ctx, s.db, rec, id); err != nil {
		return nil, ErrCountryNotFound.SetInternal(err)
	}

	return rec, nil
}

// List returns list of countrys
func (s *Country) List(ctx context.Context, authUsr *model.AuthUser, lq *database.ListQueryCondition, count *int64) ([]*model.Country, error) {
	var data []*model.Country
	if err := s.cdb.List(ctx, s.db, &data, lq, count); err != nil {
		return nil, apperr.NewHTTPInternalError("Error listing country").SetInternal(err)
	}

	return data, nil
}

// Update updates country information
func (s *Country) Update(ctx context.Context, authUsr *model.AuthUser, id int, data UpdateData) (*model.Country, error) {
	if existed, err := s.cdb.Exist(ctx, s.db, map[string]interface{}{"name": data.Name, "id__notexact": id}); err != nil || existed {
		return nil, ErrCountryNameExisted.SetInternal(err)
	}

	// optimistic update
	updates := structutil.ToMap(data)
	if err := s.cdb.Update(ctx, s.db, updates, id); err != nil {
		return nil, apperr.NewHTTPInternalError("Error updating country").SetInternal(err)
	}

	rec := new(model.Country)
	if err := s.cdb.View(ctx, s.db, rec, id); err != nil {
		return nil, ErrCountryNotFound.SetInternal(err)
	}

	return rec, nil
}

// Delete deletes a country
func (s *Country) Delete(ctx context.Context, authUsr *model.AuthUser, id int) error {
	if existed, err := s.cdb.Exist(ctx, s.db, id); err != nil || !existed {
		return ErrCountryNotFound.SetInternal(err)
	}

	if err := s.cdb.Delete(ctx, s.db, id); err != nil {
		return apperr.NewHTTPInternalError("Error deleting country").SetInternal(err)
	}

	return nil
}
