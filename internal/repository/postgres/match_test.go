package postgres

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/Moha192/OrderMatchingService/internal/repository"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

func TestAddMatches(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	pg := &Postgres{
		db:     mock,
		logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}

	matchesReq := repository.AddMatchesReq{
		OrderID:         1,
		OrderSizeFilled: "0.5",
		Matches: []repository.Match{
			{CounterOrderID: 2, Qty: "0.5", Price: "10000"},
		},
	}

	// Expectations for transaction begin
	mock.ExpectBegin()

	// Expectations for updating order
	mock.ExpectQuery(`UPDATE orders SET size_filled = \$1 WHERE id = \$2 RETURNING id, user_id, is_bid, symbol, price, qty, size_filled, status, type, created_at, closed_at`).
		WithArgs("0.5", int64(1)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "is_bid", "symbol", "price", "qty", "size_filled", "status", "type", "created_at", "closed_at"}).
			AddRow(int64(1), int64(1), true, "BTC_USDT", "10000", "1", "0.5", "filling", "limit", time.Now(), nil))

	mock.ExpectQuery(`UPDATE orders SET size_filled = \$1 WHERE id = \$2 RETURNING id, user_id, is_bid, symbol, price, qty, size_filled, status, type, created_at, closed_at`).
		WithArgs("0.5", int64(1)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "is_bid", "symbol", "price", "qty", "size_filled", "status", "type", "created_at", "closed_at"}).
			AddRow(int64(1), int64(1), true, "BTC_USDT", "10000", "1", "0.5", "filling", "limit", time.Now(), nil))

	// Expectations for inserting match
	mock.ExpectExec(`INSERT INTO matches \(order_id, order_id_counter, qty, price\) VALUES \(\$1, \$2, \$3, \$4\)`).
		WithArgs(int64(1), int64(2), "0.5", "10000").
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	// Expectations for transaction commit
	mock.ExpectCommit()

	// Call the method
	orders, err := pg.AddMatches(matchesReq)
	require.NoError(t, err)
	require.Len(t, orders, 2)
	require.Equal(t, "BTC_USDT", orders[0].Symbol)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetMatches(t *testing.T) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	defer mock.Close()

	pg := &Postgres{
		db:     mock,
		logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}

	mock.ExpectQuery(`SELECT FROM matches\(qty, price\) WHERE order_id = \$1`).
		WithArgs(int64(1)).
		WillReturnRows(pgxmock.NewRows([]string{"qty", "price"}).
			AddRow("0.5", "10000").
			AddRow("0.3", "9900"))

	matches, err := pg.GetMatches(1)
	require.NoError(t, err)
	require.Len(t, matches, 2)
	require.Equal(t, "10000", matches[0].Price)
	require.Equal(t, "0.5", matches[0].Qty)

	require.NoError(t, mock.ExpectationsWereMet())
}
