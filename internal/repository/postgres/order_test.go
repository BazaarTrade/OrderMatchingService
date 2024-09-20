package postgres

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/BazaarTrade/OrderMatchingService/internal/models"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/require"
)

func TestCreateOrder(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	pg := &Postgres{
		db:     mock,
		logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}

	mock.ExpectQuery(`INSERT INTO orders.*RETURNING id`).
		WithArgs(int64(1), true, "BTC_USD", "10000", "1", "limit", "filling").
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))

	orderID, err := pg.CreateOrder(models.PlaceOrderReq{
		UserID: 1,
		IsBid:  true,
		Symbol: "BTC_USD",
		Price:  "10000",
		Qty:    "1",
		Type:   "limit",
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), orderID)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOrderByOrderID(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	pg := &Postgres{
		db:     mock,
		logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}

	mock.ExpectQuery(`SELECT id,\s+user_id,\s+is_bid,\s+symbol,\s+price,\s+qty,\s+size_filled,\s+status,\s+type,\s+created_at,\s+closed_at\s+FROM\s+orders\s+WHERE\s+id\s+=\s+\$1`).
		WithArgs(int64(1)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "is_bid", "symbol", "price", "qty", "size_filled", "status", "type", "created_at", "closed_at"}).
			AddRow(int64(1), int64(1), true, "BTC_USDT", "10000", "1", "0", "filling", "limit", time.Now(), nil))

	order, err := pg.GetOrderByOrderID(int64(1))
	require.NoError(t, err)
	require.Equal(t, int64(1), order.ID)
	require.Equal(t, "BTC_USDT", order.Symbol)
	require.Equal(t, "10000", order.Price)
	require.Equal(t, "1", order.Qty)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetOrdersByUser(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	pg := &Postgres{
		db:     mock,
		logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}

	mock.ExpectQuery(`SELECT id,\s+user_id,\s+is_bid,\s+symbol,\s+price,\s+qty,\s+size_filled,\s+status,\s+type,\s+created_at,\s+closed_at\s+FROM\s+orders\s+WHERE\s+user_id\s+=\s+\$1`).
		WithArgs(int64(1)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "is_bid", "symbol", "price", "qty", "size_filled", "status", "type", "created_at", "closed_at"}).
			AddRow(int64(1), int64(1), true, "BTC_USDT", "10000", "1", "0", "filling", "limit", time.Now(), nil))

	orders, err := pg.GetOrdersByUser(int64(1))
	require.NoError(t, err)
	require.Len(t, orders, 1)
	require.Equal(t, "BTC_USDT", orders[0].Symbol)
	require.Equal(t, "10000", orders[0].Price)
	require.Equal(t, "1", orders[0].Qty)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetNotFilledOrdersByUser(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	pg := &Postgres{
		db:     mock,
		logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}

	// Экранируем скобки в регулярном выражении для IN
	mock.ExpectQuery(`SELECT id, user_id, is_bid, symbol, price, qty, size_filled, status, type, created_at, closed_at FROM orders WHERE user_id = \$1 AND status IN \('filling', 'canceled'\)`).
		WithArgs(int64(1)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "user_id", "is_bid", "symbol", "price", "qty", "size_filled", "status", "type", "created_at", "closed_at"}).
			AddRow(int64(1), int64(1), true, "BTC_USDT", "10000", "1", "0", "filling", "limit", time.Now(), nil))

	orders, err := pg.GetNotFilledOrdersByUser(1)
	require.NoError(t, err)
	require.Len(t, orders, 1)
	require.Equal(t, "BTC_USDT", orders[0].Symbol)
	require.Equal(t, "10000", orders[0].Price)
	require.Equal(t, "1", orders[0].Qty)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSetOrderStatusToCancel(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	pg := &Postgres{
		db:     mock,
		logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}

	mock.ExpectExec(`UPDATE orders SET status = 'canceled', closed_at = CURRENT_TIMESTAMP WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnResult(pgxmock.NewResult("UPDATE", int64(1)))

	err = pg.SetOrderStatusToCancel(int64(1))
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestSetOrderStatusToError(t *testing.T) {
	mock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	pg := &Postgres{
		db:     mock,
		logger: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}

	mock.ExpectExec(`UPDATE orders SET status = 'error', closed_at = CURRENT_TIMESTAMP WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnResult(pgxmock.NewResult("UPDATE", int64(1)))

	err = pg.SetOrderStatusToError(int64(1))
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}
