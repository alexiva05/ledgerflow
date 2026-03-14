package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	ServerPort  int    `mapstructure:"server_port"`
	Level       string `mapstructure:"level"`
	DatabaseURL string `mapstructure:"database_url"`
}

func Load() (*Config, error) {
	var cfg Config

	viper.AutomaticEnv()
	viper.SetDefault("server_port", 8080)
	viper.SetDefault("level", "info")

	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("config mapping error: %w", err)
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("config: DATABASE_URL is required")
	}

	return &cfg, nil
}
