package brek

import (
	"os"
	"path/filepath"
)

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func ConfigDir() string {
	return envOr("BREK_CONFIG_DIR", "config")
}

func WriteDir() string {
	if value := os.Getenv("BREK_WRITE_DIR"); value != "" {
		return value
	}

	return ConfigDir()
}

func LoadersFilePath() string {
	return envOr("BREK_LOADERS_FILE_PATH", "brek.loaders.js")
}

func ConfigJSONPath() string {
	return filepath.Join(WriteDir(), "config.json")
}

func ConfigLockPath() string {
	return filepath.Join(WriteDir(), "config.lock")
}
