package postgres

import "github.com/Moha192/OrderMatchingService/internal/models"

func (p *Postres) AddOrder(order models.Order) (int, error) {
	return 0, nil
}

func (p *Postres) GetOrder(userID int) (models.Order, error) {
	return models.Order{}, nil
}

func (p *Postres) GetOrders(userID int) (models.Order, error) {
	return models.Order{}, nil
}
