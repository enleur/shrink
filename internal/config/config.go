package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Server ServerConfig
	Redis  RedisConfig
}

type ServerConfig struct {
	Port int `env:"SERVER_PORT" envDefault:"8080"`
}

type RedisConfig struct {
	Address string `env:"REDIS_ADDRESS" envDefault:"localhost:6379"`
	DB      int    `env:"REDIS_DB" envDefault:"0"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return cfg, nil
}
