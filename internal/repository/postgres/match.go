package postgres

import (
	"context"

	"github.com/Moha192/OrderMatchingService/internal/models"
	"github.com/jackc/pgx/v5"
)

func (p *Postres) AddMatches(matches []models.Match) error {
	table := "matches"
	columns := []string{"order_id_bid", "order_id_ask", "qty", "price"}

	data := make([][]interface{}, len(matches))
	for i, match := range matches {
		data[i] = []interface{}{match.OrderIDBid, match.OrderIDAsk, match.Qty, match.Price}
	}

	_, err := p.DB.CopyFrom(
		context.Background(),
		pgx.Identifier{table},
		columns,
		pgx.CopyFromRows(data),
	)
	if err != nil {
		return err
	}
	return nil
}

func (p *Postres) GetMatches(orderID int) ([]models.Match, error) {
	rows, err := p.DB.Query(context.Background(), `
	SELECT FROM mathces(order_id_bid, order_id_ask, qty, price) WHERE order_id_bid OR order_id_ask = $1
	`, orderID)
	if err != nil {
		return []models.Match{}, nil
	}

	var mathces []models.Match
	for rows.Next() {
		var match models.Match
		err := rows.Scan(&match.OrderIDBid, &match.OrderIDAsk, &match.Qty, &match.Price)
		if err != nil {
			return []models.Match{}, err
		}
		mathces = append(mathces, match)
	}
	return mathces, nil
}
