package user

import (
	"context"
	"net/http"

	"github.com/kevinxvu/goone/internal/model"
	"github.com/kevinxvu/goone/pkg/database"
	"github.com/kevinxvu/goone/pkg/server/apperr"
	"github.com/kevinxvu/goone/pkg/util/crypter"
	structutil "github.com/kevinxvu/goone/pkg/util/struct"

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

// Service represents user application interface
type Service interface {
	Create(context.Context, *model.AuthUser, CreationData) (*model.User, error)
	View(context.Context, *model.AuthUser, int) (*model.User, error)
	List(context.Context, *model.AuthUser, *database.ListQueryCondition, *int64) ([]*model.User, error)
	Update(context.Context, *model.AuthUser, int, UpdateData) (*model.User, error)
	Delete(context.Context, *model.AuthUser, int) error
	Me(context.Context, *model.AuthUser) (*model.User, error)
	ChangePassword(context.Context, *model.AuthUser, PasswordChangeData) error
}

// CreationData contains user data from json request
type CreationData struct {
	Username  string `json:"username" validate:"required,min=3"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Mobile    string `json:"mobile" validate:"required,mobile"`
	Role      string `json:"role" validate:"required"`
	Blocked   bool   `json:"blocked"`
}

// UpdateData contains user data from json request
type UpdateData struct {
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Email     *string `json:"email,omitempty" validate:"omitempty,email"`
	Mobile    *string `json:"mobile,omitempty" validate:"omitempty,mobile"`
	Role      *string `json:"role,omitempty"`
	Blocked   *bool   `json:"blocked,omitempty"`
}

// PasswordChangeData contains password change request
type PasswordChangeData struct {
	OldPassword        string `json:"old_password" validate:"required"`
	NewPassword        string `json:"new_password" validate:"required,min=8"`
	NewPasswordConfirm string `json:"new_password_confirm" validate:"required,eqfield=NewPassword"`
}

// ListResp contains list of users and current page number response
type ListResp struct {
	Data       []*model.User `json:"data"`
	TotalCount int64         `json:"total_count"`
}

// Custom errors
var (
	ErrIncorrectPassword = apperr.NewHTTPError(http.StatusBadRequest, "INCORRECT_PASSWORD", "Incorrect old password")
	ErrUserNotFound      = apperr.NewHTTPError(http.StatusBadRequest, "USER_NOTFOUND", "User not found")
	ErrUsernameExisted   = apperr.NewHTTPValidationError("Username already existed")
)

// Create creates a new user account
func (s *User) Create(ctx context.Context, authUsr *model.AuthUser, data CreationData) (*model.User, error) {
	if existed, err := s.udb.Exist(ctx, s.db, map[string]interface{}{"username": data.Username}); err != nil || existed {
		return nil, ErrUsernameExisted.SetInternal(err)
	}

	rec := &model.User{
		FirstName: data.FirstName,
		LastName:  data.LastName,
		Email:     data.Email,
		Mobile:    data.Mobile,
		Username:  data.Username,
		Password:  crypter.HashPassword(data.Password),
		Blocked:   data.Blocked,
		Role:      data.Role,
	}

	if err := s.udb.Create(ctx, s.db, rec); err != nil {
		return nil, apperr.NewHTTPInternalError("Error creating user").SetInternal(err)
	}

	return rec, nil
}

// View returns single user
func (s *User) View(ctx context.Context, authUsr *model.AuthUser, id int) (*model.User, error) {
	rec := new(model.User)
	if err := s.udb.View(ctx, s.db, rec, id); err != nil {
		return nil, ErrUserNotFound.SetInternal(err)
	}

	return rec, nil
}

// List returns list of users
func (s *User) List(ctx context.Context, authUsr *model.AuthUser, lq *database.ListQueryCondition, count *int64) ([]*model.User, error) {
	var data []*model.User
	if err := s.udb.List(ctx, s.db, &data, lq, count); err != nil {
		return nil, apperr.NewHTTPInternalError("Error listing user").SetInternal(err)
	}

	return data, nil
}

// Update updates user information
func (s *User) Update(ctx context.Context, authUsr *model.AuthUser, id int, data UpdateData) (*model.User, error) {
	// optimistic update
	updates := structutil.ToMap(data)
	if err := s.udb.Update(ctx, s.db, updates, id); err != nil {
		return nil, apperr.NewHTTPInternalError("Error updating user").SetInternal(err)
	}

	rec := new(model.User)
	if err := s.udb.View(ctx, s.db, rec, id); err != nil {
		return nil, ErrUserNotFound.SetInternal(err)
	}

	return rec, nil
}

// Delete deletes a user
func (s *User) Delete(ctx context.Context, authUsr *model.AuthUser, id int) error {
	if existed, err := s.udb.Exist(ctx, s.db, id); err != nil || !existed {
		return ErrUserNotFound.SetInternal(err)
	}

	if err := s.udb.Delete(ctx, s.db, id); err != nil {
		return apperr.NewHTTPInternalError("Error deleting user").SetInternal(err)
	}

	return nil
}

// Me returns authenticated user
func (s *User) Me(ctx context.Context, authUsr *model.AuthUser) (*model.User, error) {
	rec := new(model.User)
	if err := s.udb.View(ctx, s.db, rec, authUsr.ID); err != nil {
		return nil, ErrUserNotFound.SetInternal(err)
	}
	return rec, nil
}

// ChangePassword changes authenticated user password
func (s *User) ChangePassword(ctx context.Context, authUsr *model.AuthUser, data PasswordChangeData) error {
	rec, err := s.Me(ctx, authUsr)
	if err != nil {
		return err
	}

	if !crypter.CompareHashAndPassword(rec.Password, data.OldPassword) {
		return ErrIncorrectPassword
	}

	hashedPwd := crypter.HashPassword(data.NewPassword)
	if err = s.udb.Update(ctx, s.db, map[string]interface{}{"password": hashedPwd}, rec.ID); err != nil {
		return apperr.NewHTTPInternalError("Error changing password").SetInternal(err)
	}

	return nil
}
