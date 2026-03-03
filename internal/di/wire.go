//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/labstack/echo/v4"
	"github.com/vuduongtp/go-core/config"
	"github.com/vuduongtp/go-core/internal/api/auth"
	"github.com/vuduongtp/go-core/internal/api/country"
	"github.com/vuduongtp/go-core/internal/api/user"
	userdb "github.com/vuduongtp/go-core/internal/db/user"
	"github.com/vuduongtp/go-core/internal/model"
	dbutil "github.com/vuduongtp/go-core/internal/util/db"
	"github.com/vuduongtp/go-core/pkg/server"
	"github.com/vuduongtp/go-core/pkg/server/middleware/jwt"
	"github.com/vuduongtp/go-core/pkg/util/crypter"
	pkgdb "github.com/vuduongtp/go-core/pkg/util/db"
	"gorm.io/gorm"
)

// ProvideConfig loads configuration
func ProvideConfig() (*config.Configuration, error) {
	return config.Load()
}

// ProvideDB initializes database connection
func ProvideDB(cfg *config.Configuration) (*gorm.DB, error) {
	return dbutil.New(cfg.DbType, cfg.DbDsn, cfg.DbLog)
}

// ProvideUserDB creates user database repository
func ProvideUserDB() *userdb.DB {
	return userdb.NewDB()
}

// ProvideCountryDB creates country database repository
func ProvideCountryDB() *pkgdb.DB {
	return country.NewDB()
}

// ProvideCrypter creates crypter service
func ProvideCrypter() *crypter.Service {
	return crypter.New()
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
func ProvideAuthJWT(jwtSvc *jwt.Service) auth.JWT {
	return jwtSvc
}

// ProvideAuthService creates auth service
func ProvideAuthService(db *gorm.DB, userDB *userdb.DB, jwtSvc auth.JWT, crypterSvc *crypter.Service) auth.Service {
	return auth.New(db, userDB, jwtSvc, crypterSvc)
}

// ProvideUserService creates user service
func ProvideUserService(db *gorm.DB, userDB *userdb.DB, crypterSvc *crypter.Service) user.Service {
	return user.New(db, userDB, crypterSvc)
}

// ProvideCountryService creates country service
func ProvideCountryService(db *gorm.DB, countryDB *pkgdb.DB) country.Service {
	return country.New(db, countryDB)
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
	AuthSvc    auth.Service
	UserSvc    user.Service
	CountrySvc country.Service
}

// InitializeApplication uses wire to build all dependencies
func InitializeApplication() (*Application, error) {
	wire.Build(
		ProvideConfig,
		ProvideDB,
		ProvideUserDB,
		ProvideCountryDB,
		ProvideCrypter,
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
