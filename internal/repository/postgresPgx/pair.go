package postgresPgx

import "context"

func (p *Postgres) CreatePair(pair string, pricePrecisions []int32, qtyPrecision int32) error {
	_, err := p.db.Exec(context.Background(), `
	INSERT INTO quote.pairs
	(pair, pricePrecisions, qtyPrecision)
	VALUES
	($1, $2, $3)
	`, pair, pricePrecisions, qtyPrecision)
	if err != nil {
		p.logger.Error("failed to insert pair", "error", err)
		return err
	}
	return nil
}

func (p *Postgres) GetPairs() ([]string, error) {
	rows, err := p.db.Query(context.Background(), `
	SELECT pair
	FROM quote.pairs
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
