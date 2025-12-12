package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DatabaseURL          string        `mapstructure:"database_url"`
	ServerAddress        string        `mapstructure:"server_address"`
	TokenSymmetricKey    string        `mapstructure:"token_symmetric_key"`
	AccessTokenDuration  time.Duration `mapstructure:"access_token_duration"`
	RefreshTokenDuration time.Duration `mapstructure:"refresh_token_duration"`
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
	v.SetDefault("access_token_duration", "15m")
	v.SetDefault("refresh_token_duration", "720h") // 30 days

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
	if cfg.TokenSymmetricKey == "" {
		cfg.TokenSymmetricKey = os.Getenv("TOKEN_SYMMETRIC_KEY")
		if cfg.TokenSymmetricKey == "" {
			return nil, fmt.Errorf("token_symmetric_key / TOKEN_SYMMETRIC_KEY is required")
		}
	}

	// Resolve durations from viper to allow string inputs such as "15m"
	cfg.AccessTokenDuration = v.GetDuration("access_token_duration")
	if cfg.AccessTokenDuration == 0 {
		cfg.AccessTokenDuration = 15 * time.Minute
	}
	cfg.RefreshTokenDuration = v.GetDuration("refresh_token_duration")
	if cfg.RefreshTokenDuration == 0 {
		cfg.RefreshTokenDuration = 720 * time.Hour
	}

	return &cfg, nil
}
