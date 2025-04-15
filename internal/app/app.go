package app

import (
	"log/slog"
	"os"

	server "github.com/BazaarTrade/OrderMatchingService/internal/api/gRPC"
	"github.com/BazaarTrade/OrderMatchingService/internal/repository/postgresPgx"
	"github.com/BazaarTrade/OrderMatchingService/internal/service"
	"github.com/joho/godotenv"
)

func Run() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	if _, err := os.Stat("../.env"); err == nil {
		if err := godotenv.Load("../.env"); err != nil {
			logger.Error("failed to load .env file", "error", err)
			return
		}
	}

	logger.Info("starting aplication...")

	repo, err := postgresPgx.NewPostgres(logger)
	if err != nil {
		return
	}

	service := service.New(repo, logger)
	server := server.New(service, logger)

	if err := InitOrderBooks(server, service); err != nil {
		return
	}

	if err := server.Run(); err != nil {
		return
	}
}

func InitOrderBooks(server *server.Server, service *service.Service) error {
	pairs, err := service.GetPairs()
	if err != nil {
		return err
	}

	for _, pair := range pairs {
		service.CreateOrderBook(pair)
		server.InitChans(pair)
	}
	return nil
}
