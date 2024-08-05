package exchange

import (
	"github.com/Moha192/OrderMatchingService/internal/repository"
	"github.com/Moha192/OrderMatchingService/internal/service"
)

type Exchange struct {
	DB         *repository.Database
	OrderBooks map[string]*service.OrderService
}

func NewExchange(db *repository.Database) *Exchange {
	return &Exchange{
		DB:         db,
		OrderBooks: make(map[string]*service.OrderService),
	}
}

func (e *Exchange) AddOrderBook(pair string) {

}

func (e *Exchange) DeleteOrderBook(pair string) {

}
