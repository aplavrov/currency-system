package config

import (
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	Storage StorageConfig
	Server  gRPCConfig
	Log     LogConfig
}

type StorageConfig struct {
	PostgresHost     string
	PostgresPort     string
	PostgresDatabase string
	PostgresUsername string
	PostgresPassword string
}

type gRPCConfig struct {
	GRPCPort string
}

type LogConfig struct {
	Level string
}

func NewConfig() Config {
	storage := StorageConfig{
		PostgresHost:     os.Getenv("DB_HOST"),
		PostgresPort:     os.Getenv("DB_PORT"),
		PostgresDatabase: os.Getenv("DB_NAME"),
		PostgresUsername: os.Getenv("DB_USER"),
		PostgresPassword: os.Getenv("DB_PASSWORD"),
	}
	server := gRPCConfig{
		GRPCPort: os.Getenv("GRPC_PORT"),
	}

	logCfg := LogConfig{
		Level: os.Getenv("LOG_LEVEL"),
	}

	config := Config{
		Storage: storage,
		Server:  server,
		Log:     logCfg,
	}
	return config
}

func ParseLogLevel(level string) slog.Level {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
