package app

import (
	"log/slog"
	"os"

	"github.com/Moha192/OrderMatchingService/internal/api/gRPC"
	"github.com/Moha192/OrderMatchingService/internal/repository/postgres"
	"github.com/Moha192/OrderMatchingService/internal/service/exchange.go"
)

func Run() {
	handler := slog.NewTextHandler(os.Stdout, nil)
	logger := slog.New(handler)

	logger.Info("Starting aplication")

	repo, err := postgres.NewPostgres("user=postgres password=postgres dbname=postgres sslmode=disable host=localhost port=5432", logger)
	if err != nil {
		logger.Error("Failed to initialize database", "error", err)
		return
	}

	service := exchange.NewExchange(repo, logger)
	server := gRPC.NewServer(service, logger)
	server.StartGRPCServer()
}
