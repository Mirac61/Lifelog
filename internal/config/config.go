package config

import (
	"fmt"
	"os"
)

type Config struct {
	DatabasePath  string
	GitHubToken   string
	ListenAddress string
}

func Load() (Config, error) {
	cfg := Config{
		DatabasePath:  envOrDefault("DATABASE_PATH", "data.db"),
		GitHubToken:   os.Getenv("GITHUB_TOKEN"),
		ListenAddress: envOrDefault("LISTEN_ADDRESS", "127.0.0.1:8080"),
	}

	if cfg.GitHubToken == "" {
		return Config{}, fmt.Errorf("GITHUB_TOKEN is not set")
	}
	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
