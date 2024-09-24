package postgres

import (
	"context"

	"github.com/BazaarTrade/OrderMatchingService/internal/models"
)

func (p *Postgres) CreateOrder(order models.PlaceOrderReq) (int64, error) {
	var orderID int
	err := p.db.QueryRow(context.Background(), `
	INSERT INTO orders
	(userID, isBid, symbol, price, qty, type, status)
	VALUES
	($1, $2, $3, $4, $5, $6, $7)
	RETURNING
	id
	`, order.UserID, order.IsBid, order.Symbol, order.Price, order.Qty, order.Type, "filling").Scan(&orderID)
	if err != nil {
		p.logger.Error("Error inserting order", "error", err)
		return 0, err
	}
	return int64(orderID), nil
}

func (p *Postgres) GetOrderByOrderID(orderID int64) (models.Order, error) {
	var order models.Order
	row := p.db.QueryRow(context.Background(), `
	SELECT id, userID, isBid, symbol, price, qty, sizeFilled, status, type, createdAt, closedAt
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
		p.logger.Error("Error scanning order", "error", err)
		return models.Order{}, err
	}
	return order, nil
}

func (p *Postgres) GetOrdersByUser(userID int64) ([]models.Order, error) {
	rows, err := p.db.Query(context.Background(), `
	SELECT id, userID, isBid, symbol, price, qty, sizeFilled, status, type, createdAt, closedAt
	FROM orders
	WHERE userID = $1
	`, userID)
	if err != nil {
		p.logger.Error("Error selecting orders", "error", err)
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
			p.logger.Error("Error scanning order", "error", err)
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func (p *Postgres) GetNotFilledOrdersByUser(userID int64) ([]models.Order, error) {
	rows, err := p.db.Query(context.Background(), `
	SELECT id, userID, isBid, symbol, price, qty, sizeFilled, status, type, createdAt, closedAt
	FROM orders
	WHERE userID = $1 AND status IN ('filling', 'canceled')
	`, userID)
	if err != nil {
		p.logger.Error("Error selecting orders", "error", err)
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
			p.logger.Error("Error scanning order", "error", err)
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

func (p *Postgres) SetOrderStatusToCancel(orderID int64) error {
	_, err := p.db.Exec(context.Background(), `
	UPDATE orders SET status = 'canceled', closedAt = CURRENT_TIMESTAMP 
	WHERE id = $1
	`, orderID)
	if err != nil {
		p.logger.Error("Error updating order", "error", err)
		return err
	}
	return nil
}

func (p *Postgres) SetOrderStatusToError(orderID int64) error {
	_, err := p.db.Exec(context.Background(), `
	UPDATE orders SET status = 'error', closedAt = CURRENT_TIMESTAMP 
	WHERE id = $1
	`, orderID)
	if err != nil {
		p.logger.Error("Error updating order", "error", err)
		return err
	}
	return nil
}
