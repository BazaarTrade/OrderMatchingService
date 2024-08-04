package exchange

import (
	"log/slog"

	"github.com/shopspring/decimal"
)

type Limit struct {
	Price     decimal.Decimal
	Orders    []*Order
	TotalSize decimal.Decimal
}

func NewLimit(price decimal.Decimal) *Limit {
	return &Limit{
		Price:     price,
		Orders:    []*Order{},
		TotalSize: decimal.NewFromFloat(0),
	}
}

func (l *Limit) matchOrders(order *Order) bool {
	var countFilledOrders int
	defer func() {
		if countFilledOrders > 0 {
			l.removeEmptyOrders(countFilledOrders)
		}
	}()

	for _, bestOrder := range l.Orders {
		switch order.Size.Cmp(bestOrder.Size) {
		case 1: // order.Size > bestOrder.Size
			order.Size = order.Size.Sub(bestOrder.Size)
			l.TotalSize = l.TotalSize.Sub(bestOrder.Size)
			bestOrder.Size = decimal.NewFromFloat(0)
			countFilledOrders++
		case -1: // order.Size < bestOrder.Size
			bestOrder.Size = bestOrder.Size.Sub(order.Size)
			l.TotalSize = l.TotalSize.Sub(order.Size)
			order.Size = decimal.NewFromFloat(0)
		case 0: // order.Size == bestOrder.Size
			l.TotalSize = l.TotalSize.Sub(order.Size)
			order.Size = decimal.NewFromFloat(0)
			bestOrder.Size = decimal.NewFromFloat(0)
			countFilledOrders++
		}

		if order.Size.IsZero() {
			slog.Info("Order filled:", "price", l.Price)
			return true
		}
	}
	return false
}

func (l *Limit) removeOrder(orderID int) bool {
	for i, order := range l.Orders {
		if orderID == order.ID {
			l.TotalSize = l.TotalSize.Sub(order.Size)
			l.Orders = append(l.Orders[:i], l.Orders[i+1:]...)
			return true
		}
	}
	return false
}

func (l *Limit) removeEmptyOrders(countOrdersToDelete int) {
	l.Orders = l.Orders[countOrdersToDelete:]
}
