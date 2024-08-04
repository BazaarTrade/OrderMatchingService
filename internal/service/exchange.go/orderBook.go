package exchange

import (
	"errors"
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/Moha192/OrderMatchingService/internal/repository"
	"github.com/shopspring/decimal"
)

type Order struct {
	ID        int
	UserID    int
	IsBid     bool
	Type      string
	Price     decimal.Decimal
	Size      decimal.Decimal
	Timestamp time.Time
}

func NewOrder(userID int, isBid bool, size string) *Order {
	sizeDecimal, err := decimal.NewFromString(size)
	if err != nil {
		slog.Error("Error parsing size:", "error", err)
		return nil
	}

	return &Order{
		UserID:    userID,
		IsBid:     isBid,
		Size:      sizeDecimal,
		Timestamp: time.Now(),
	}
}

type OrderBook struct {
	db *repository.Database

	askMutex sync.RWMutex
	bidMutex sync.RWMutex

	BestBidLimits []*Limit
	BestAskLimits []*Limit

	BidLimits map[string]*Limit
	AskLimits map[string]*Limit
}

func NewOrderBook(db *repository.Database) *OrderBook {
	return &OrderBook{
		db:            db,
		BestBidLimits: make([]*Limit, 0),
		BestAskLimits: make([]*Limit, 0),
		BidLimits:     make(map[string]*Limit),
		AskLimits:     make(map[string]*Limit),
	}
}

func (ob *OrderBook) PlaceLimitOrder(price string, order *Order) {
	slog.Info("Placing limit order:", "orderID", order.ID, "userID", order.UserID, "isBid", order.IsBid, "size", order.Size, "price", price)

	var (
		err   error
		limit *Limit
	)
	order.Type = "Limit"
	order.Price, err = decimal.NewFromString(price)
	if err != nil {
		slog.Error("Error parsing size:", "error", err)
	}

	switch {
	case order.IsBid:
		ob.askMutex.RLock()
		if len(ob.BestAskLimits) > 0 && order.Price.Cmp(ob.BestAskLimits[0].Price) >= 0 { //if limit order can be filled or partialy filled instantly
			ob.askMutex.RUnlock()
			if ob.fillOrder(order) {
				return
			}
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
			if ob.fillOrder(order) {
				return
			}
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
	limit.TotalSize = limit.TotalSize.Add(order.Size)
	slog.Info("Limit order placed:", "totalSize", limit.TotalSize)
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

func (ob *OrderBook) CancelLimitOrder(orderID int, orderPrice string, isBid bool) error {
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

func (ob *OrderBook) PlaceMarketOrder(order *Order) {
	slog.Info("Placing market order:", "orderID", order.ID, "userID", order.UserID, "isBid", order.IsBid, "size", order.Size)

	order.Type = "Market"
	if !ob.fillOrder(order) {
		slog.Info("Not enough volume")
	}
}

func (ob *OrderBook) fillOrder(order *Order) bool {
	var emptyLimits []string

	switch {
	case order.IsBid:
		ob.askMutex.Lock()
		defer ob.askMutex.Unlock()

		defer func() {
			if emptyLimits != nil {
				ob.removeEmptyLimits(emptyLimits, order.IsBid)
			}
		}()

		for _, bestAskLimit := range ob.BestAskLimits {
			if order.Type == "Limit" && order.Price.Cmp(bestAskLimit.Price) < 0 {
				return false
			}

			if bestAskLimit.matchOrders(order) {
				if bestAskLimit.TotalSize.IsZero() {
					emptyLimitPriceString := bestAskLimit.Price.String()
					emptyLimits = append(emptyLimits, emptyLimitPriceString)
				}
				return true
			}

			emptyLimitPriceString := bestAskLimit.Price.String()
			emptyLimits = append(emptyLimits, emptyLimitPriceString)
		}

	case !order.IsBid:
		ob.bidMutex.Lock()
		defer ob.bidMutex.Unlock()

		defer func() {
			if emptyLimits != nil {
				ob.removeEmptyLimits(emptyLimits, order.IsBid)
			}
		}()

		for _, bestBidLimit := range ob.BestBidLimits {
			if order.Type == "Limit" && order.Price.Cmp(bestBidLimit.Price) > 0 {
				return false
			}

			if bestBidLimit.matchOrders(order) {
				if bestBidLimit.TotalSize.IsZero() {
					emptyLimitPriceString := bestBidLimit.Price.String()
					emptyLimits = append(emptyLimits, emptyLimitPriceString)
				}
				return true
			}

			emptyLimitPriceString := bestBidLimit.Price.String()
			emptyLimits = append(emptyLimits, emptyLimitPriceString)
		}
	}

	return false
}

func (ob *OrderBook) removeEmptyLimits(emptyLimits []string, isBid bool) {
	slog.Info("Removing empty limits", "prices", emptyLimits)
	var (
		lenLimits  int
		limitPrice string
	)

	switch {
	case isBid:
		for lenLimits, limitPrice = range emptyLimits {
			delete(ob.AskLimits, limitPrice)
		}
		ob.BestAskLimits = ob.BestAskLimits[lenLimits+1:]

	case !isBid:
		for lenLimits, limitPrice = range emptyLimits {
			delete(ob.BidLimits, limitPrice)
		}
		ob.BestBidLimits = ob.BestBidLimits[lenLimits+1:]

	}
}
