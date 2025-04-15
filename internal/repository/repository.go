package repository

import (
	"github.com/BazaarTrade/OrderMatchingService/internal/models"
)

type Repository interface {
	CreateOrder(order models.PlaceOrderReq) (int, error)
	GetOrderByOrderID(orderID int) (models.Order, error)

	SetOrderStatusToError(orderID int) error
	SetOrderStatusToCancel(orderID int) error

	AddMatch(orderID int, match models.Match) error
	GetMatches(orderID int) ([]models.Match, error)

	CreatePair(pair string, pricePrecisions []int32, qtyPecision int32) error
	GetPairs() ([]string, error)

	UpdateOrderPrice(orderID int, price string) error
	UpdateOrderSizeFilled(orderID int, sizeFilled string) (models.Order, error)
}
