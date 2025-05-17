package config

import (
	"strings"

	"github.com/spf13/viper"
)

// Config holds application settings
type Config struct {
	Env          string   // "dev" or "prod"
	KafkaBrokers []string // e.g. ["kafka:9092"]
	LogLevel     string   // "debug", "info", "error"
}

// Load reads from env or config files, with profiles.
func Load() (*Config, error) {
	// Look for APP_ENV (dev/prod) and APP_LOG_LEVEL
	viper.SetEnvPrefix("APP")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Default environment
	env := viper.GetString("ENV")
	if env == "" {
		env = "dev"
	}
	// Try to read config/<env>.yaml if present
	viper.SetConfigName(env)
	viper.AddConfigPath("./config")
	_ = viper.ReadInConfig()

	// Defaults
	viper.SetDefault("KAFKA_BROKERS", []string{"localhost:9092"})
	viper.SetDefault("LOG_LEVEL", "info")

	return &Config{
		Env:          env,
		KafkaBrokers: viper.GetStringSlice("KAFKA_BROKERS"),
		LogLevel:     viper.GetString("LOG_LEVEL"),
	}, nil
}
