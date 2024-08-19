package postgres

import (
	"context"

	"github.com/Moha192/OrderMatchingService/internal/models"
)

func (p *Postgres) CreateOrder(order models.PlaceOrderReq) (int64, error) {
	var orderID int64
	err := p.DB.QueryRow(context.Background(), `
	INSERT INTO orders
	(user_id, is_bid, symbol, price, qty, type, status)
	VALUES
	($1, $2, $3, $4, $5, $6, $7)
	RETURNING
	order_id
	`, order.UserID, order.IsBid, order.Symbol, order.Price, order.Qty, order.Type, "Filling").Scan(&orderID)
	if err != nil {
		return 0, err
	}
	return orderID, nil
}

func (p *Postgres) GetOrderByOrderID(orderID int64) (models.Order, error) {
	var order models.Order
	row := p.DB.QueryRow(context.Background(), `SELECT
	id, user_id, is_bid, symbol, price, qty, size_filled, status, orderd_type, created_at, closed_at
	FROM orders
	WHERE id = $1
	`, orderID)

	err := row.Scan(
		&order.ID,
		&order.UserID,
		&order.IsBid,
		&order.Symbol,
		&order.Price,
		&order.Qty,
		&order.SizeFilled,
		&order.Status,
		&order.Type,
		&order.CreatedAt,
		&order.ClosedAt,
	)
	if err != nil {
		return models.Order{}, err
	}
	return order, nil
}

func (p *Postgres) GetOrdersByUser(userID int64) ([]models.Order, error) {
	rows, err := p.DB.Query(context.Background(), `SELECT
	id, user_id, is_bid, symbol, price, qty, size_filled, status, orderd_type, created_at, closed_at
	FROM orders
	WHERE user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.IsBid,
			&order.Symbol,
			&order.Price,
			&order.Qty,
			&order.SizeFilled,
			&order.Status,
			&order.Type,
			&order.CreatedAt,
			&order.ClosedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func (p *Postgres) GetNotFilledOrdersByUser(userID int64) ([]models.Order, error) {
	rows, err := p.DB.Query(context.Background(), `SELECT
	id, user_id, is_bid, symbol, price, qty, size_filled, status, orderd_type, created_at, closed_at
	FROM orders
	WHERE user_id = $1 AND status IN ('Filling', 'Canceled')
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.IsBid,
			&order.Symbol,
			&order.Price,
			&order.Qty,
			&order.SizeFilled,
			&order.Status,
			&order.Type,
			&order.CreatedAt,
			&order.ClosedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}
