package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	DatabaseURL   string `mapstructure:"database_url"`
	ServerAddress string `mapstructure:"server_address"`
}

// Load reads configuration from config.yaml (if present) and environment variables.
// Precedence (highest to lowest):
//  1. Environment variables DATABASE_URL / SERVER_ADDRESS
//  2. Environment variables CHARITY_DATABASE_URL / CHARITY_SERVER_ADDRESS
//  3. config.yaml values (database_url, server_address)
//  4. Built-in default for server_address (":8080")
func Load() (*Config, error) {
	v := viper.New()

	// Look for config.yaml in the project root (working directory)
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	// Environment variables with prefix CHARITY_ map to fields, e.g. CHARITY_DATABASE_URL
	v.SetEnvPrefix("CHARITY")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Defaults
	v.SetDefault("server_address", ":8080")

	// Load config file if present; it's optional
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshalling config: %w", err)
	}

	// Allow plain env vars without prefix to override everything
	if envDSN := os.Getenv("DATABASE_URL"); envDSN != "" {
		cfg.DatabaseURL = envDSN
	}
	if envAddr := os.Getenv("SERVER_ADDRESS"); envAddr != "" {
		cfg.ServerAddress = envAddr
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("database_url / DATABASE_URL is required")
	}
	if cfg.ServerAddress == "" {
		cfg.ServerAddress = ":8080"
	}

	return &cfg, nil
}
