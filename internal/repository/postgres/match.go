package postgres

import (
	"context"

	"github.com/Moha192/OrderMatchingService/internal/models"
	"github.com/Moha192/OrderMatchingService/internal/repository"
)

func (p *Postgres) AddMatches(matches repository.AddMatchesReq) ([]models.Order, error) {
	tx, err := p.DB.Begin(context.Background())
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(context.Background())

	var updatedOrder models.Order
	err = tx.QueryRow(context.Background(), `
		UPDATE orders
		SET size_filled = $1
		WHERE id = $2
		RETURNING id, user_id, is_bid, symbol, price, qty, size_filled, status, type, created_at, closed_at
	`, matches.OrderSizeFilled, matches.OrderID).Scan(
		&updatedOrder.ID,
		&updatedOrder.UserID,
		&updatedOrder.IsBid,
		&updatedOrder.Symbol,
		&updatedOrder.Price,
		&updatedOrder.Qty,
		&updatedOrder.SizeFilled,
		&updatedOrder.Status,
		&updatedOrder.Type,
		&updatedOrder.CreatedAt,
		&updatedOrder.ClosedAt,
	)
	if err != nil {
		return nil, err
	}

	var updatedOrders []models.Order
	updatedOrders = append(updatedOrders, updatedOrder)

	for _, match := range matches.Matches {
		var counterOrder models.Order
		err = tx.QueryRow(context.Background(), `
			UPDATE orders
			SET size_filled = $1
			WHERE id = $2
			RETURNING id, user_id, is_bid, symbol, price, qty, size_filled, status, type, created_at, closed_at
		`, match.CounterOrderSizeFilled, match.CounterOrderID).Scan(
			&counterOrder.ID,
			&counterOrder.UserID,
			&counterOrder.IsBid,
			&counterOrder.Symbol,
			&counterOrder.Price,
			&counterOrder.Qty,
			&counterOrder.SizeFilled,
			&counterOrder.Status,
			&counterOrder.Type,
			&counterOrder.CreatedAt,
			&counterOrder.ClosedAt,
		)
		if err != nil {
			return nil, err
		}
		updatedOrders = append(updatedOrders, counterOrder)

		_, err = tx.Exec(context.Background(), `
			INSERT INTO matches (order_id, order_id_counter, qty, price)
			VALUES ($1, $2, $3, $4)
		`, matches.OrderID, match.CounterOrderID, match.Qty, match.Price)
		if err != nil {
			return nil, err
		}
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return nil, err
	}

	return updatedOrders, nil
}

func (p *Postgres) GetMatches(orderID int64) ([]models.Match, error) {
	rows, err := p.DB.Query(context.Background(), `
	SELECT FROM mathces(qty, price) WHERE order_id = $1
	`, orderID)
	if err != nil {
		return nil, err
	}

	var mathces []models.Match
	for rows.Next() {
		var match models.Match
		err := rows.Scan(&match.Qty, &match.Price)
		if err != nil {
			return nil, err
		}
		mathces = append(mathces, match)
	}
	return mathces, nil
}
