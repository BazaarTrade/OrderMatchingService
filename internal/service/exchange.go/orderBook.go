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

	bestBidLimits []*Limit
	bestAskLimits []*Limit

	bidLimits map[string]*Limit
	askLimits map[string]*Limit

	bidVolume decimal.Decimal
	askVolume decimal.Decimal

	logger *slog.Logger
}

func NewOrderBook(logger *slog.Logger) *OrderBook {
	return &OrderBook{
		bestBidLimits: make([]*Limit, 0),
		bestAskLimits: make([]*Limit, 0),
		bidLimits:     make(map[string]*Limit),
		askLimits:     make(map[string]*Limit),
		logger:        logger,
	}
}

type Order struct {
	ID         int64
	isBid      bool
	orderType  string
	price      decimal.Decimal
	qty        decimal.Decimal
	sizeFilled decimal.Decimal
}

type Match struct {
	qty                    decimal.Decimal
	price                  decimal.Decimal
	counterOrderID         int64
	counterOrderSizeFilled decimal.Decimal
}

func (ob *OrderBook) placeLimitOrder(price string, order *Order) (*[]Match, error) {
	var (
		limit   *Limit
		matches *[]Match
	)

	switch {
	case order.isBid:
		ob.askMutex.RLock()
		if len(ob.bestAskLimits) > 0 && order.price.Cmp(ob.bestAskLimits[0].price) >= 0 { //if limit order can be filled or partialy filled instantly
			ob.askMutex.RUnlock()
			matches = ob.fillOrder(order)
			if order.qty.IsZero() {
				return matches, nil
			}
		} else {
			ob.askMutex.RUnlock()
		}

		ob.bidMutex.Lock()
		defer ob.bidMutex.Unlock()

		if limit = ob.bidLimits[price]; limit == nil { //get or create limit if not exists
			limit = NewLimit(order.price)
			ob.bidLimits[price] = limit
			ob.bestBidLimits = append(ob.bestBidLimits, limit)
			ob.sortBestLimits(order.isBid)
		}

		ob.bidVolume = ob.bidVolume.Add(order.qty)

	case !order.isBid:
		ob.bidMutex.RLock()
		if len(ob.bestBidLimits) > 0 && order.price.Cmp(ob.bestBidLimits[0].price) <= 0 { //if limit order can be filled or partialy filled instantly
			ob.bidMutex.RUnlock()
			matches = ob.fillOrder(order)
			if order.qty.IsZero() {
				return matches, nil
			}
		} else {
			ob.bidMutex.RUnlock()
		}

		ob.askMutex.Lock()
		defer ob.askMutex.Unlock()

		if limit = ob.askLimits[price]; limit == nil { //get or create limit if not exists
			limit = NewLimit(order.price)
			ob.askLimits[price] = limit
			ob.bestAskLimits = append(ob.bestAskLimits, limit)
			ob.sortBestLimits(order.isBid)
		}

		ob.askVolume = ob.askVolume.Add(order.qty)
	}

	limit.orders = append(limit.orders, order)
	limit.totalSize = limit.totalSize.Add(order.qty)
	return matches, nil
}

func (ob *OrderBook) placeMarketOrder(order *Order) (*[]Match, error) {
	matches := ob.fillOrder(order)
	if !order.qty.IsZero() {
		return nil, errors.New("not enough volume")
	}

	return matches, nil
}

func (ob *OrderBook) cancelLimitOrder(orderID int64, orderPrice string, isBid bool) error {
	switch {
	case isBid:
		ob.bidMutex.Lock()
		defer ob.bidMutex.Unlock()
		limit, ok := ob.bidLimits[orderPrice]
		if !ok {
			return errors.New("limit not found")
		}

		if !limit.removeOrder(orderID) {
			return errors.New("order not found")
		}

		if limit.totalSize.IsZero() {
			removeEmptyLimits([]string{orderPrice}, &ob.bestBidLimits, ob.bidLimits)
		}

	case !isBid:
		ob.askMutex.Lock()
		defer ob.askMutex.Unlock()
		limit, ok := ob.askLimits[orderPrice]
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
	case order.isBid:
		ob.askMutex.Lock()

		defer func() {
			if emptyLimits != nil {
				removeEmptyLimits(emptyLimits, &ob.bestAskLimits, ob.askLimits)
			}
			ob.askMutex.Unlock()
		}()

		for _, bestAskLimit := range ob.bestAskLimits {
			if order.orderType == "limit" && order.price.Cmp(bestAskLimit.price) < 0 {
				return matches
			}

			if bestAskLimit.matchOrders(order, matches) {
				if bestAskLimit.totalSize.IsZero() {
					emptyLimits = append(emptyLimits, bestAskLimit.price.String())
				}
				return matches
			}

			emptyLimitPriceString := bestAskLimit.price.String()
			emptyLimits = append(emptyLimits, emptyLimitPriceString)
		}

	case !order.isBid:
		ob.bidMutex.Lock()
		defer func() {
			if emptyLimits != nil {
				removeEmptyLimits(emptyLimits, &ob.bestBidLimits, ob.bidLimits)
			}
			ob.bidMutex.Unlock()
		}()

		for _, bestBidLimit := range ob.bestBidLimits {
			if order.orderType == "limit" && order.price.Cmp(bestBidLimit.price) > 0 {
				return matches
			}

			if bestBidLimit.matchOrders(order, matches) {
				if bestBidLimit.totalSize.IsZero() {
					emptyLimits = append(emptyLimits, bestBidLimit.price.String())
				}
				return matches
			}

			emptyLimitPriceString := bestBidLimit.price.String()
			emptyLimits = append(emptyLimits, emptyLimitPriceString)
		}
	}
	return matches
}

func (ob *OrderBook) sortBestLimits(isBid bool) {
	switch {
	case isBid:
		sort.Slice(ob.bestBidLimits, func(i, j int) bool {
			return ob.bestBidLimits[i].price.Cmp(ob.bestBidLimits[j].price) > 0
		})

	case !isBid:
		sort.Slice(ob.bestAskLimits, func(i, j int) bool {
			return ob.bestAskLimits[i].price.Cmp(ob.bestAskLimits[j].price) < 0
		})
	}
}

func removeEmptyLimits(emptyLimits []string, bestLimits *[]*Limit, limits map[string]*Limit) {
	for _, limitPrice := range emptyLimits {
		delete(limits, limitPrice)
	}

	*bestLimits = (*bestLimits)[len(emptyLimits):]
}
