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

func (l *Limit) matchOrders(order *Order, matches *[]Match) bool {
	var countFilledOrders int

	defer func() {
		if countFilledOrders > 0 {
			l.removeEmptyOrders(countFilledOrders)
		}
	}()

	for _, bestOrder := range l.Orders {
		var match = Match{
			Price:          l.Price,
			CounterOrderID: bestOrder.ID,
		}

		switch order.Qty.Cmp(bestOrder.Qty) {
		case 1: // order.Qty > bestOrder.Qty
			bestOrder.SizeFilled = bestOrder.SizeFilled.Add(bestOrder.Qty)
			order.Qty = order.Qty.Sub(bestOrder.Qty)
			order.SizeFilled = order.SizeFilled.Add(bestOrder.Qty)
			l.TotalSize = l.TotalSize.Sub(bestOrder.Qty)
			match.Qty = bestOrder.Qty
			bestOrder.Qty = decimal.NewFromFloat(0)
			countFilledOrders++
		case -1: // order.Qty < bestOrder.Qty
			bestOrder.SizeFilled = bestOrder.SizeFilled.Add(order.Qty)
			bestOrder.Qty = bestOrder.Qty.Sub(order.Qty)
			order.SizeFilled = order.SizeFilled.Add(order.Qty)
			l.TotalSize = l.TotalSize.Sub(order.Qty)
			match.Qty = order.Qty
			order.Qty = decimal.NewFromFloat(0)
		case 0: // order.Qty == bestOrder.Qty
			bestOrder.SizeFilled = bestOrder.SizeFilled.Add(order.Qty)
			order.SizeFilled = order.SizeFilled.Add(order.Qty)
			l.TotalSize = l.TotalSize.Sub(order.Qty)
			match.Qty = order.Qty
			order.Qty = decimal.NewFromFloat(0)
			bestOrder.Qty = decimal.NewFromFloat(0)
			countFilledOrders++
		}

		match.CounterOrderSizeFilled = bestOrder.SizeFilled
		*matches = append(*matches, match)

		if order.Qty.IsZero() {
			slog.Info("Order filled:", "price", l.Price)
			return true
		}
	}
	return false
}

func (l *Limit) removeOrder(orderID int64) bool {
	for i, order := range l.Orders {
		if orderID == order.ID {
			l.TotalSize = l.TotalSize.Sub(order.Qty)
			l.Orders = append(l.Orders[:i], l.Orders[i+1:]...)
			return true
		}
	}
	return false
}

func (l *Limit) removeEmptyOrders(countOrdersToDelete int) {
	l.Orders = l.Orders[countOrdersToDelete:]
}
