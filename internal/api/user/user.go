package user

import (
	"context"
	"net/http"

	"github.com/vuduongtp/go-core/internal/model"
	"github.com/vuduongtp/go-core/pkg/server"
	dbutil "github.com/vuduongtp/go-core/pkg/util/db"
	structutil "github.com/vuduongtp/go-core/pkg/util/struct"
)

// Custom errors
var (
	ErrIncorrectPassword = server.NewHTTPError(http.StatusBadRequest, "INCORRECT_PASSWORD", "Incorrect old password")
	ErrUserNotFound      = server.NewHTTPError(http.StatusBadRequest, "USER_NOTFOUND", "User not found")
	ErrUsernameExisted   = server.NewHTTPValidationError("Username already existed")
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
		Password:  s.cr.HashPassword(data.Password),
		Blocked:   data.Blocked,
		Role:      data.Role,
	}

	if err := s.udb.Create(ctx, s.db, rec); err != nil {
		return nil, server.NewHTTPInternalError("Error creating user").SetInternal(err)
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
func (s *User) List(ctx context.Context, authUsr *model.AuthUser, lq *dbutil.ListQueryCondition, count *int64) ([]*model.User, error) {
	var data []*model.User
	if err := s.udb.List(ctx, s.db, &data, lq, count); err != nil {
		return nil, server.NewHTTPInternalError("Error listing user").SetInternal(err)
	}

	return data, nil
}

// Update updates user information
func (s *User) Update(ctx context.Context, authUsr *model.AuthUser, id int, data UpdateData) (*model.User, error) {
	// optimistic update
	updates := structutil.ToMap(data)
	if err := s.udb.Update(ctx, s.db, updates, id); err != nil {
		return nil, server.NewHTTPInternalError("Error updating user").SetInternal(err)
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
		return server.NewHTTPInternalError("Error deleting user").SetInternal(err)
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

	if !s.cr.CompareHashAndPassword(rec.Password, data.OldPassword) {
		return ErrIncorrectPassword
	}

	hashedPwd := s.cr.HashPassword(data.NewPassword)
	if err = s.udb.Update(ctx, s.db, map[string]interface{}{"password": hashedPwd}, rec.ID); err != nil {
		return server.NewHTTPInternalError("Error changing password").SetInternal(err)
	}

	return nil
}
