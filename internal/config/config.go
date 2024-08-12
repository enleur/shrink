package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Server ServerConfig
	Redis  RedisConfig
	Otel   OtelConfig
}

type ServerConfig struct {
	Port int    `env:"SERVER_PORT" envDefault:"8080"`
	Mode string `env:"GIN_MODE" envDefault:"debug"`
}

type RedisConfig struct {
	Address string `env:"REDIS_ADDRESS" envDefault:"localhost:6379"`
	DB      int    `env:"REDIS_DB" envDefault:"0"`
}

type OtelConfig struct {
	Endpoint string `env:"OTEL_EXPORTER_OTLP_ENDPOINT" envDefault:"localhost:4317"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return cfg, nil
}
