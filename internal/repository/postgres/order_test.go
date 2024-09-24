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
		WithArgs(int64(1), true, "BTC/USDT", "10000", "1", "limit", "filling").
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(1))

	orderID, err := pg.CreateOrder(models.PlaceOrderReq{
		UserID: 1,
		IsBid:  true,
		Symbol: "BTC/USD",
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

	mock.ExpectQuery(`SELECT id,\s+userID,\s+isBid,\s+symbol,\s+price,\s+qty,\s+sizeFilled,\s+status,\s+type,\s+createdAt,\s+closedAt\s+FROM\s+orders\s+WHERE\s+id\s+=\s+\$1`).
		WithArgs(int64(1)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "userID", "isBid", "symbol", "price", "qty", "sizeFilled", "status", "type", "createdAt", "closedAt"}).
			AddRow(int64(1), int64(1), true, "BTC/USDT", "10000", "1", "0", "filling", "limit", time.Now(), nil))

	order, err := pg.GetOrderByOrderID(int64(1))
	require.NoError(t, err)
	require.Equal(t, int64(1), order.ID)
	require.Equal(t, "BTC/USDT", order.Symbol)
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

	mock.ExpectQuery(`SELECT id,\s+userID,\s+isBid,\s+symbol,\s+price,\s+qty,\s+sizeFilled,\s+status,\s+type,\s+createdAt,\s+closedAt\s+FROM\s+orders\s+WHERE\s+userID\s+=\s+\$1`).
		WithArgs(int64(1)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "userID", "isBid", "symbol", "price", "qty", "sizeFilled", "status", "type", "createdAt", "closedAt"}).
			AddRow(int64(1), int64(1), true, "BTC/USDT", "10000", "1", "0", "filling", "limit", time.Now(), nil))

	orders, err := pg.GetOrdersByUser(int64(1))
	require.NoError(t, err)
	require.Len(t, orders, 1)
	require.Equal(t, "BTC/USDT", orders[0].Symbol)
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
	mock.ExpectQuery(`SELECT id, userID, isBid, symbol, price, qty, sizeFilled, status, type, createdAt, closedAt FROM orders WHERE userID = \$1 AND status IN \('filling', 'canceled'\)`).
		WithArgs(int64(1)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "userID", "isBid", "symbol", "price", "qty", "sizeFilled", "status", "type", "createdAt", "closedAt"}).
			AddRow(int64(1), int64(1), true, "BTC/USDT", "10000", "1", "0", "filling", "limit", time.Now(), nil))

	orders, err := pg.GetNotFilledOrdersByUser(1)
	require.NoError(t, err)
	require.Len(t, orders, 1)
	require.Equal(t, "BTC/USDT", orders[0].Symbol)
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

	mock.ExpectExec(`UPDATE orders SET status = 'canceled', closedAt = CURRENT_TIMESTAMP WHERE id = \$1`).
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

	mock.ExpectExec(`UPDATE orders SET status = 'error', closedAt = CURRENT_TIMESTAMP WHERE id = \$1`).
		WithArgs(int64(1)).
		WillReturnResult(pgxmock.NewResult("UPDATE", int64(1)))

	err = pg.SetOrderStatusToError(int64(1))
	require.NoError(t, err)

	require.NoError(t, mock.ExpectationsWereMet())
}
