package config

import (
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	Storage StorageConfig
	Log     LogConfig
	Kafka   KafkaConfig
}

type StorageConfig struct {
	MongoConn           string
	MongoDBName         string
	MongoCollectionName string
}

type LogConfig struct {
	Level string
}

type KafkaConfig struct {
	Address string
	Topic   string
	GroupID string
}

func NewConfig() Config {
	storage := StorageConfig{
		MongoConn:           os.Getenv("DB_PATH"),
		MongoDBName:         os.Getenv("DB_NAME"),
		MongoCollectionName: os.Getenv("DB_COLLECTION_NAME"),
	}

	logCfg := LogConfig{
		Level: os.Getenv("LOG_LEVEL"),
	}

	kafkaCfg := KafkaConfig{
		Address: os.Getenv("KAFKA_ADDRESS"),
		Topic:   os.Getenv("KAFKA_TOPIC"),
		GroupID: os.Getenv("KAFKA_GROUP_ID"),
	}

	config := Config{
		Storage: storage,
		Log:     logCfg,
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
