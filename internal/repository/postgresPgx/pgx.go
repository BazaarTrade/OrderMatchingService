package postgresPgx

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

func New(DB_CONNECTION string, logger *slog.Logger) (*Postgres, error) {
	var (
		p   = &Postgres{logger: logger}
		err error
	)

	p.db, err = pgxpool.New(context.Background(), DB_CONNECTION)
	if err != nil {
		logger.Error("failed to create pgxPool connection", "error", err)
		return nil, err
	}

	err = p.createTables()
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Postgres) createTables() error {
	queries := []string{
		`CREATE SCHEMA IF NOT EXISTS quote;`,
		`CREATE SCHEMA IF NOT EXISTS matchingEngine;`,

		`CREATE TABLE IF NOT EXISTS quote.candlesticks (
			id SERIAL PRIMARY KEY,
			pair VARCHAR NOT NULL,
			timeframe VARCHAR NOT NULL,
			openTime TIMESTAMP NOT NULL,
			closeTime TIMESTAMP NOT NULL,
			openPrice VARCHAR NOT NULL,
			closePrice VARCHAR NOT NULL,
			highPrice VARCHAR NOT NULL,
			lowPrice VARCHAR NOT NULL,
			turnover VARCHAR NOT NULL,
			volume VARCHAR NOT NULL
		);`,

		`CREATE TABLE IF NOT EXISTS quote.pairs (
			pair VARCHAR PRIMARY KEY,
			orderBookPricePrecisions INTEGER[] NOT NULL,
			qtyPrecision INTEGER NOT NULL,
			candleStickTimeframes VARCHAR[] NOT NULL,
			createdAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,

		`CREATE TABLE IF NOT EXISTS matchingEngine.orders (
			id SERIAL PRIMARY KEY,
			userID INTEGER NOT NULL,
			isBid BOOLEAN NOT NULL,
			pair VARCHAR NOT NULL REFERENCES quote.pairs(pair),
			price VARCHAR NOT NULL,
			qty VARCHAR NOT NULL,
			type VARCHAR NOT NULL,
			sizeFilled VARCHAR NOT NULL DEFAULT '0',
			status VARCHAR NOT NULL DEFAULT 'filling',
			createdAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			closedAt TIMESTAMP
		);`,

		`CREATE TABLE IF NOT EXISTS matchingEngine.matches (
			orderID INTEGER NOT NULL REFERENCES matchingEngine.orders(id),
			orderIDCounter INTEGER NOT NULL REFERENCES matchingEngine.orders(id),
			qty VARCHAR NOT NULL,
			price VARCHAR NOT NULL,
			PRIMARY KEY (orderID, orderIDCounter)
		);`,

		`CREATE OR REPLACE FUNCTION matchingEngine.update_order_status()
		RETURNS TRIGGER AS $$
		BEGIN
			IF NEW.sizeFilled <> OLD.sizeFilled AND NEW.qty = NEW.sizeFilled THEN
				NEW.status := 'filled';
				NEW.closedAt := CURRENT_TIMESTAMP;
			END IF;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;`,

		`DROP TRIGGER IF EXISTS trigger_update_order_status ON matchingEngine.orders;
		
		CREATE TRIGGER trigger_update_order_status
		BEFORE UPDATE ON matchingEngine.orders
		FOR EACH ROW
		EXECUTE FUNCTION matchingEngine.update_order_status();`,

		`INSERT INTO quote.pairs (pair, orderBookPricePrecisions, qtyPrecision, candleStickTimeframes)
		VALUES ('BTC_USDT', ARRAY[2, 1, 0], 6, ARRAY['1s', '1m', '5m', '15m', '30m', '1h', '2h', '4h', '24h'])
		ON CONFLICT (pair) DO NOTHING;`,
	}

	for _, query := range queries {
		_, err := p.db.Exec(context.Background(), query)
		if err != nil {
			p.logger.Error("failed to execute query", "query", query, "error", err)
			return err
		}
	}

	return nil
}

func (p *Postgres) Close() {
	p.db.Close()
}
