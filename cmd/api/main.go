package main

import (
	"fmt"

	"github.com/vuduongtp/go-core/config"
	"github.com/vuduongtp/go-core/docs"
	_ "github.com/vuduongtp/go-core/docs"
	"github.com/vuduongtp/go-core/internal/api/auth"
	"github.com/vuduongtp/go-core/internal/api/country"
	"github.com/vuduongtp/go-core/internal/api/user"
	userdb "github.com/vuduongtp/go-core/internal/db/user"
	"github.com/vuduongtp/go-core/internal/rbac"
	dbutil "github.com/vuduongtp/go-core/internal/util/db"
	"github.com/vuduongtp/go-core/pkg/server"
	"github.com/vuduongtp/go-core/pkg/server/middleware/jwt"
	"github.com/vuduongtp/go-core/pkg/util/crypter"
	"github.com/vuduongtp/go-core/pkg/util/logger"
	swaggerutil "github.com/vuduongtp/go-core/pkg/util/swagger"
)

//	@title			GoCore Example API
//	@version		1.0
//	@description	This is a sample server Core server.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	DuongVu
//	@contact.url	http://www.swagger.io/support
//	@contact.email	vuduongcalvin@gmail.com

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@schemes					http https
//	@BasePath					/
//	@query.collection.format	multi

// @securityDefinitions.apikey	BearerToken
// @in							header
// @name						Authorization
func main() {
	cfg, err := config.Load()
	checkErr(err)

	logger.SetConfig(logger.Config{
		IngnoredPaths: []string{"runtime/", "gorm.io/gorm/", "github.com/labstack/echo/"},
		DefaultFields: map[string]interface{}{"env": "production"},
	})

	db, err := dbutil.New(cfg.DbType, cfg.DbDsn, cfg.DbLog)
	checkErr(err)
	// connection.Close() is not available for GORM 1.20.0
	// defer db.Close()

	sqlDB, err := db.DB()
	defer sqlDB.Close()

	// Initialize HTTP server
	e := server.New(&server.Config{
		Stage:        cfg.Stage,
		Port:         cfg.Port,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		AllowOrigins: cfg.AllowOrigins,
		Debug:        cfg.Debug,
	})

	// Static page for Swagger API specs
	if cfg.IsEnableAIPDocs {
		docs.SwaggerInfo.Host = fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
		e.GET(fmt.Sprintf("/%s/*", cfg.APIDocsPath), swaggerutil.WrapHandler)
	}

	// Initialize DB interfaces
	userDB := userdb.NewDB()
	countryDB := country.NewDB()

	// Initialize services
	crypterSvc := crypter.New()
	rbacSvc := rbac.New(cfg.Debug)
	jwtSvc := jwt.New(cfg.JwtAlgorithm, cfg.JwtSecret, cfg.JwtDuration)
	authSvc := auth.New(db, userDB, jwtSvc, crypterSvc)
	userSvc := user.New(db, userDB, rbacSvc, crypterSvc)
	countrySvc := country.New(db, countryDB, rbacSvc)

	// Initialize root API
	auth.NewHTTP(authSvc, e)

	// Initialize v1 API
	v1Router := e.Group("/v1")
	v1Router.Use(jwtSvc.MWFunc())

	user.NewHTTP(userSvc, authSvc, v1Router.Group("/users"))
	country.NewHTTP(countrySvc, authSvc, v1Router.Group("/countries"))

	// Start the HTTP server
	server.Start(e, cfg.Stage == "development")
}

func checkErr(err error) {
	if err != nil {
		logger.Panic(err)
		panic(err)
	}
}
