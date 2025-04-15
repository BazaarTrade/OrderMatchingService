package postgresPgx

import (
	"context"

	"github.com/BazaarTrade/OrderMatchingService/internal/models"
)

func (p *Postgres) CreateOrder(order models.PlaceOrderReq) (int, error) {
	var orderID int
	err := p.db.QueryRow(context.Background(), `
	INSERT INTO matchingEngine.orders
	(userID, isBid, pair, price, qty, type, status)
	VALUES
	($1, $2, $3, $4, $5, $6, $7)
	RETURNING
	id
	`, order.UserID, order.IsBid, order.Pair, order.Price, order.Qty, order.Type, "filling").Scan(&orderID)
	if err != nil {
		p.logger.Error("failed to insert order", "error", err)
		return 0, err
	}
	return orderID, nil
}

func (p *Postgres) GetOrderByOrderID(orderID int) (models.Order, error) {
	var order models.Order
	row := p.db.QueryRow(context.Background(), `
	SELECT id, userID, isBid, pair, price, qty, sizeFilled, status, type, createdAt, closedAt
	FROM matchingEngine.orders
	WHERE id = $1
	`, orderID)

	err := row.Scan(
		&order.ID,
		&order.UserID,
		&order.IsBid,
		&order.Pair,
		&order.Price,
		&order.Qty,
		&order.SizeFilled,
		&order.Status,
		&order.Type,
		&order.CreatedAt,
		&order.ClosedAt,
	)
	if err != nil {
		p.logger.Error("failed to scan order", "error", err)
		return models.Order{}, err
	}
	return order, nil
}

func (p *Postgres) SetOrderStatusToCancel(orderID int) error {
	_, err := p.db.Exec(context.Background(), `
	UPDATE matchingEngine.orders SET status = 'canceled', closedAt = CURRENT_TIMESTAMP 
	WHERE id = $1
	`, orderID)
	if err != nil {
		p.logger.Error("failed to update order", "error", err)
		return err
	}
	return nil
}

func (p *Postgres) SetOrderStatusToError(orderID int) error {
	_, err := p.db.Exec(context.Background(), `
	UPDATE matchingEngine.orders SET status = 'error', closedAt = CURRENT_TIMESTAMP 
	WHERE id = $1
	`, orderID)
	if err != nil {
		p.logger.Error("failed to update order status to error", "error", err)
		return err
	}
	return nil
}

func (p *Postgres) UpdateOrderPrice(orderID int, price string) error {
	_, err := p.db.Exec(context.Background(), `
	UPDATE matchingEngine.orders SET price = $1
	WHERE id = $2
	`, price, orderID)
	if err != nil {
		p.logger.Error("failed to update order price", "error", err)
		return err
	}
	return nil
}

func (p *Postgres) UpdateOrderSizeFilled(orderID int, sizeFilled string) (models.Order, error) {
	var order models.Order
	err := p.db.QueryRow(context.Background(), `
		UPDATE matchingEngine.orders
		SET sizeFilled = $1
		WHERE id = $2
		RETURNING id, userID, isBid, pair, price, qty, sizeFilled, status, type, createdAt, closedAt
	`, sizeFilled, orderID).Scan(
		&order.ID, &order.UserID, &order.IsBid, &order.Pair, &order.Price,
		&order.Qty, &order.SizeFilled, &order.Status, &order.Type,
		&order.CreatedAt, &order.ClosedAt,
	)
	if err != nil {
		p.logger.Error("failed to update order sizeFilled", "error", err)
		return models.Order{}, err
	}
	return order, nil
}
