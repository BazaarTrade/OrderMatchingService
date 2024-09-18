package postgres

import (
	"context"

	"github.com/Moha192/OrderMatchingService/internal/models"
	"github.com/Moha192/OrderMatchingService/internal/repository"
)

func (p *Postgres) AddMatches(matches repository.AddMatchesReq) ([]models.Order, error) {
	tx, err := p.db.Begin(context.Background())
	if err != nil {
		p.logger.Error("Error creating transaction", "error", err)
		return nil, err
	}
	defer tx.Rollback(context.Background())

	updateOrder := func(sizeFilled string, orderID int64) (models.Order, error) {
		var order models.Order
		err := tx.QueryRow(context.Background(), `
			UPDATE orders
			SET size_filled = $1
			WHERE id = $2
			RETURNING id, user_id, is_bid, symbol, price, qty, size_filled, status, type, created_at, closed_at
		`, sizeFilled, orderID).Scan(
			&order.ID, &order.UserID, &order.IsBid, &order.Symbol, &order.Price,
			&order.Qty, &order.SizeFilled, &order.Status, &order.Type,
			&order.CreatedAt, &order.ClosedAt,
		)
		if err != nil {
			return models.Order{}, err
		}
		return order, nil
	}

	updatedOrder, err := updateOrder(matches.OrderSizeFilled, matches.OrderID)
	if err != nil {
		p.logger.Error("Error updating size_filled", "error", err)
		return nil, err
	}

	var updatedOrders = []models.Order{updatedOrder}

	for _, match := range matches.Matches {
		updatedOrder, err = updateOrder(matches.OrderSizeFilled, matches.OrderID)
		if err != nil {
			p.logger.Error("Error updating size_filled", "error", err)
			return nil, err
		}
		updatedOrders = append(updatedOrders, updatedOrder)

		_, err = tx.Exec(context.Background(), `
			INSERT INTO matches (order_id, order_id_counter, qty, price)
			VALUES ($1, $2, $3, $4)
		`, matches.OrderID, match.CounterOrderID, match.Qty, match.Price)
		if err != nil {
			p.logger.Error("Error inserting matches", "error", err)
			return nil, err
		}
	}

	err = tx.Commit(context.Background())
	if err != nil {
		p.logger.Error("Error commiting transaction", "error", err)
		return nil, err
	}

	return updatedOrders, nil
}

func (p *Postgres) GetMatches(orderID int64) ([]models.Match, error) {
	rows, err := p.db.Query(context.Background(), `
	SELECT FROM matches(qty, price) WHERE order_id = $1
	`, orderID)
	if err != nil {
		p.logger.Error("Error selecting matches", "error", err)
		return nil, err
	}

	var mathces []models.Match
	for rows.Next() {
		var match models.Match
		err := rows.Scan(&match.Qty, &match.Price)
		if err != nil {
			p.logger.Error("Error scanning matches", "error", err)
			return nil, err
		}
		mathces = append(mathces, match)
	}
	return mathces, nil
}
