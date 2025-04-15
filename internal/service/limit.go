package service

import (
	"time"

	"github.com/shopspring/decimal"
)

type Limit struct {
	price  decimal.Decimal
	orders []*Order
	qty    decimal.Decimal
}

func newLimit(price decimal.Decimal) *Limit {
	return &Limit{
		price:  price,
		orders: []*Order{},
		qty:    decimal.Zero,
	}
}

func (l *Limit) matchOrders(order *Order) []Match {
	var (
		matches           = []Match{}
		countFilledOrders int
	)

	defer func() {
		if countFilledOrders > 0 {
			// delete filled orders
			l.orders = l.orders[countFilledOrders:]
		}
	}()

	for _, bestOrder := range l.orders {
		var match = Match{
			IsBid: order.isBid,
			Price: l.price,
		}

		switch {
		case order.qty.GreaterThan(bestOrder.qty):
			bestOrder.sizeFilled = bestOrder.sizeFilled.Add(bestOrder.qty)
			order.sizeFilled = order.sizeFilled.Add(bestOrder.qty)
			order.qty = order.qty.Sub(bestOrder.qty)
			l.qty = l.qty.Sub(bestOrder.qty)
			match.Qty = bestOrder.qty
			bestOrder.qty = decimal.Zero
			countFilledOrders++

		case order.qty.LessThan(bestOrder.qty):
			bestOrder.sizeFilled = bestOrder.sizeFilled.Add(order.qty)
			bestOrder.qty = bestOrder.qty.Sub(order.qty)
			order.sizeFilled = order.sizeFilled.Add(order.qty)
			l.qty = l.qty.Sub(order.qty)
			match.Qty = order.qty
			order.qty = decimal.Zero

		case order.qty.Equal(bestOrder.qty):
			bestOrder.sizeFilled = bestOrder.sizeFilled.Add(order.qty)
			order.sizeFilled = order.sizeFilled.Add(order.qty)
			l.qty = l.qty.Sub(order.qty)
			match.Qty = order.qty
			order.qty = decimal.Zero
			bestOrder.qty = decimal.Zero
			countFilledOrders++
		}

		match.Order = *bestOrder
		match.Time = time.Now()

		matches = append(matches, match)

		if order.qty.IsZero() {
			return matches
		}
	}
	return matches
}

func (l *Limit) removeOrder(orderID int) bool {
	for i, order := range l.orders {
		if orderID == order.ID {
			l.qty = l.qty.Sub(order.qty)
			l.orders = append(l.orders[:i], l.orders[i+1:]...)
			return true
		}
	}
	return false
}
