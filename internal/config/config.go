package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	HTTPAddr string `env:"HTTP_ADDR" env-default:"127.0.0.1:8080"`

	DatabaseURL string `env:"DATABASE_URL" env-default:"postgres://postgres:postgres@localhost:5432/ledger_db?sslmode=disable"`
}

func Load() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	return &cfg, nil
}
