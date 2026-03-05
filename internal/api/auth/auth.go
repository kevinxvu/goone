package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/vuduongtp/go-core/internal/model"
	"github.com/vuduongtp/go-core/pkg/server/apperr"
	"github.com/vuduongtp/go-core/pkg/util/crypter"
)

// Custom errors
var (
	ErrInvalidCredentials  = apperr.NewHTTPError(http.StatusUnauthorized, "INVALID_CREDENTIALS", "Username or password is incorrect")
	ErrUserBlocked         = apperr.NewHTTPError(http.StatusUnauthorized, "USER_BLOCKED", "Your account has been blocked and may not login")
	ErrInvalidRefreshToken = apperr.NewHTTPError(http.StatusUnauthorized, "INVALID_REFRESH_TOKEN", "Invalid refresh token")
)

// LoginUser logs in the given user, returns access token
func (s *Auth) LoginUser(ctx context.Context, u *model.User) (*model.AuthToken, error) {
	claims := map[string]interface{}{
		"id":       u.ID,
		"username": u.Username,
		"email":    u.Email,
		"role":     u.Role,
	}
	token, expiresin, err := s.jwt.GenerateToken(claims, nil)
	if err != nil {
		return nil, apperr.NewHTTPInternalError("Error generating token").SetInternal(err)
	}

	refreshToken := crypter.UID()
	err = s.udb.Update(ctx, s.db, map[string]interface{}{"refresh_token": refreshToken, "last_login": time.Now()}, u.ID)
	if err != nil {
		return nil, apperr.NewHTTPInternalError("Error updating user").SetInternal(err)
	}

	return &model.AuthToken{AccessToken: token, TokenType: "bearer", ExpiresIn: expiresin, RefreshToken: refreshToken}, nil
}

// Authenticate tries to authenticate the user provided by given credentials
func (s *Auth) Authenticate(ctx context.Context, data Credentials) (*model.AuthToken, error) {
	usr, err := s.udb.FindByUsername(ctx, s.db, data.Username)
	if err != nil || usr == nil {
		return nil, ErrInvalidCredentials.SetInternal(err)
	}
	if !crypter.CompareHashAndPassword(usr.Password, data.Password) {
		return nil, ErrInvalidCredentials
	}
	if usr.Blocked {
		return nil, ErrUserBlocked
	}

	return s.LoginUser(ctx, usr)
}

// RefreshToken returns the new access token with expired time extended
func (s *Auth) RefreshToken(ctx context.Context, data RefreshTokenData) (*model.AuthToken, error) {
	usr, err := s.udb.FindByRefreshToken(ctx, s.db, data.RefreshToken)
	if err != nil || usr == nil {
		return nil, ErrInvalidRefreshToken.SetInternal(err)
	}
	return s.LoginUser(ctx, usr)
}

// User returns user data stored in jwt token
func (s *Auth) User(c echo.Context) *model.AuthUser {
	id, _ := c.Get("id").(float64)
	user, _ := c.Get("username").(string)
	email, _ := c.Get("email").(string)
	role, _ := c.Get("role").(string)
	return &model.AuthUser{
		ID:       int(id),
		Username: user,
		Email:    email,
		Role:     role,
	}
}
