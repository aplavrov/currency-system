package config

import (
	"log/slog"
	"os"
	"strings"
	"time"
)

type Config struct {
	Storage StorageConfig
	Server  GinConfig
	Client  gRPCConfig
	Log     LogConfig
	Auth    AuthConfig
	Kafka   KafkaConfig
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

type GinConfig struct {
	GINPort string
}

type AuthConfig struct {
	JWTSecret string
	TokenTTL  time.Duration
}

type KafkaConfig struct {
	Address string
	Topic   string
	GroupID string
}

func NewConfig() Config {
	storage := StorageConfig{
		PostgresHost:     os.Getenv("DB_HOST"),
		PostgresPort:     os.Getenv("DB_PORT"),
		PostgresDatabase: os.Getenv("DB_NAME"),
		PostgresUsername: os.Getenv("DB_USER"),
		PostgresPassword: os.Getenv("DB_PASSWORD"),
	}
	server := GinConfig{
		GINPort: os.Getenv("GIN_PORT"),
	}
	client := gRPCConfig{
		GRPCPort: os.Getenv("GRPC_PORT"),
	}

	logCfg := LogConfig{
		Level: os.Getenv("LOG_LEVEL"),
	}

	ttlStr := os.Getenv("JWT_TTL")
	ttl, _ := time.ParseDuration(ttlStr)
	authCfg := AuthConfig{
		JWTSecret: os.Getenv("JWT_SECRET"),
		TokenTTL:  ttl,
	}

	kafkaCfg := KafkaConfig{
		Address: os.Getenv("KAFKA_ADDRESS"),
		Topic:   os.Getenv("KAFKA_TOPIC"),
		GroupID: os.Getenv("KAFKA_GROUP_ID"),
	}

	config := Config{
		Storage: storage,
		Server:  server,
		Client:  client,
		Log:     logCfg,
		Auth:    authCfg,
		Kafka:   kafkaCfg,
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
