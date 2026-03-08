package country

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/kevinxvu/goone/internal/api/service/country"
	"github.com/kevinxvu/goone/internal/model"
	"github.com/kevinxvu/goone/pkg/server/apperr"
	"github.com/kevinxvu/goone/pkg/util/request"
	"github.com/labstack/echo/v4"
)

// HTTP represents country http service
type HTTP struct {
	svc  country.Service
	auth model.Auth
}

// NewHTTP creates new country http service
func NewHTTP(svc country.Service, auth model.Auth, eg *echo.Group) {
	h := HTTP{svc, auth}

	eg.POST("", h.create)
	eg.GET("/:id", h.view)
	eg.GET("", h.list)
	eg.PATCH("/:id", h.update)
	eg.DELETE("/:id", h.delete)
}

// @Security		BearerToken
// @Summary		Creates new country
// @Description	Creates new country
// @Accept			json
// @Produce		json
// @Tags			countries
// @ID				countriesCreate
// @Param			request			body		country.CreationData	true	"CreationData"
// @Success		200				{object}	model.Country
// @Failure		401				{object}	SwaggErrDetailsResp
// @Failure		403				{object}	SwaggErrDetailsResp
// @Failure		500				{object}	SwaggErrDetailsResp
// @Router			/v1/countries	[post]
func (h *HTTP) create(c echo.Context) error {
	r := country.CreationData{}
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

	return c.JSON(http.StatusOK, country.ListResp{Data: resp, TotalCount: count})
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
	r := country.UpdateData{}
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
