package exchange

import (
	"errors"
	"log/slog"
	"sync"

	"github.com/BazaarTrade/OrderMatchingService/internal/models"
	"github.com/BazaarTrade/OrderMatchingService/internal/repository"
	"github.com/shopspring/decimal"
)

type Service struct {
	db         repository.Repository
	orderBooks map[string]*OrderBook
	mu         sync.RWMutex
	logger     *slog.Logger
}

func New(db repository.Repository, logger *slog.Logger) *Service {
	return &Service{
		db:         db,
		orderBooks: make(map[string]*OrderBook),
		logger:     logger,
	}
}

func (s *Service) CreateOrderBook(pair string, pricePrecisions []int32, qtyPrecision int32) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.orderBooks[pair]; exists {
		s.logger.Error("order book already exists")
		return errors.New("order book already exists")
	}

	err := s.db.CreatePair(pair, pricePrecisions, qtyPrecision)
	if err != nil {
		return err
	}

	s.orderBooks[pair] = NewOrderBook(pair, s.logger)
	s.logger.Info("OrderBook created successfully", "pair", pair)
	return nil
}

func (s *Service) AddOrderBook(pair string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.orderBooks[pair]; exists {
		s.logger.Error("order book already exists")
		return errors.New("order book already exists")
	}

	s.orderBooks[pair] = NewOrderBook(pair, s.logger)
	s.logger.Info("OrderBook added successfully", "pair", pair)
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

	s.logger.Info("Order book deleted successfully", "pair", pair)
	return nil
}

func (s *Service) PlaceOrder(placeOrderReq models.PlaceOrderReq) (models.Order, []models.Order, models.OrderBookSnapshot, error) {
	ob, exists := s.orderBooks[placeOrderReq.Pair]
	if !exists {
		s.logger.Error("failed to find order book", "Pair", placeOrderReq.Pair)
		return models.Order{}, nil, models.OrderBookSnapshot{}, errors.New("failed to find order book")
	}

	qtyDecimal, err := decimal.NewFromString(placeOrderReq.Qty)
	if err != nil {
		s.logger.Error("failed to convert qty to decimal", "qty", placeOrderReq.Qty, "error", err)
		return models.Order{}, nil, models.OrderBookSnapshot{}, err
	} else if qtyDecimal.IsNegative() {
		s.logger.Error("invalid qty value", "qty", placeOrderReq.Qty)
		return models.Order{}, nil, models.OrderBookSnapshot{}, errors.New("invalid qty value")
	}

	if placeOrderReq.Type == "market" && !ob.isEnoughQty(placeOrderReq.IsBid, qtyDecimal) {
		switch {
		case placeOrderReq.IsBid:
			s.logger.Error("not enough ask qty", "askQty", ob.askQty, "required Qty", qtyDecimal)
			return models.Order{}, nil, models.OrderBookSnapshot{}, errors.New("not enough ask qty")

		case !placeOrderReq.IsBid:
			s.logger.Error("not enough bid qty", "bidQty", ob.bidQty, "required Qty", qtyDecimal)
			return models.Order{}, nil, models.OrderBookSnapshot{}, errors.New("not enough bid qty")
		}
	}

	orderID, err := s.db.CreateOrder(placeOrderReq)
	if err != nil {
		s.logger.Error("failed to create order", "userID", placeOrderReq.UserID, "pair", placeOrderReq.Pair, "error", err)
		return models.Order{}, nil, models.OrderBookSnapshot{}, err
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
			qty:       qtyDecimal,
		}
		orderSizeFilled decimal.Decimal
	)

	switch placeOrder.orderType {
	case "limit":
		placeOrder.price, err = decimal.NewFromString(placeOrderReq.Price)
		if err != nil {
			s.logger.Error("failed to convert price to decimal", "price", placeOrderReq.Price, "error", err)
			return models.Order{}, nil, models.OrderBookSnapshot{}, err

		} else if placeOrder.price.LessThanOrEqual(decimal.Zero) {
			s.logger.Error("invalid price value", "qty", placeOrderReq.Qty)
			return models.Order{}, nil, models.OrderBookSnapshot{}, errors.New("invalid price value")
		}

		s.logger.Info(
			"Placing limit Order",
			"userID", placeOrderReq.UserID,
			"orderID", orderID,
			"Pair", placeOrderReq.Pair,
			"isBid", placeOrderReq.IsBid,
			"price", placeOrderReq.Price,
			"qty", placeOrderReq.Qty,
		)

		matches, orderSizeFilled, err = ob.placeLimitOrder(placeOrderReq.Price, placeOrder)
		if err != nil {
			s.logger.Error("failed to place limit order", "error", err)
			return models.Order{}, nil, models.OrderBookSnapshot{}, err
		}

	case "market":
		s.logger.Info(
			"Placing market Order",
			"userID", placeOrderReq.UserID,
			"orderID", orderID,
			"Pair", placeOrderReq.Pair,
			"isBid", placeOrderReq.IsBid,
			"qty", placeOrderReq.Qty,
		)

		matches, err = ob.placeMarketOrder(placeOrder)
		if err != nil {
			s.logger.Error("failed to place market order", "error", err)
			return models.Order{}, nil, models.OrderBookSnapshot{}, err
		}

		if len(matches) < 1 {
			s.logger.Warn("no matches found for order", "orderID", orderID)
			return models.Order{}, nil, models.OrderBookSnapshot{}, errors.New("no matches found for order")
		}

		orderSizeFilled = placeOrder.sizeFilled
	}

	var matchOrders = make([]models.Order, len(matches))
	if len(matches) > 0 {
		err = s.db.UpdateOrderPrice(orderID, avgPrice(matches).String())
		if err != nil {
			return models.Order{}, nil, models.OrderBookSnapshot{}, err
		}

		for i, match := range matches {
			matchOrders[i], err = s.db.UpdateOrderSizeFilled(match.Order.ID, match.Order.sizeFilled.String())
			if err != nil {
				return models.Order{}, nil, models.OrderBookSnapshot{}, err
			}

			err = s.db.AddMatch(orderID, models.Match{
				OrderID: match.Order.ID,
				Qty:     match.Qty.String(),
				Price:   match.Order.price.String(),
			})
			if err != nil {
				return models.Order{}, nil, models.OrderBookSnapshot{}, err
			}
		}
	}

	order, err := s.db.UpdateOrderSizeFilled(orderID, orderSizeFilled.String())
	if err != nil {
		return models.Order{}, nil, models.OrderBookSnapshot{}, err
	}

	return order, matchOrders, ob.orderBookSnapshot(), nil
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

func (s *Service) GetCurrentOrders(userID int) ([]models.Order, error) {
	return s.db.GetNotFilledOrdersByUser(userID)
}

func (s *Service) GetOrders(userID int) ([]models.Order, error) {
	return s.db.GetOrdersByUser(userID)
}

func (s *Service) GetPairs() ([]string, error) {
	return s.db.GetPairs()
}

func (s *Service) GetPairsParams() ([]models.PairParams, error) {
	return s.db.GetPairsParams()
}

func (s *Service) GetPairPricePrecisions(pair string) ([]int32, error) {
	return s.db.GetPairPricePrecisions(pair)
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
