package router

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/vuduongtp/go-core/internal/api/handler/auth"
	"github.com/vuduongtp/go-core/internal/api/handler/country"
	"github.com/vuduongtp/go-core/internal/api/handler/user"
	"github.com/vuduongtp/go-core/internal/di"
)

// RegisterRoutes registers all API routes
func RegisterRoutes(app *di.Application) {
	// Health check endpoint (no authentication required)
	app.Server.GET("/health", healthCheck)

	// Auth routes (no JWT middleware)
	auth.NewHTTP(app.AuthSvc, app.Server)

	// Protected v1 routes with JWT middleware
	v1Router := app.Server.Group("/v1")
	v1Router.Use(app.JWT.MWFunc())

	// Register module routes on sub-groups
	user.NewHTTP(app.UserSvc, app.Auth, v1Router.Group("/users"))
	country.NewHTTP(app.CountrySvc, app.Auth, v1Router.Group("/countries"))
}

// healthCheck is a simple health check endpoint
func healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"message": "Server is running",
	})
}
