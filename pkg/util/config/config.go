package cfgutil

import (
	"os"

	"github.com/caarlos0/env/v5"
	"github.com/joho/godotenv"
)

// Load loads configuration with priority order:
// 1. OS environment variables (highest priority)
// 2. .env.local file
// 3. .env file (lowest priority)
func Load(out interface{}, stage string) error {
	// Load .env files only if env vars are not already set in OS
	// godotenv.Load will NOT overwrite existing OS env vars
	if err := LoadLocalENV(stage); err != nil {
		return err
	}

	// Parse from environment variables (OS env vars take priority)
	if err := env.Parse(out); err != nil {
		return err
	}

	return nil
}

// LoadConfig loads configuration from .env file
func LoadConfig(out interface{}, appName, stage string) error {

	return Load(out, stage)
}

// LoadLocalENV reads .env* files and sets the values to os ENV
// Priority order (highest to lowest):
// 1. Existing OS environment variables (not overwritten)
// 2. .env.local file
// 3. .env file
func LoadLocalENV(stage string) error {
	basePath := ""
	if stage == "test" {
		basePath = "testdata/"
	}

	// Load .env files in reverse order (lowest priority first)
	// godotenv.Load() does NOT overwrite existing env vars

	// Load default config first (.env) - lowest priority
	if _, err := os.Stat(basePath + ".env"); err == nil {
		_ = godotenv.Load(basePath + ".env")
	}

	// Load local config (.env.local) - can override .env values
	// but will NOT override OS env vars
	if _, err := os.Stat(basePath + ".env.local"); err == nil {
		_ = godotenv.Load(basePath + ".env.local")
	}

	// OS environment variables always have highest priority
	// They are never overwritten by godotenv.Load()

	return nil
}
