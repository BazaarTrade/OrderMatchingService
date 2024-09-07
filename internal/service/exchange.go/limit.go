package exchange

import (
	"github.com/shopspring/decimal"
)

type Limit struct {
	price     decimal.Decimal
	orders    []*Order
	totalSize decimal.Decimal
}

func NewLimit(price decimal.Decimal) *Limit {
	return &Limit{
		price:     price,
		orders:    []*Order{},
		totalSize: decimal.NewFromFloat(0),
	}
}

func (l *Limit) matchOrders(order *Order, matches *[]Match) bool {
	var countFilledOrders int

	defer func() {
		if countFilledOrders > 0 {
			l.removeEmptyOrders(countFilledOrders)
		}
	}()

	for _, bestOrder := range l.orders {
		var match = Match{
			price:          l.price,
			counterOrderID: bestOrder.ID,
		}

		switch order.qty.Cmp(bestOrder.qty) {
		case 1: // order.qty > bestOrder.qty
			bestOrder.sizeFilled = bestOrder.sizeFilled.Add(bestOrder.qty)
			order.qty = order.qty.Sub(bestOrder.qty)
			order.sizeFilled = order.sizeFilled.Add(bestOrder.qty)
			l.totalSize = l.totalSize.Sub(bestOrder.qty)
			match.qty = bestOrder.qty
			bestOrder.qty = decimal.NewFromFloat(0)
			countFilledOrders++
		case -1: // order.qty < bestOrder.qty
			bestOrder.sizeFilled = bestOrder.sizeFilled.Add(order.qty)
			bestOrder.qty = bestOrder.qty.Sub(order.qty)
			order.sizeFilled = order.sizeFilled.Add(order.qty)
			l.totalSize = l.totalSize.Sub(order.qty)
			match.qty = order.qty
			order.qty = decimal.NewFromFloat(0)
		case 0: // order.qty == bestOrder.qty
			bestOrder.sizeFilled = bestOrder.sizeFilled.Add(order.qty)
			order.sizeFilled = order.sizeFilled.Add(order.qty)
			l.totalSize = l.totalSize.Sub(order.qty)
			match.qty = order.qty
			order.qty = decimal.NewFromFloat(0)
			bestOrder.qty = decimal.NewFromFloat(0)
			countFilledOrders++
		}

		match.counterOrderSizeFilled = bestOrder.sizeFilled
		*matches = append(*matches, match)

		if order.qty.IsZero() {
			return true
		}
	}
	return false
}

func (l *Limit) removeOrder(orderID int64) bool {
	for i, order := range l.orders {
		if orderID == order.ID {
			l.totalSize = l.totalSize.Sub(order.qty)
			l.orders = append(l.orders[:i], l.orders[i+1:]...)
			return true
		}
	}
	return false
}

func (l *Limit) removeEmptyOrders(countOrdersToDelete int) {
	l.orders = l.orders[countOrdersToDelete:]
}
