package postgres

import "github.com/Moha192/OrderMatchingService/internal/models"

func (p *Postres) AddMatches([]models.Match) error {
	return nil
}

func (p *Postres) GetMatches(orderID int) ([]models.Match, error) {
	return []models.Match{}, nil
}
