package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/aplavrov/currency-system/gw-notification/internal/config"
	consumer "github.com/aplavrov/currency-system/gw-notification/internal/delivery/kafka"
	"github.com/aplavrov/currency-system/gw-notification/internal/service"
	mongodb "github.com/aplavrov/currency-system/gw-notification/internal/storage/mongo"
	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/joho/godotenv"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := godotenv.Load("config.env"); err != nil {
		log.Fatalf("failed to load env: %v", err)
	}

	cfg := config.NewConfig()

	level := config.ParseLogLevel(cfg.Log.Level)

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})).With("service", "gw-notification")

	logger.Info("starting application")

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.Storage.MongoConn))
	if err != nil {
		logger.Error("error connecting to MongoDB", "error", err)
		os.Exit(1)
	}

	coll := mongodb.NewOperationStorage(client, cfg.Storage.MongoDBName, cfg.Storage.MongoCollectionName, logger.With("layer", "repository", "db", "mongodb"))
	defer coll.Collection.Database().Client().Disconnect(ctx)

	logger.Info("connected to MongoDB")

	notificationService := service.NewNotificationService(coll, logger.With("layer", "service"))

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{cfg.Kafka.Address},
		Topic:     cfg.Kafka.Topic,
		GroupID:   cfg.Kafka.GroupID,
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
		Partition: 0,
	})
	defer reader.Close()

	kafkaConsumer := consumer.NewConsumer(notificationService, logger, reader)
	kafkaConsumer.Start(ctx)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	select {
	case <-stop:
		logger.Info("shutdown signal received")
	case <-ctx.Done():
		logger.Info("context cancelled")
	}
}
