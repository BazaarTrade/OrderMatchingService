package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postres struct {
	DB *pgxpool.Pool
}

func NewPostgres(databaseConn string) (*Postres, error) {
	conn, err := pgxpool.New(context.Background(), databaseConn)
	if err != nil {
		return nil, err
	}

	return &Postres{
		DB: conn,
	}, nil
}
