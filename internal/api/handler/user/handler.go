package user

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/vuduongtp/go-core/internal/api/service/user"
	"github.com/vuduongtp/go-core/internal/model"
	"github.com/vuduongtp/go-core/pkg/util/request"
)

// HTTP represents user http service
type HTTP struct {
	svc  user.Service
	auth model.Auth
}

// NewHTTP creates new user http service
func NewHTTP(svc user.Service, auth model.Auth, eg *echo.Group) {
	h := HTTP{svc, auth}

	eg.POST("", h.create)
	eg.GET("/:id", h.view)
	eg.GET("", h.list)
	eg.PATCH("/:id", h.update)
	eg.DELETE("/:id", h.delete)
	eg.GET("/me", h.me)
	eg.PATCH("/me/password", h.changePassword)
}

// @Security		BearerToken
// @Summary		Creates new user
// @Description	The new user
// @Accept			json
// @Produce		json
// @Tags			users
// @ID				usersCreate
// @Param			request		body		user.CreationData	true	"CreationData"
// @Success		200			{object}	model.User
// @Failure		400			{object}	SwaggErrDetailsResp
// @Failure		401			{object}	SwaggErrDetailsResp
// @Failure		403			{object}	SwaggErrDetailsResp
// @Failure		500			{object}	SwaggErrDetailsResp
// @Router			/v1/users	[post]
func (h *HTTP) create(c echo.Context) error {
	r := user.CreationData{}
	if err := c.Bind(&r); err != nil {
		return err
	}
	r.Email = strings.TrimSpace(r.Email)
	r.FirstName = strings.TrimSpace(r.FirstName)
	r.LastName = strings.TrimSpace(r.LastName)
	r.Mobile = strings.TrimSpace(strings.Replace(r.Mobile, " ", "", -1))
	r.Role = strings.TrimSpace(r.Role)

	resp, err := h.svc.Create(c.Request().Context(), h.auth.User(c), r)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

// @Security		BearerToken
// @Summary		Returns a single user
// @Description	Returns a single user
// @Accept			json
// @Produce		json
// @Tags			users
// @ID				usersView
// @Param			id				path		int	true	"User ID"
// @Success		200				{object}	model.User
// @Failure		400				{object}	SwaggErrDetailsResp
// @Failure		401				{object}	SwaggErrDetailsResp
// @Failure		403				{object}	SwaggErrDetailsResp
// @Failure		404				{object}	SwaggErrDetailsResp
// @Failure		500				{object}	SwaggErrDetailsResp
// @Router			/v1/users/{id}	[get]
func (h *HTTP) view(c echo.Context) error {
	id, err := request.ReqID(c)
	if err != nil {
		return err
	}
	resp, err := h.svc.View(c.Request().Context(), h.auth.User(c), id)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

// @Security		BearerToken
// @Summary		Get list user
// @Description	Get list user
// @Accept			json
// @Produce		json
// @Tags			users
// @ID				usersList
// @Param			q			query		ListRequest	false	"QueryListRequest"
// @Success		200			{object}	user.ListResp
// @Failure		400			{object}	SwaggErrDetailsResp
// @Failure		401			{object}	SwaggErrDetailsResp
// @Failure		403			{object}	SwaggErrDetailsResp
// @Failure		500			{object}	SwaggErrDetailsResp
// @Router			/v1/users	[get]
func (h *HTTP) list(c echo.Context) error {
	lq, err := request.ReqListQuery(c)
	if err != nil {
		return err
	}
	var count int64 = 0
	resp, err := h.svc.List(c.Request().Context(), h.auth.User(c), lq, &count)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, user.ListResp{Data: resp, TotalCount: count})
}

// @Security		BearerToken
// @Summary		Updates user information
// @Description	Updates user information
// @Accept			json
// @Produce		json
// @Tags			users
// @ID				usersUpdate
// @Param			id				path		int				true	"User ID"
// @Param			request			body		user.UpdateData	true	"UpdateData"
// @Success		200				{object}	model.User
// @Failure		400				{object}	SwaggErrDetailsResp
// @Failure		401				{object}	SwaggErrDetailsResp
// @Failure		403				{object}	SwaggErrDetailsResp
// @Failure		404				{object}	SwaggErrDetailsResp
// @Failure		500				{object}	SwaggErrDetailsResp
// @Router			/v1/users/{id}	[patch]
func (h *HTTP) update(c echo.Context) error {
	id, err := request.ReqID(c)
	if err != nil {
		return err
	}
	r := user.UpdateData{}
	if err := c.Bind(&r); err != nil {
		return err
	}
	r.Email = request.TrimSpacePointer(r.Email)
	r.FirstName = request.TrimSpacePointer(r.FirstName)
	r.LastName = request.TrimSpacePointer(r.LastName)
	r.Mobile = request.RemoveSpacePointer(r.Mobile)
	r.Role = request.RemoveSpacePointer(r.Role)

	resp, err := h.svc.Update(c.Request().Context(), h.auth.User(c), id, r)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

// @Security		BearerToken
// @Summary		Deletes an user
// @Description	Deletes an user
// @Accept			json
// @Produce		json
// @Tags			users
// @ID				usersDelete
// @Param			id				path		int	true	"User ID"
// @Success		200				{object}	SwaggOKResp
// @Failure		400				{object}	SwaggErrDetailsResp
// @Failure		401				{object}	SwaggErrDetailsResp
// @Failure		403				{object}	SwaggErrDetailsResp
// @Failure		404				{object}	SwaggErrDetailsResp
// @Failure		500				{object}	SwaggErrDetailsResp
// @Router			/v1/users/{id}	[delete]
func (h *HTTP) delete(c echo.Context) error {
	id, err := request.ReqID(c)
	if err != nil {
		return err
	}
	if err := h.svc.Delete(c.Request().Context(), h.auth.User(c), id); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

// @Security		BearerToken
// @Summary		Returns authenticated user
// @Description	Returns authenticated user
// @Accept			json
// @Produce		json
// @Tags			users
// @ID				usersMe
// @Success		200				{object}	model.User
// @Failure		400				{object}	SwaggErrDetailsResp
// @Failure		401				{object}	SwaggErrDetailsResp
// @Failure		403				{object}	SwaggErrDetailsResp
// @Failure		404				{object}	SwaggErrDetailsResp
// @Failure		500				{object}	SwaggErrDetailsResp
// @Router			/v1/users/me	[get]
func (h *HTTP) me(c echo.Context) error {
	resp, err := h.svc.Me(c.Request().Context(), h.auth.User(c))
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

// @Security		BearerToken
// @Summary		Changes authenticated user password
// @Description	Changes authenticated user password
// @Accept			json
// @Produce		json
// @Tags			users
// @ID				usersChangePwd
// @Param			request				body		user.PasswordChangeData	true	"PasswordChangeData"
// @Success		200					{object}	SwaggOKResp
// @Failure		400					{object}	SwaggErrDetailsResp
// @Failure		401					{object}	SwaggErrDetailsResp
// @Failure		403					{object}	SwaggErrDetailsResp
// @Failure		404					{object}	SwaggErrDetailsResp
// @Failure		500					{object}	SwaggErrDetailsResp
// @Router			/v1/users/password	[get]
func (h *HTTP) changePassword(c echo.Context) error {
	r := user.PasswordChangeData{}
	if err := c.Bind(&r); err != nil {
		return err
	}
	if err := h.svc.ChangePassword(c.Request().Context(), h.auth.User(c), r); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}
