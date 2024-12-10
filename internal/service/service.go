package service

import (
	"github.com/BazaarTrade/OrderMatchingService/internal/models"
)

type Service interface {
	AddOrderBook(pair string) error
	CreateOrderBook(pair string, pricePrecisions []int32, qtyPrecision int32) error
	DeleteOrderBook(pair string) error

	PlaceOrder(order models.PlaceOrderReq) (models.Order, []models.Order, models.OrderBookSnapshot, error)
	CancelOrder(orderID int) (models.Order, models.OrderBookSnapshot, error)

	GetOrders(userID int) ([]models.Order, error)
	GetCurrentOrders(userID int) ([]models.Order, error)

	GetPairs() ([]string, error)
	GetPairsParams() ([]models.PairParams, error)
	GetPairPricePrecisions(pair string) ([]int32, error)
}
