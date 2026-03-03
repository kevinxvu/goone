package jwt

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/vuduongtp/go-core/internal/model"
	"github.com/vuduongtp/go-core/pkg/logging"
	"github.com/vuduongtp/go-core/pkg/server"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
)

// New generates new JWT service necessery for auth middleware
func New(algo, secret string, duration int) *Service {
	signingMethod := jwt.GetSigningMethod(algo)
	if signingMethod == nil {
		panic("invalid jwt signing method")
	}
	return &Service{
		algo:     signingMethod,
		key:      []byte(secret),
		duration: time.Duration(duration) * time.Second,
	}
}

// Service provides a Json-Web-Token authentication implementation
type Service struct {
	// Secret key used for signing.
	key []byte
	// Duration (in seconds) for which the jwt token is valid.
	duration time.Duration
	// Service signing algorithm
	algo jwt.SigningMethod
}

// MWFunc makes JWT implement the Middleware interface.
func (j *Service) MWFunc() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			token, err := j.ParseTokenFromHeader(c)
			if err != nil || !token.Valid {
				if err != nil {
					logging.FromContext(c.Request().Context()).Sugar().Errorf("error parsing token: %+v", err.Error())
				}
				return server.NewHTTPError(http.StatusUnauthorized, "UNAUTHORIZED", "Your session is unauthorized or has expired.").SetInternal(err)
			}

			claims := token.Claims.(jwt.MapClaims)
			for key, val := range claims {
				c.Set(key, val)
			}

			return next(c)
		}
	}
}

// ParseTokenFromHeader parses token from Authorization header
func (j *Service) ParseTokenFromHeader(c echo.Context) (*jwt.Token, error) {
	token := c.Request().Header.Get("Authorization")
	if token == "" {
		return nil, fmt.Errorf("token not found")
	}
	parts := strings.SplitN(token, " ", 2)
	if !(len(parts) == 2 && strings.ToLower(parts[0]) == "bearer") {
		return nil, fmt.Errorf("token invalid")
	}

	return j.ParseToken(parts[1])
}

// ParseToken parses token from string
func (j *Service) ParseToken(input string) (*jwt.Token, error) {
	return jwt.Parse(input, func(token *jwt.Token) (interface{}, error) {
		if j.algo != token.Method {
			return nil, fmt.Errorf("token method mismatched")
		}
		return j.key, nil
	})
}

// GenerateToken generates new Service token and populates it with user data
func (j *Service) GenerateToken(claims map[string]interface{}, expire *time.Time) (string, int, error) {
	if expire == nil {
		expTime := time.Now().Add(j.duration)
		expire = &expTime
	}
	claims["exp"] = expire.Unix()

	token := jwt.NewWithClaims(j.algo, jwt.MapClaims(claims))
	tokenString, err := token.SignedString(j.key)

	return tokenString, int(time.Until(*expire).Seconds()), err
}

// User returns user data stored in jwt token
func (j *Service) User(c echo.Context) *model.AuthUser {
	id, _ := c.Get("id").(float64)
	username, _ := c.Get("username").(string)
	email, _ := c.Get("email").(string)
	role, _ := c.Get("role").(string)

	return &model.AuthUser{
		ID:       int(id),
		Username: username,
		Email:    email,
		Role:     role,
	}
}
