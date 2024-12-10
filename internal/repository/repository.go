package repository

import (
	"github.com/BazaarTrade/OrderMatchingService/internal/models"
)

type Repository interface {
	CreateOrder(order models.PlaceOrderReq) (int, error)
	GetOrdersByUser(userID int) ([]models.Order, error)
	GetOrderByOrderID(orderID int) (models.Order, error)
	GetNotFilledOrdersByUser(userID int) ([]models.Order, error)

	SetOrderStatusToError(orderID int) error
	SetOrderStatusToCancel(orderID int) error

	AddMatch(orderID int, matche models.Match) error
	GetMatches(orderID int) ([]models.Match, error)

	CreatePair(pair string, pricePrecisions []int32, qtyPecision int32) error
	GetPairs() ([]string, error)
	GetPairsParams() ([]models.PairParams, error)
	GetPairPricePrecisions(pair string) ([]int32, error)

	UpdateOrderPrice(orderID int, price string) error
	UpdateOrderSizeFilled(orderID int, sizeFilled string) (models.Order, error)
}
