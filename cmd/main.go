package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	server "github.com/BazaarTrade/OrderMatchingService/internal/api/gRPC"
	"github.com/BazaarTrade/OrderMatchingService/internal/repository/postgresPgx"
	"github.com/BazaarTrade/OrderMatchingService/internal/service"
	"github.com/joho/godotenv"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	//load .env file if it exists
	//i use this to load the env variables from docker compose
	if _, err := os.Stat("../.env"); err == nil {
		if err := godotenv.Load("../.env"); err != nil {
			logger.Error("failed to load .env file", "error", err)
			return
		}
	}

	logger.Info("starting aplication...")

	DB_CONNECTION := os.Getenv("DB_CONNECTION")
	if DB_CONNECTION == "" {
		logger.Error("DB_CONNECTION environment variable is not set")
		return
	}

	repository, err := postgresPgx.New(DB_CONNECTION, logger)
	if err != nil {
		return
	}

	service := service.New(repository, logger)
	server := server.New(service, logger)

	if err := initOrderBooks(server, service); err != nil {
		return
	}

	GRPC_PORT := os.Getenv("GRPC_PORT")
	if GRPC_PORT == "" {
		logger.Error("ADDR environment variable is not set")
		return
	}

	go func() {
		if err := server.Run(GRPC_PORT); err != nil {
			os.Exit(1)
		}
	}()

	//Graceful shutdown

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop
	logger.Info("shutting down...")

	server.Stop()
	logger.Info("stopped gRPC server")

	repository.Close()
	logger.Info("closed database connection")

	logger.Info("gracefully stopped")
}

func initOrderBooks(server *server.Server, service *service.Service) error {
	pairs, err := service.GetPairs()
	if err != nil {
		return err
	}

	for _, pair := range pairs {
		service.CreateOrderBook(pair)
		server.NewStreamHub(pair)
	}
	return nil
}
