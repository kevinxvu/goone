//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/labstack/echo/v4"
	"github.com/vuduongtp/go-core/config"
	authSvc "github.com/vuduongtp/go-core/internal/api/service/auth"
	countrySvc "github.com/vuduongtp/go-core/internal/api/service/country"
	userSvc "github.com/vuduongtp/go-core/internal/api/service/user"
	"github.com/vuduongtp/go-core/internal/model"
	"github.com/vuduongtp/go-core/internal/repository"
	"github.com/vuduongtp/go-core/pkg/database"
	"github.com/vuduongtp/go-core/pkg/server"
	"github.com/vuduongtp/go-core/pkg/server/middleware/jwt"
	"gorm.io/gorm"
)

// ProvideConfig loads configuration
func ProvideConfig() (*config.Configuration, error) {
	return config.Load()
}

// ProvideDB initializes database connection
func ProvideDB(cfg *config.Configuration) (*gorm.DB, error) {
	return database.New(cfg.DbType, cfg.DbDsn, cfg.DbLog)
}

// ProvideUserDB creates user database repository
func ProvideUserDB() *repository.UserRepository {
	return repository.NewUserRepository()
}

// ProvideCountryDB creates country database repository
func ProvideCountryDB() *repository.CountryRepository {
	return repository.NewCountryRepository()
}

// ProvideJWT creates JWT service
func ProvideJWT(cfg *config.Configuration) *jwt.Service {
	return jwt.New(cfg.JwtAlgorithm, cfg.JwtSecret, cfg.JwtDuration)
}

// ProvideAuth creates Auth interface from JWT service
func ProvideAuth(jwtSvc *jwt.Service) model.Auth {
	return jwtSvc
}

// ProvideAuthJWT creates auth.JWT interface from JWT service
func ProvideAuthJWT(jwtSvc *jwt.Service) authSvc.JWT {
	return jwtSvc
}

// ProvideAuthService creates auth service
func ProvideAuthService(db *gorm.DB, userDB *repository.UserRepository, jwtSvc authSvc.JWT) authSvc.Service {
	return authSvc.New(db, userDB, jwtSvc)
}

// ProvideUserService creates user service
func ProvideUserService(db *gorm.DB, userDB *repository.UserRepository) userSvc.Service {
	return userSvc.New(db, userDB)
}

// ProvideCountryService creates country service
func ProvideCountryService(db *gorm.DB, countryDB *repository.CountryRepository) countrySvc.Service {
	return countrySvc.New(db, countryDB)
}

// ProvideServer creates Echo server
func ProvideServer(cfg *config.Configuration) *echo.Echo {
	return server.New(&server.Config{
		Stage:        cfg.Stage,
		Port:         cfg.Port,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		AllowOrigins: cfg.AllowOrigins,
		Debug:        cfg.Debug,
	})
}

// Application holds all initialized services
type Application struct {
	Config     *config.Configuration
	DB         *gorm.DB
	Server     *echo.Echo
	JWT        *jwt.Service
	Auth       model.Auth
	AuthSvc    authSvc.Service
	UserSvc    userSvc.Service
	CountrySvc countrySvc.Service
}

// InitializeApplication uses wire to build all dependencies
func InitializeApplication() (*Application, error) {
	wire.Build(
		ProvideConfig,
		ProvideDB,
		ProvideUserDB,
		ProvideCountryDB,
		ProvideJWT,
		ProvideAuth,
		ProvideAuthJWT,
		ProvideAuthService,
		ProvideUserService,
		ProvideCountryService,
		ProvideServer,
		wire.Struct(new(Application), "*"),
	)
	return nil, nil
}
