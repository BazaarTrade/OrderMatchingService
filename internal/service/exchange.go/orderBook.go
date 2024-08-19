package exchange

import (
	"errors"
	"log/slog"
	"sort"
	"sync"

	"github.com/shopspring/decimal"
)

type OrderBook struct {
	askMutex sync.RWMutex
	bidMutex sync.RWMutex

	BestBidLimits []*Limit
	BestAskLimits []*Limit

	BidLimits map[string]*Limit
	AskLimits map[string]*Limit

	BidVolume decimal.Decimal
	AskVolume decimal.Decimal
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		BestBidLimits: make([]*Limit, 0),
		BestAskLimits: make([]*Limit, 0),
		BidLimits:     make(map[string]*Limit),
		AskLimits:     make(map[string]*Limit),
	}
}

type Order struct {
	ID         int64
	IsBid      bool
	Type       string
	Price      decimal.Decimal
	Qty        decimal.Decimal
	SizeFilled decimal.Decimal
}

type Match struct {
	Qty                    decimal.Decimal
	Price                  decimal.Decimal
	CounterOrderID         int64
	CounterOrderStatus     string
	CounterOrderSizeFilled decimal.Decimal
}

func (ob *OrderBook) placeLimitOrder(price string, order *Order) (*[]Match, error) {
	slog.Info("Placing limit order:", "orderID", order.ID, "isBid", order.IsBid, "size", order.Qty, "price", price)

	var (
		limit   *Limit
		matches *[]Match
	)

	switch {
	case order.IsBid:
		ob.askMutex.RLock()
		if len(ob.BestAskLimits) > 0 && order.Price.Cmp(ob.BestAskLimits[0].Price) >= 0 { //if limit order can be filled or partialy filled instantly
			ob.askMutex.RUnlock()
			matches = ob.fillOrder(order)
			if order.Qty.IsZero() {
				return matches, nil
			}
		} else {
			ob.askMutex.RUnlock()
		}

		ob.bidMutex.Lock()
		defer ob.bidMutex.Unlock()

		if limit = ob.BidLimits[price]; limit == nil { //get or create limit if not exists
			slog.Info("New limt:", "price", order.Price)

			limit = NewLimit(order.Price)
			ob.BidLimits[price] = limit
			ob.BestBidLimits = append(ob.BestBidLimits, limit)
			ob.sortBestLimits(order.IsBid)
		}

	case !order.IsBid:
		ob.bidMutex.RLock()
		if len(ob.BestBidLimits) > 0 && order.Price.Cmp(ob.BestBidLimits[0].Price) <= 0 { //if limit order can be filled or partialy filled instantly
			ob.bidMutex.RUnlock()
			matches = ob.fillOrder(order)
			if order.Qty.IsZero() {
				return matches, nil
			}
		} else {
			ob.bidMutex.RUnlock()
		}

		ob.askMutex.Lock()
		defer ob.askMutex.Unlock()
		if limit = ob.AskLimits[price]; limit == nil { //get or create limit if not exists
			slog.Info("New limt:", "price", order.Price)

			limit = NewLimit(order.Price)
			ob.AskLimits[price] = limit
			ob.BestAskLimits = append(ob.BestAskLimits, limit)
			ob.sortBestLimits(order.IsBid)
		}
	}

	limit.Orders = append(limit.Orders, order)
	limit.TotalSize = limit.TotalSize.Add(order.Qty)
	slog.Info("Limit order placed:", "totalSize", limit.TotalSize)
	return matches, nil
}

func (ob *OrderBook) placeMarketOrder(order *Order) (*[]Match, error) {
	slog.Info("Placing market order:", "orderID", order.ID, "isBid", order.IsBid, "size", order.Qty)

	matches := ob.fillOrder(order)
	if !order.Qty.IsZero() {
		slog.Info("Not enough volume")
		return nil, errors.New("not enough volume")
	}

	return matches, nil
}

func (ob *OrderBook) cancelLimitOrder(orderID int64, orderPrice string, isBid bool) error {
	slog.Info("Cancelling limit order", "orderID", orderID)

	switch {
	case isBid:
		ob.bidMutex.Lock()
		defer ob.bidMutex.Unlock()
		limit, ok := ob.BidLimits[orderPrice]
		if !ok {
			return errors.New("limit not found")
		}

		if !limit.removeOrder(orderID) {
			return errors.New("order not found")
		}

	case !isBid:
		ob.askMutex.Lock()
		defer ob.askMutex.Unlock()
		limit, ok := ob.AskLimits[orderPrice]
		if !ok {
			return errors.New("limit not found")
		}

		if !limit.removeOrder(orderID) {
			return errors.New("order not found")
		}
	}
	return nil
}

func (ob *OrderBook) fillOrder(order *Order) *[]Match {
	var (
		emptyLimits []string
		matches     = &[]Match{}
	)

	switch {
	case order.IsBid:
		ob.askMutex.Lock()

		defer func() {
			if emptyLimits != nil {
				removeEmptyLimits(emptyLimits, &ob.BestAskLimits, ob.AskLimits)
			}
			ob.askMutex.Unlock()
		}()

		for _, bestAskLimit := range ob.BestAskLimits {
			if order.Type == "Limit" && order.Price.Cmp(bestAskLimit.Price) < 0 {
				return matches
			}

			if bestAskLimit.matchOrders(order, matches) {
				if bestAskLimit.TotalSize.IsZero() {
					emptyLimitPriceString := bestAskLimit.Price.String()
					emptyLimits = append(emptyLimits, emptyLimitPriceString)
				}
				return matches
			}

			emptyLimitPriceString := bestAskLimit.Price.String()
			emptyLimits = append(emptyLimits, emptyLimitPriceString)
		}

	case !order.IsBid:
		ob.bidMutex.Lock()
		defer func() {
			if emptyLimits != nil {
				removeEmptyLimits(emptyLimits, &ob.BestBidLimits, ob.BidLimits)
			}
			ob.bidMutex.Unlock()
		}()

		for _, bestBidLimit := range ob.BestBidLimits {
			if order.Type == "Limit" && order.Price.Cmp(bestBidLimit.Price) > 0 {
				return matches
			}

			if bestBidLimit.matchOrders(order, matches) {
				if bestBidLimit.TotalSize.IsZero() {
					emptyLimitPriceString := bestBidLimit.Price.String()
					emptyLimits = append(emptyLimits, emptyLimitPriceString)
				}
				return matches
			}

			emptyLimitPriceString := bestBidLimit.Price.String()
			emptyLimits = append(emptyLimits, emptyLimitPriceString)
		}
	}
	return matches
}

func (ob *OrderBook) sortBestLimits(isBid bool) {
	switch {
	case isBid:
		sort.Slice(ob.BestBidLimits, func(i, j int) bool {
			return ob.BestBidLimits[i].Price.Cmp(ob.BestBidLimits[j].Price) > 0
		})

	case !isBid:
		sort.Slice(ob.BestAskLimits, func(i, j int) bool {
			return ob.BestAskLimits[i].Price.Cmp(ob.BestAskLimits[j].Price) < 0
		})
	}
}

func removeEmptyLimits(emptyLimits []string, bestLimits *[]*Limit, limits map[string]*Limit) {
	slog.Info("Removing empty limits", "prices", emptyLimits)
	for _, limitPrice := range emptyLimits {
		delete(limits, limitPrice)
	}

	*bestLimits = (*bestLimits)[len(emptyLimits):]
}
