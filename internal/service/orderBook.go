package service

import (
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/BazaarTrade/OrderMatchingService/internal/models"
	"github.com/shopspring/decimal"
)

type OrderBook struct {
	pair string

	askMutex sync.RWMutex
	bidMutex sync.RWMutex

	bestBidLimits []*Limit
	bestAskLimits []*Limit

	bidLimits map[string]*Limit
	askLimits map[string]*Limit

	bidQty decimal.Decimal
	askQty decimal.Decimal

	logger *slog.Logger
}

func NewOrderBook(pair string, logger *slog.Logger) *OrderBook {
	return &OrderBook{
		pair:          pair,
		bestBidLimits: make([]*Limit, 0),
		bestAskLimits: make([]*Limit, 0),
		bidLimits:     make(map[string]*Limit),
		askLimits:     make(map[string]*Limit),
		logger:        logger,
	}
}

type Order struct {
	ID         int
	isBid      bool
	orderType  string
	price      decimal.Decimal
	qty        decimal.Decimal
	sizeFilled decimal.Decimal
}

type Match struct {
	Order Order
	Pair  string
	IsBid bool
	Price decimal.Decimal
	Qty   decimal.Decimal
	Time  time.Time
}

func (ob *OrderBook) placeLimitOrder(price string, order *Order) []Match {
	var (
		limit   *Limit
		matches []Match
	)

	switch {
	case order.isBid:
		ob.askMutex.Lock()
		if len(ob.bestAskLimits) > 0 && order.price.GreaterThanOrEqual(ob.bestAskLimits[0].price) { //if limit order can be filled or partialy filled instantly
			matches = ob.fillOrder(order)
		}
		ob.askMutex.Unlock()

		if order.qty.IsZero() {
			return matches
		}

		ob.bidMutex.Lock()
		defer ob.bidMutex.Unlock()

		if limit = ob.bidLimits[price]; limit == nil { //get or create limit if not exists
			limit = newLimit(order.price)
			ob.bidLimits[price] = limit
			ob.insertLimit(limit, order.isBid)
		}
		ob.bidQty = ob.bidQty.Add(order.qty)

	case !order.isBid:
		ob.bidMutex.Lock()
		if len(ob.bestBidLimits) > 0 && order.price.LessThanOrEqual(ob.bestBidLimits[0].price) { //if limit order can be filled or partialy filled instantly
			matches = ob.fillOrder(order)
		}
		ob.bidMutex.Unlock()

		if order.qty.IsZero() {
			return matches
		}

		ob.askMutex.Lock()
		defer ob.askMutex.Unlock()

		if limit = ob.askLimits[price]; limit == nil { //get or create limit if not exists
			limit = newLimit(order.price)
			ob.askLimits[price] = limit
			ob.insertLimit(limit, order.isBid)
		}
		ob.askQty = ob.askQty.Add(order.qty)
	}
	limit.orders = append(limit.orders, order)
	limit.qty = limit.qty.Add(order.qty)
	return matches
}

func (ob *OrderBook) placeMarketOrder(order *Order) []Match {
	switch {
	case order.isBid:
		ob.askMutex.Lock()
		defer ob.askMutex.Unlock()

	case !order.isBid:
		ob.bidMutex.Lock()
		defer ob.bidMutex.Unlock()
	}
	return ob.fillOrder(order)
}

func (ob *OrderBook) cancelLimitOrder(order models.Order) bool {
	switch {
	case order.IsBid:
		ob.bidMutex.Lock()
		defer ob.bidMutex.Unlock()

		limit, ok := ob.bidLimits[order.Price]
		if !ok {
			ob.logger.Error("failed to find limit", "price", order.Price)
			return false
		}

		if !limit.removeOrder(order.ID) {
			ob.logger.Error("failed to find order", "orderID", order.ID)
			return false
		}

		orderQtyDecimal, err := decimal.NewFromString(order.Qty)
		if err != nil {
			ob.logger.Error("failed to convert order qty to decimal", "orderID", order.ID)
			return false
		}

		ob.bidQty = ob.bidQty.Sub(orderQtyDecimal)

		if limit.qty.IsZero() {
			delete(ob.bidLimits, order.Price)
			for i, bestLimit := range ob.bestBidLimits {
				if bestLimit == limit {
					ob.bestBidLimits = append(ob.bestBidLimits[:i], ob.bestBidLimits[i+1:]...)
				}
			}
		}

	case !order.IsBid:
		ob.askMutex.Lock()
		defer ob.askMutex.Unlock()

		limit, ok := ob.askLimits[order.Price]
		if !ok {
			ob.logger.Error("failed to find limit", "price", order.Price)
			return false
		}

		if !limit.removeOrder(order.ID) {
			ob.logger.Error("failed to find order", "orderID", order.ID)
			return false
		}

		orderQtyDecimal, err := decimal.NewFromString(order.Qty)
		if err != nil {
			ob.logger.Error("failed to convert order qty to decimal", "orderID", order.ID)
			return false
		}

		ob.askQty = ob.askQty.Sub(orderQtyDecimal)

		if limit.qty.IsZero() {
			delete(ob.askLimits, order.Price)
			for i, bestLimit := range ob.bestAskLimits {
				if bestLimit == limit {
					ob.bestAskLimits = append(ob.bestAskLimits[:i], ob.bestAskLimits[i+1:]...)
				}
			}
		}
	}
	return true
}

func (ob *OrderBook) fillOrder(order *Order) []Match {
	var matches = []Match{}

	switch {
	case order.isBid:
		for _, bestAskLimit := range ob.bestAskLimits {
			if order.orderType == "limit" && order.price.LessThan(bestAskLimit.price) {
				break
			}

			matches = append(matches, bestAskLimit.matchOrders(order)...)

			if bestAskLimit.qty.IsZero() {
				delete(ob.askLimits, bestAskLimit.price.String())
			}

			if order.qty.IsZero() {
				break
			}
		}

		ob.askQty = ob.askQty.Sub(order.sizeFilled)

		if len(ob.bestAskLimits) != len(ob.askLimits) {
			ob.bestAskLimits = ob.bestAskLimits[len(ob.bestAskLimits)-len(ob.askLimits):]
		}

	case !order.isBid:
		for _, bestBidLimit := range ob.bestBidLimits {
			if order.orderType == "limit" && order.price.GreaterThan(bestBidLimit.price) {
				break
			}

			matches = append(matches, bestBidLimit.matchOrders(order)...)

			if bestBidLimit.qty.IsZero() {
				delete(ob.bidLimits, bestBidLimit.price.String())
			}

			if order.qty.IsZero() {
				break
			}
		}

		ob.bidQty = ob.bidQty.Sub(order.sizeFilled)

		if len(ob.bestBidLimits) != len(ob.bidLimits) {
			ob.bestBidLimits = ob.bestBidLimits[len(ob.bestBidLimits)-len(ob.bidLimits):]
		}
	}
	return matches
}

func (ob *OrderBook) insertLimit(limit *Limit, isBid bool) {
	switch {
	case isBid:
		pos := sort.Search(len(ob.bestBidLimits), func(i int) bool {
			return ob.bestBidLimits[i].price.LessThanOrEqual(limit.price)
		})
		ob.bestBidLimits = append(ob.bestBidLimits[:pos], append([]*Limit{limit}, ob.bestBidLimits[pos:]...)...)

	case !isBid:
		pos := sort.Search(len(ob.bestAskLimits), func(i int) bool {
			return ob.bestAskLimits[i].price.GreaterThanOrEqual(limit.price)
		})
		ob.bestAskLimits = append(ob.bestAskLimits[:pos], append([]*Limit{limit}, ob.bestAskLimits[pos:]...)...)
	}
}

func (ob *OrderBook) orderBookSnapshot() models.OrderBookSnapshot {
	ob.bidMutex.RLock()
	ob.askMutex.RLock()
	defer func() {
		ob.bidMutex.RUnlock()
		ob.askMutex.RUnlock()
	}()

	var (
		OrderBookSnapshot = models.OrderBookSnapshot{
			Pair:    ob.pair,
			Bids:    make([]models.Limit, len(ob.bestBidLimits)),
			Asks:    make([]models.Limit, len(ob.bestAskLimits)),
			BidsQty: ob.bidQty.String(),
			AsksQty: ob.askQty.String(),
		}
	)

	for i, limit := range ob.bestBidLimits {
		OrderBookSnapshot.Bids[i] = models.Limit{
			Price: limit.price.String(),
			Qty:   limit.qty.String(),
		}
	}

	for i, limit := range ob.bestAskLimits {
		OrderBookSnapshot.Asks[i] = models.Limit{
			Price: limit.price.String(),
			Qty:   limit.qty.String(),
		}
	}
	return OrderBookSnapshot
}

func (ob *OrderBook) isEnoughQty(isBid bool, qty decimal.Decimal) bool {
	return isBid && ob.askQty.GreaterThanOrEqual(qty) || !isBid && ob.bidQty.GreaterThanOrEqual(qty)
}
