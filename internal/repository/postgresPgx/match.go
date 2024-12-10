package postgresPgx

import (
	"context"

	"github.com/BazaarTrade/OrderMatchingService/internal/models"
)

func (p *Postgres) AddMatch(orderID int, match models.Match) error {
	_, err := p.db.Exec(context.Background(), `
			INSERT INTO matches (orderID, orderIDCounter, qty, price)
			VALUES ($1, $2, $3, $4)
		`, orderID, match.OrderID, match.Qty, match.Price)
	if err != nil {
		p.logger.Error("failed to add matches", "orderID", orderID, "error", err)
		return err
	}
	return nil
}

func (p *Postgres) GetMatches(orderID int) ([]models.Match, error) {
	rows, err := p.db.Query(context.Background(), `
	SELECT FROM matches(qty, price) WHERE orderID = $1
	`, orderID)
	if err != nil {
		p.logger.Error("failed to select matches", "error", err)
		return nil, err
	}

	var mathces []models.Match
	for rows.Next() {
		var match models.Match
		err := rows.Scan(&match.Qty, &match.Price)
		if err != nil {
			p.logger.Error("failed to scan matches", "error", err)
			return nil, err
		}
		mathces = append(mathces, match)
	}
	return mathces, nil
}
