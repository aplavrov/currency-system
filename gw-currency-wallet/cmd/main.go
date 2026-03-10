package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aplavrov/currency-system/gw-currency-wallet/internal/config"
	server "github.com/aplavrov/currency-system/gw-currency-wallet/internal/delivery/http"
	"github.com/aplavrov/currency-system/gw-currency-wallet/internal/grpcclient"
	"github.com/aplavrov/currency-system/gw-currency-wallet/internal/service"
	"github.com/aplavrov/currency-system/gw-currency-wallet/internal/storage/cache"
	"github.com/aplavrov/currency-system/gw-currency-wallet/internal/storage/kafka"
	postgresql "github.com/aplavrov/currency-system/gw-currency-wallet/internal/storage/postgres"
	"github.com/aplavrov/currency-system/gw-currency-wallet/internal/storage/tx_manager"
	exchange_grpc "github.com/aplavrov/currency-system/proto-exchange/exchange"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// @title Currency Wallet API
// @version 1.0
// @description API for wallet and currency exchange
// @host localhost:9000
// @BasePath /api/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := godotenv.Load("config.env"); err != nil {
		log.Fatalf("failed to load env: %v", err)
	}

	cfg := config.NewConfig()

	level := config.ParseLogLevel(cfg.Log.Level)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})).With("service", "gw-currency-wallet")

	logger.Info("starting application")
	dbPool, err := postgresql.NewDB(ctx, cfg)
	if err != nil {
		logger.Error("db connection failed", "error", err)
		os.Exit(1)
	}
	defer dbPool.GetPool().Close()
	logger.Info("db connected")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", cfg.Storage.PostgresHost, cfg.Storage.PostgresPort, cfg.Storage.PostgresUsername, cfg.Storage.PostgresPassword, cfg.Storage.PostgresDatabase)
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()

	if err := goose.Up(sqlDB, "internal/storage/migrations"); err != nil {
		log.Fatal(err)
	}

	txManager := tx_manager.NewTxManager(dbPool)
	walletStorage := postgresql.NewWalletStorage(txManager, logger.With("layer", "repository", "db", "postgres"))
	userStorage := postgresql.NewUserStorage(txManager, logger.With("layer", "repository", "db", "postgres"))

	conn, err := grpc.NewClient(cfg.Client.GRPCPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("gRPC connection failed", "error", err)
		os.Exit(1)
	}
	defer conn.Close()
	logger.Info("gRPC client connected")

	grpcClient := exchange_grpc.NewExchangeServiceClient(conn)

	rateCache := cache.NewRateCache(1 * time.Minute)
	rateProvider := grpcclient.NewExchangeClient(grpcClient, rateCache, logger)

	kafkaSender := kafka.NewProducer(cfg.Kafka.Address, cfg.Kafka.Topic)
	defer kafkaSender.Close()

	authService := service.NewAuthService(userStorage, walletStorage, txManager, cfg.Auth.JWTSecret, cfg.Auth.TokenTTL, logger.With("layer", "service"))
	walletService := service.NewWalletService(walletStorage, kafkaSender, txManager, logger.With("layer", "service"))
	exchangeService := service.NewExchangeService(walletStorage, txManager, rateProvider, logger.With("layer", "service"))
	handler := server.NewWalletServer(authService, walletService, exchangeService, cfg.Auth.JWTSecret, logger.With("layer", "transport", "protocol", "http"))

	go func() {
		if err := handler.Start(cfg.Server.GINPort); err != nil {
			logger.Error("failed to start GIN server", "error", err)
			cancel()
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	select {
	case <-stop:
		logger.Info("shutdown signal received")
	case <-ctx.Done():
		logger.Info("context cancelled")
	}
}
