package main

import (
	"log"

	"github.com/Moha192/OrderMatchingService/internal/api/gRPC"
	"github.com/Moha192/OrderMatchingService/internal/repository/postgres"
	"github.com/Moha192/OrderMatchingService/internal/service/exchange.go"
)

func main() {
	db, err := postgres.NewPostgres("user=postgres password=postgres dbname=postgres sslmode=disable host=localhost port=5432")
	if err != nil {
		log.Println(err)
	}

	service := exchange.NewExchange(db)

	gRPC.StartGRPCServer(service)
}
