package repository

import (
	"github.com/Moha192/OrderMatchingService/internal/models"
)

type Storer interface {
	CreateOrder(order models.PlaceOrderReq) (int64, error)
	GetOrdersByUser(userID int64) ([]models.Order, error)
	GetOrderByOrderID(orderID int64) (models.Order, error)
	SetStatusToCancel(orderID int64) (models.Order, error)
	GetNotFilledOrdersByUser(userID int64) ([]models.Order, error)

	AddMatches(matches AddMatchesReq) ([]models.Order, error)
	GetMatches(orderID int64) ([]models.Match, error)
}

type AddMatchesReq struct {
	OrderID         int64
	OrderSizeFilled string
	Matches         []Match
}

type Match struct {
	Qty                    string
	Price                  string
	CounterOrderID         int64
	CounterOrderSizeFilled string
}
