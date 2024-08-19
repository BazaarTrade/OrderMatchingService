package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	DB *pgxpool.Pool
}

func NewPostgres(databaseConn string) (*Postgres, error) {
	conn, err := pgxpool.New(context.Background(), databaseConn)
	if err != nil {
		return nil, err
	}

	return &Postgres{
		DB: conn,
	}, nil
}
