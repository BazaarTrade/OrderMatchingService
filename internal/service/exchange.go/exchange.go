package exchange

import (
	"github.com/Moha192/OrderMatchingService/internal/repository"
	"github.com/Moha192/OrderMatchingService/internal/service"
)

type Exchange struct {
	DB         *repository.Database
	OrderBooks map[string]*service.Service
}

func NewExchange(db *repository.Database) *Exchange {
	return &Exchange{
		DB:         db,
		OrderBooks: make(map[string]*service.Service),
	}
}
