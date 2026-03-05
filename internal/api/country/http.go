package country

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"github.com/vuduongtp/go-core/internal/model"
	"github.com/vuduongtp/go-core/pkg/database"
	"github.com/vuduongtp/go-core/pkg/server/apperr"
	"github.com/vuduongtp/go-core/pkg/util/request"

	"github.com/labstack/echo/v4"
)

// HTTP represents country http service
type HTTP struct {
	svc  Service
	auth model.Auth
}

// Service represents country application interface
type Service interface {
	Create(context.Context, *model.AuthUser, CreationData) (*model.Country, error)
	View(context.Context, *model.AuthUser, int) (*model.Country, error)
	List(context.Context, *model.AuthUser, *database.ListQueryCondition, *int64) ([]*model.Country, error)
	Update(context.Context, *model.AuthUser, int, UpdateData) (*model.Country, error)
	Delete(context.Context, *model.AuthUser, int) error
}

// NewHTTP creates new country http service
func NewHTTP(svc Service, auth model.Auth, eg *echo.Group) {
	h := HTTP{svc, auth}

	eg.POST("", h.create)
	eg.GET("/:id", h.view)
	eg.GET("", h.list)
	eg.PATCH("/:id", h.update)
	eg.DELETE("/:id", h.delete)
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

// @Security		BearerToken
// @Summary		Creates new country
// @Description	Creates new country
// @Accept			json
// @Produce		json
// @Tags			countries
// @ID				countriesCreate
// @Param			request			body		country.CreationData	true	"CreationData"
// @Success		200				{object}	model.AuthToken
// @Failure		401				{object}	SwaggErrDetailsResp
// @Failure		403				{object}	SwaggErrDetailsResp
// @Failure		500				{object}	SwaggErrDetailsResp
// @Router			/v1/countries	[post]
func (h *HTTP) create(c echo.Context) error {
	r := CreationData{}
	if err := c.Bind(&r); err != nil {
		return err
	}
	r.Name = strings.TrimSpace(r.Name)
	r.Code = strings.ToUpper(strings.TrimSpace(r.Code))
	r.PhoneCode = strings.ReplaceAll(r.PhoneCode, " ", "")

	if regexp.MustCompile(`^\+\d+$`).Match([]byte(r.PhoneCode)) == false {
		return apperr.NewHTTPValidationError("PhoneCode is invalid")
	}

	resp, err := h.svc.Create(c.Request().Context(), h.auth.User(c), r)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, resp)
}

// @Security		BearerToken
// @Summary		Returns a single country
// @Description	Returns a single country
// @Accept			json
// @Produce		json
// @Tags			countries
// @ID				countriesView
// @Param			id					path		int	true	"Country ID"
// @Success		200					{object}	model.Country
// @Failure		401					{object}	SwaggErrDetailsResp
// @Failure		403					{object}	SwaggErrDetailsResp
// @Failure		500					{object}	SwaggErrDetailsResp
// @Router			/v1/countries/{id}	[get]
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
// @Summary		Get list country
// @Description	Get list country
// @Accept			json
// @Produce		json
// @Tags			countries
// @ID				countriesList
// @Param			q				query		ListRequest	false	"QueryListRequest"
// @Success		200				{object}	country.ListResp
// @Failure		400				{object}	SwaggErrDetailsResp
// @Failure		401				{object}	SwaggErrDetailsResp
// @Failure		403				{object}	SwaggErrDetailsResp
// @Failure		500				{object}	SwaggErrDetailsResp
// @Router			/v1/countries	[get]
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

	return c.JSON(http.StatusOK, ListResp{resp, count})
}

// @Security		BearerToken
// @Summary		Updates country information
// @Description	Updates country information
// @Accept			json
// @Produce		json
// @Tags			countries
// @ID				countriesUpdate
// @Param			id					path		int					true	"Country ID"
// @Param			request				body		country.UpdateData	true	"UpdateData"
// @Success		200					{object}	model.Country
// @Failure		400					{object}	SwaggErrDetailsResp
// @Failure		401					{object}	SwaggErrDetailsResp
// @Failure		403					{object}	SwaggErrDetailsResp
// @Failure		404					{object}	SwaggErrDetailsResp
// @Failure		500					{object}	SwaggErrDetailsResp
// @Router			/v1/countries/{id}	[patch]
func (h *HTTP) update(c echo.Context) error {
	id, err := request.ReqID(c)
	if err != nil {
		return err
	}
	r := UpdateData{}
	if err := c.Bind(&r); err != nil {
		return err
	}
	r.Name = request.TrimSpacePointer(r.Name)
	r.Code = request.TrimSpacePointer(r.Code)
	r.PhoneCode = request.RemoveSpacePointer(r.PhoneCode)

	usr, err := h.svc.Update(c.Request().Context(), h.auth.User(c), id, r)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, usr)
}

// @Security		BearerToken
// @Summary		Deletes an country
// @Description	Deletes an country
// @Accept			json
// @Produce		json
// @Tags			countries
// @ID				countriesDelete
// @Param			id					path		int	true	"Country ID"
// @Success		200					{object}	SwaggOKResp
// @Failure		400					{object}	SwaggErrDetailsResp
// @Failure		401					{object}	SwaggErrDetailsResp
// @Failure		403					{object}	SwaggErrDetailsResp
// @Failure		404					{object}	SwaggErrDetailsResp
// @Failure		500					{object}	SwaggErrDetailsResp
// @Router			/v1/countries/{id}	[delete]
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
