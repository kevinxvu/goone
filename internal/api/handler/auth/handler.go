package auth

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/vuduongtp/go-core/internal/api/service/auth"
)

// HTTP represents auth http service
type HTTP struct {
	svc auth.Service
}

// NewHTTP creates new auth http service
func NewHTTP(svc auth.Service, e *echo.Echo) {
	h := HTTP{svc}

	e.POST("/login", h.login)
	e.POST("/refresh-token", h.refreshToken)
}

// @Summary		Logs in user by username and password
// @Description	Logs in user by username and password
// @Accept			json
// @Produce		json
// @Tags			auth
// @ID				authLogin
// @Param			request	body		auth.Credentials	true	"Credentials"
// @Success		200		{object}	AuthToken
// @Failure		401		{object}	SwaggErrDetailsResp
// @Failure		500		{object}	SwaggErrDetailsResp
// @Router			/login [post]
func (h *HTTP) login(c echo.Context) error {
	r := auth.Credentials{}
	if err := c.Bind(&r); err != nil {
		return err
	}
	resp, err := h.svc.Authenticate(c.Request().Context(), r)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

// @Summary		Refresh access token
// @Description	Refresh access token
// @Accept			json
// @Produce		json
// @Tags			auth
// @ID				authRefreshToken
// @Param			request	body		auth.RefreshTokenData	true	"RefreshTokenData"
// @Success		200		{object}	AuthToken
// @Failure		401		{object}	SwaggErrDetailsResp
// @Failure		500		{object}	SwaggErrDetailsResp
// @Router			/refresh-token [post]
func (h *HTTP) refreshToken(c echo.Context) error {
	r := auth.RefreshTokenData{}
	if err := c.Bind(&r); err != nil {
		return err
	}
	resp, err := h.svc.RefreshToken(c.Request().Context(), r)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}
