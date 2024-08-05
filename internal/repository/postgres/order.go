package postgres

import (
	"context"

	"github.com/Moha192/OrderMatchingService/internal/models"
)

func (p *Postres) AddOrder(order models.Order) (int, error) {
	var orderID int
	err := p.DB.QueryRow(context.Background(), `
	INSERT INTO orders
	(user_id, currency_pair, price, qty, order_type, is_bid, is_filled)
	VALUES
	($1, $2, $3, $4, $5, $6, $7)
	RETURNING
	order_id
	`).Scan(&orderID)
	if err != nil {
		return 0, err
	}
	return 0, nil
}

func (p *Postres) GetOrder(orderID int) (models.Order, error) {
	var order models.Order
	row := p.DB.QueryRow(context.Background(), `SELECT
	id, user_id, currency_pair, price, qty, orderd_type, is_bid, is_filled
	FROM orders
	WHERE id = $1
	`, orderID)

	err := row.Scan(&order.ID, &order.UserID, &order.CurrencyPair, &order.Price, &order.Qty, &order.OrderType, &order.IsBid, &order.IsFilled)
	if err != nil {
		return models.Order{}, err
	}
	return order, nil
}

func (p *Postres) GetOrders(userID int) ([]models.Order, error) {
	rows, err := p.DB.Query(context.Background(), `SELECT
	id, user_id, currency_pair, price, qty, orderd_type, is_bid, is_filled
	FROM orders
	WHERE user_id = $1
	`, userID)
	if err != nil {
		return []models.Order{}, err
	}

	var orders []models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(&order.ID, &order.UserID, &order.CurrencyPair, &order.Price, &order.Qty, &order.OrderType, &order.IsBid, &order.IsFilled)
		if err != nil {
			return []models.Order{}, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}
