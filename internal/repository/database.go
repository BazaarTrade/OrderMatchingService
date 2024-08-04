package repository

import "github.com/Moha192/OrderMatchingService/internal/models"

type Database interface {
	AddOrder(models.Order) int
	GetOrder(int) models.Order
	GetOrders(int) []models.Order

	AddMatches([]models.Match) error
	GetMatches(int) ([]models.Match, error)
}
