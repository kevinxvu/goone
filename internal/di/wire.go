//go:build wireinject
// +build wireinject

package di

import (
	"github.com/google/wire"
	"github.com/kevinxvu/goone/config"
	authSvc "github.com/kevinxvu/goone/internal/api/service/auth"
	countrySvc "github.com/kevinxvu/goone/internal/api/service/country"
	userSvc "github.com/kevinxvu/goone/internal/api/service/user"
	"github.com/kevinxvu/goone/internal/model"
	"github.com/kevinxvu/goone/internal/repository"
	"github.com/kevinxvu/goone/pkg/database"
	openaiPkg "github.com/kevinxvu/goone/pkg/openai"
	"github.com/kevinxvu/goone/pkg/server"
	"github.com/kevinxvu/goone/pkg/server/middleware/jwt"
	"github.com/labstack/echo/v4"
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

// ProvideOpenAIService creates OpenAI service
func ProvideOpenAIService(cfg *config.Configuration) *openaiPkg.Service {
	return openaiPkg.New(openaiPkg.Config{
		APIKey:     cfg.OpenAIAPIKey,
		BaseURL:    cfg.OpenAIBaseURL,
		Timeout:    cfg.OpenAITimeout,
		MaxRetries: cfg.OpenAIMaxRetries,
		TextModel:  cfg.OpenAITextModel,
		AudioModel: cfg.OpenAIAudioModel,
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
	OpenAI     *openaiPkg.Service
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
		ProvideOpenAIService,
		wire.Struct(new(Application), "*"),
	)
	return nil, nil
}
