package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Port                   int    `mapstructure:"PORT"`
	DatabaseURL            string `mapstructure:"DATABASE_URL"`
	NatsURL                string `mapstructure:"NATS_URL"`
	QuickwitURL            string `mapstructure:"QUICKWIT_URL"`
	ClickhouseURL          string `mapstructure:"CLICKHOUSE_URL"`
	ClientOrigin           string `mapstructure:"CLIENT_ORIGIN"`
	JWTSecret              string `mapstructure:"JWT_SECRET"`
	LogLevel               string `mapstructure:"LOG_LEVEL"`
	ServerHost             string `mapstructure:"SERVER_HOST"`
	BlobStorageEndpoint    string `mapstructure:"BLOB_STORAGE_ENDPOINT"`
	BlobStorageAccessKey   string `mapstructure:"BLOB_STORAGE_ACCESS_KEY"`
	BlobStorageSecretKey   string `mapstructure:"BLOB_STORAGE_SECRET_KEY"`
	BlobStorageBucket      string `mapstructure:"BLOB_STORAGE_BUCKET"`
	BlobStorageSecure      bool   `mapstructure:"BLOB_STORAGE_SECURE"`
}

func LoadConfig() Config {
	var config Config

	// Set defaults
	viper.SetDefault("PORT", 8000)
	viper.SetDefault("SERVER_HOST", "0.0.0.0")
	viper.SetDefault("DATABASE_URL", "postgresql://horizon_user:horizon_pass@postgres:5432/horizon")
	viper.SetDefault("NATS_URL", "nats://nats:4222")
	viper.SetDefault("QUICKWIT_URL", "http://quickwit:7280")
	viper.SetDefault("CLICKHOUSE_URL", "http://clickhouse:8123")
	viper.SetDefault("CLIENT_ORIGIN", "http://localhost:3000")
	viper.SetDefault("JWT_SECRET", "default-secret-change-in-production")
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("BLOB_STORAGE_ENDPOINT", "minio:9000")
	viper.SetDefault("BLOB_STORAGE_ACCESS_KEY", "minio_access_key")
	viper.SetDefault("BLOB_STORAGE_SECRET_KEY", "minio_secret_key")
	viper.SetDefault("BLOB_STORAGE_BUCKET", "horizon-logs")
	viper.SetDefault("BLOB_STORAGE_SECURE", false)

	// Read configuration
	err := viper.Unmarshal(&config)
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	return config
}