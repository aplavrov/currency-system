package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/aplavrov/currency-system/gw-exchanger/internal/config"
	server "github.com/aplavrov/currency-system/gw-exchanger/internal/delivery/grpc"
	"github.com/aplavrov/currency-system/gw-exchanger/internal/service"
	"github.com/aplavrov/currency-system/gw-exchanger/internal/storage/postgresql"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := godotenv.Load("config.env"); err != nil {
		log.Fatalf("failed to load env: %v", err)
	}

	cfg := config.NewConfig()

	level := config.ParseLogLevel(cfg.Log.Level)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})).With("service", "gw-exchanger")

	logger.Info("starting application")

	dbPool, err := postgresql.NewDB(ctx, cfg)
	if err != nil {
		logger.Error("db connection failed", "error", err)
		os.Exit(1)
	}
	defer dbPool.GetPool().Close()

	logger.Info("db connected")

	exchangeStorage := postgresql.NewExchangeStorage(dbPool, logger.With("layer", "repository", "db", "postgres"))
	exchangeService := service.New(exchangeStorage, logger.With("layer", "service"))

	gRPCServer := grpc.NewServer()
	server.Register(gRPCServer, exchangeService, logger.With("layer", "transport", "protocol", "grpc"))
	l, err := net.Listen("tcp", fmt.Sprintf(":%v", cfg.Server.GRPCPort))
	if err != nil {
		logger.Error("listen failed", "error", err)
		os.Exit(1)
	}

	logger.Info("grpc server started", "port", l.Addr().String())

	go func() {
		if err = gRPCServer.Serve(l); err != nil {
			logger.Error("failed to serve", "error", err)
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

	logger.Info("app shutting down")
	gRPCServer.GracefulStop()
}
