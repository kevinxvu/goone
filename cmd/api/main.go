package main

import (
	"fmt"

	"github.com/vuduongtp/go-core/internal/api/docs"
	"github.com/vuduongtp/go-core/internal/api/router"
	"github.com/vuduongtp/go-core/internal/di"
	"github.com/vuduongtp/go-core/pkg/logging"
	"github.com/vuduongtp/go-core/pkg/server"
	swaggerutil "github.com/vuduongtp/go-core/pkg/util/swagger"
	"go.uber.org/zap/zapcore"
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
	// Initialize application with Wire DI
	app, err := di.InitializeApplication()
	checkErr(err)

	// Configure logging
	logLevel := zapcore.InfoLevel
	if app.Config.Debug {
		logLevel = zapcore.DebugLevel
	}
	logging.SetConfig(&logging.Config{
		Level:      logLevel,
		FilePath:   "logs/app.log",
		TimeFormat: "2006-01-02 15:04:05",
	})

	// Ensure DB connection is closed on exit
	sqlDB, err := app.DB.DB()
	checkErr(err)
	defer sqlDB.Close()

	// Static page for Swagger API specs
	if app.Config.IsEnableAIPDocs {
		docs.SwaggerInfo.Host = fmt.Sprintf("%s:%d", app.Config.Host, app.Config.Port)
		e := app.Server
		e.GET(fmt.Sprintf("/%s/*", app.Config.APIDocsPath), swaggerutil.WrapHandler)
	}

	// Register all API routes
	router.RegisterRoutes(app)

	// Start the HTTP server
	server.Start(app.Server, app.Config.Stage == "development")
}

func checkErr(err error) {
	if err != nil {
		logging.DefaultLogger().Panic(err.Error())
	}
}
