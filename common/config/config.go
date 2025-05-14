package config

import (
	"github.com/spf13/viper"
)

// Config holds application settings
type Config struct {
	KafkaBrokers []string
	SchemaURL    string
}

// Load reads from env or defaults
func Load() (*Config, error) {
	viper.SetEnvPrefix("APP")
	viper.AutomaticEnv()

	viper.SetDefault("KAFKA_BROKERS", []string{"localhost:9092"})
	viper.SetDefault("SCHEMA_URL", "http://localhost:8081")

	return &Config{
		KafkaBrokers: viper.GetStringSlice("KAFKA_BROKERS"),
		SchemaURL:    viper.GetString("SCHEMA_URL"),
	}, nil
}
