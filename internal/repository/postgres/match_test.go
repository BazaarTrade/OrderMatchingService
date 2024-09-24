package postgres

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/BazaarTrade/OrderMatchingService/internal/repository"
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
	mock.ExpectQuery(`UPDATE orders SET sizeFilled = \$1 WHERE id = \$2 RETURNING id, userID, isBid, symbol, price, qty, sizeFilled, status, type, createdAt, closedAt`).
		WithArgs("0.5", int64(1)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "userID", "isBid", "symbol", "price", "qty", "sizeFilled", "status", "type", "createdAt", "closedAt"}).
			AddRow(int64(1), int64(1), true, "BTC/USDT", "10000", "1", "0.5", "filling", "limit", time.Now(), nil))

	mock.ExpectQuery(`UPDATE orders SET sizeFilled = \$1 WHERE id = \$2 RETURNING id, userID, isBid, symbol, price, qty, sizeFilled, status, type, createdAt, closedAt`).
		WithArgs("0.5", int64(1)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "userID", "isBid", "symbol", "price", "qty", "sizeFilled", "status", "type", "createdAt", "closedAt"}).
			AddRow(int64(1), int64(1), true, "BTC/USDT", "10000", "1", "0.5", "filling", "limit", time.Now(), nil))

	// Expectations for inserting match
	mock.ExpectExec(`INSERT INTO matches \(orderID, orderIDCounter, qty, price\) VALUES \(\$1, \$2, \$3, \$4\)`).
		WithArgs(int64(1), int64(2), "0.5", "10000").
		WillReturnResult(pgxmock.NewResult("INSERT", 1))

	// Expectations for transaction commit
	mock.ExpectCommit()

	// Call the method
	orders, err := pg.AddMatches(matchesReq)
	require.NoError(t, err)
	require.Len(t, orders, 2)
	require.Equal(t, "BTC/USDT", orders[0].Symbol)

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

	mock.ExpectQuery(`SELECT FROM matches\(qty, price\) WHERE orderID = \$1`).
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
