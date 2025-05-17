package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	KafkaBrokers []string
	SchemaURL    string
}

func Load() (*Config, error) {
	viper.SetEnvPrefix("APP")
	viper.AutomaticEnv()

	viper.SetDefault("KAFKA_BROKERS", []string{"localhost:9092"})
	viper.SetDefault("SCHEMA_URL", "http://localhost:8081")

	brokers := viper.GetStringSlice("KAFKA_BROKERS")
	log.Printf("⚙️  Using Kafka brokers: %v", brokers)

	return &Config{
		KafkaBrokers: brokers,
		SchemaURL:    viper.GetString("SCHEMA_URL"),
	}, nil
}
