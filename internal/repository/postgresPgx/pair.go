package postgresPgx

import (
	"context"

	"github.com/BazaarTrade/OrderMatchingService/internal/models"
)

func (p *Postgres) CreatePair(pair string, pricePrecisions []int32, qtyPecision int32) error {
	_, err := p.db.Exec(context.Background(), `
	INSERT INTO pairs
	(pair, pricePrecisions, qtyPecision)
	VALUES
	($1, $2, $3)
	`, pair, pricePrecisions, qtyPecision)
	if err != nil {
		p.logger.Error("failed to insert pair", "error", err)
		return err
	}
	return nil
}

func (p *Postgres) GetPairs() ([]string, error) {
	rows, err := p.db.Query(context.Background(), `
	SELECT pair
	FROM pairs
	`)
	if err != nil {
		p.logger.Error("failed to select pairs", "error", err)
		return nil, err
	}
	defer rows.Close()

	var pairs []string
	for rows.Next() {
		var pair string
		err := rows.Scan(&pair)
		if err != nil {
			p.logger.Error("failed to scan pair", "error", err)
			return nil, err
		}
		pairs = append(pairs, pair)
	}
	return pairs, nil
}

func (p *Postgres) GetPairPricePrecisions(pair string) ([]int32, error) {
	var pricePrecisions []int32
	err := p.db.QueryRow(context.Background(), `
	SELECT pricePrecisions
	FROM pairs
	WHERE pair = $1
	`, pair).Scan(&pricePrecisions)
	if err != nil {
		p.logger.Error("failed to select price precisions", "error", err)
		return nil, err
	}
	return pricePrecisions, nil
}

func (p *Postgres) GetPairsParams() ([]models.PairParams, error) {
	rows, err := p.db.Query(context.Background(), `
	SELECT pair, pricePrecisions, qtyPecision
	FROM pairs
	`)
	if err != nil {
		p.logger.Error("failed to select pairs", "error", err)
		return nil, err
	}
	defer rows.Close()

	var pairsParams []models.PairParams
	for rows.Next() {
		var pairParams models.PairParams
		err := rows.Scan(&pairParams.Pair, &pairParams.PricePrecisions, &pairParams.QtyPecision)
		if err != nil {
			p.logger.Error("failed to scan pair", "error", err)
			return nil, err
		}
		pairsParams = append(pairsParams, pairParams)
	}
	return pairsParams, nil
}
