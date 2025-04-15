package service

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/BazaarTrade/OrderMatchingService/internal/models"
	"github.com/BazaarTrade/OrderMatchingService/internal/repository"
	"github.com/shopspring/decimal"
)

type Service struct {
	orderBooks map[string]*OrderBook
	mu         sync.RWMutex
	db         repository.Repository
	logger     *slog.Logger
}

func New(db repository.Repository, logger *slog.Logger) *Service {
	return &Service{
		orderBooks: make(map[string]*OrderBook),
		db:         db,
		logger:     logger,
	}
}

func (s *Service) CreateOrderBook(pair string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.orderBooks[pair]; exists {
		s.logger.Error("order book already exists")
		return errors.New("order book already exists")
	}

	s.orderBooks[pair] = NewOrderBook(pair, s.logger)
	return nil
}

func (s *Service) DeleteOrderBook(pair string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.orderBooks[pair]; !exists {
		s.logger.Error("failed to find order book")
		return errors.New("failed to find order book")
	}

	delete(s.orderBooks, pair)

	return nil
}

func (s *Service) PlaceOrder(placeOrderReq models.PlaceOrderReq) (models.Order, []models.Order, []Match, models.OrderBookSnapshot, error) {
	s.mu.RLock()
	ob, exists := s.orderBooks[placeOrderReq.Pair]
	s.mu.RUnlock()
	if !exists {
		s.logger.Error("failed to find order book", "Pair", placeOrderReq.Pair)
		return models.Order{}, nil, nil, models.OrderBookSnapshot{}, errors.New("failed to find order book")
	}

	qty := decimal.RequireFromString(placeOrderReq.Qty)

	if placeOrderReq.Type == "market" && !ob.isEnoughQty(placeOrderReq.IsBid, qty) {
		switch {
		case placeOrderReq.IsBid:
			s.logger.Error("not enough ask qty", "askQty", ob.askQty, "required Qty", qty)
			return models.Order{}, nil, nil, models.OrderBookSnapshot{}, errors.New("not enough ask qty")

		case !placeOrderReq.IsBid:
			s.logger.Error("not enough bid qty", "bidQty", ob.bidQty, "required Qty", qty)
			return models.Order{}, nil, nil, models.OrderBookSnapshot{}, errors.New("not enough bid qty")
		}
	}

	orderID, err := s.db.CreateOrder(placeOrderReq)
	if err != nil {
		s.logger.Error("failed to create order", "userID", placeOrderReq.UserID, "pair", placeOrderReq.Pair, "error", err)
		return models.Order{}, nil, nil, models.OrderBookSnapshot{}, err
	}

	defer func() {
		if err != nil {
			go s.db.SetOrderStatusToError(orderID)
		}
	}()

	var (
		matches    []Match
		placeOrder = &Order{
			ID:        orderID,
			isBid:     placeOrderReq.IsBid,
			orderType: placeOrderReq.Type,
			qty:       qty,
		}
	)

	switch placeOrder.orderType {
	case "limit":
		placeOrder.price = decimal.RequireFromString(placeOrderReq.Price)
		matches = ob.placeLimitOrder(placeOrderReq.Price, placeOrder)

	case "market":
		matches = ob.placeMarketOrder(placeOrder)
		if len(matches) < 1 {
			s.logger.Error("no matches found for order", "orderID", orderID)
			return models.Order{}, nil, nil, models.OrderBookSnapshot{}, errors.New("no matches found for order")
		}
	}

	order, err := s.db.UpdateOrderSizeFilled(orderID, placeOrder.sizeFilled.String())
	if err != nil {
		return models.Order{}, nil, nil, models.OrderBookSnapshot{}, err
	}

	var matchOrders = make([]models.Order, len(matches))
	if len(matches) > 0 {
		err = s.db.UpdateOrderPrice(orderID, avgPrice(matches).String())
		if err != nil {
			return models.Order{}, nil, nil, models.OrderBookSnapshot{}, err
		}

		for i, match := range matches {
			matchOrders[i], err = s.db.UpdateOrderSizeFilled(match.Order.ID, match.Order.sizeFilled.String())
			if err != nil {
				return models.Order{}, nil, nil, models.OrderBookSnapshot{}, err
			}

			err = s.db.AddMatch(orderID, models.Match{
				OrderID: match.Order.ID,
				Qty:     match.Qty.String(),
				Price:   match.Order.price.String(),
			})
			if err != nil {
				return models.Order{}, nil, nil, models.OrderBookSnapshot{}, err
			}

			matches[i].Pair = placeOrderReq.Pair
		}
	}

	return order, matchOrders, matches, ob.orderBookSnapshot(), nil
}

func (s *Service) CancelOrder(orderID int) (models.Order, models.OrderBookSnapshot, error) {
	order, err := s.db.GetOrderByOrderID(orderID)
	if err != nil {
		return models.Order{}, models.OrderBookSnapshot{}, err
	}

	ob, exists := s.orderBooks[order.Pair]
	if !exists {
		s.logger.Error("failed to find order book")
		return models.Order{}, models.OrderBookSnapshot{}, errors.New("failed to find order book")
	}

	if !ob.cancelLimitOrder(order) {
		return models.Order{}, models.OrderBookSnapshot{}, errors.New("failed to cancel order")
	}

	if err = s.db.SetOrderStatusToCancel(orderID); err != nil {
		return models.Order{}, models.OrderBookSnapshot{}, err
	}

	order, err = s.db.GetOrderByOrderID(orderID)
	if err != nil {
		return models.Order{}, models.OrderBookSnapshot{}, err
	}

	return order, ob.orderBookSnapshot(), nil
}

func (s *Service) GetPairs() ([]string, error) {
	return s.db.GetPairs()
}

func avgPrice(matches []Match) decimal.Decimal {
	var (
		totalValue decimal.Decimal
		totalQty   decimal.Decimal
	)

	for _, match := range matches {
		totalValue = totalValue.Add(match.Order.price.Mul(match.Qty))
		totalQty = totalQty.Add(match.Qty)
	}

	return totalValue.Div(totalQty)
}
