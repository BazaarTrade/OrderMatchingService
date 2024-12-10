package postgresPgx

import (
	"context"
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

func NewPostgres(logger *slog.Logger) (*Postgres, error) {
	conn, err := pgxpool.New(context.Background(), os.Getenv("DB_CONNECTION"))
	if err != nil {
		logger.Error("failed to create pgxPool connection", "error", err)
		return nil, err
	}

	m, err := migrate.New(os.Getenv("MIGRATIONS"), os.Getenv("DB_CONNECTION_URL"))
	if err != nil {
		logger.Error("failed to create migrate instance", "error", err)
		return nil, err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Error("failed to apply migrations", "error", err)
		return nil, err
	}

	return &Postgres{
		db:     conn,
		logger: logger,
	}, nil
}
