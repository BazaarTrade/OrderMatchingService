package service

import (
	"github.com/Moha192/OrderMatchingService/internal/models"
)

type Exchanger interface {
	AddOrderBook(symbol string) error
	DeleteOrderBook(symbol string) error

	PlaceOrder(order models.PlaceOrderReq) ([]models.Order, error)
	CancelOrder(orderID int64) error
}
