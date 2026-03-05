package config

import (
	"fmt"
	"os"

	cfgutil "github.com/vuduongtp/go-core/pkg/util/config"
)

// Configuration holds data necessery for configuring application
type Configuration struct {
	Stage           string   `env:"STAGE"`
	Host            string   `env:"HOST"`
	Port            int      `env:"PORT"`
	ReadTimeout     int      `env:"READ_TIMEOUT"`
	WriteTimeout    int      `env:"WRITE_TIMEOUT"`
	AllowOrigins    []string `env:"ALLOW_ORIGINS"`
	Debug           bool     `env:"DEBUG"`
	DbLog           bool     `env:"DB_LOG"`
	DbType          string   `env:"DB_TYPE"`
	DbDsn           string   `env:"DB_DSN"`
	JwtSecret       string   `env:"JWT_SECRET"`
	JwtDuration     int      `env:"JWT_DURATION"`
	JwtAlgorithm    string   `env:"JWT_ALGORITHM"`
	IsEnableAIPDocs bool     `env:"IS_ENABLE_API_DOCS"`
	APIDocsPath     string   `env:"API_DOCS_PATH"`
}

// Load returns Configuration struct
func Load() (*Configuration, error) {
	appName := os.Getenv("AWS_LAMBDA_FUNCTION_NAME")
	if configname := os.Getenv("CONFIG_NAME"); configname != "" {
		appName = configname
	}
	stage := os.Getenv("STAGE")
	if configstage := os.Getenv("CONFIG_STAGE"); configstage != "" {
		stage = configstage
	}

	cfg := new(Configuration)
	if err := cfgutil.LoadConfig(cfg, appName, stage); err != nil {
		return nil, fmt.Errorf("Error parsing environment config: %s", err)
	}
	return cfg, nil
}
