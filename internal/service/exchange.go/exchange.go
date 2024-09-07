package exchange

import (
	"errors"
	"log/slog"

	"github.com/Moha192/OrderMatchingService/internal/models"
	"github.com/Moha192/OrderMatchingService/internal/repository"
	"github.com/shopspring/decimal"
)

type Exchange struct {
	db         repository.Storer
	orderBooks map[string]*OrderBook
	logger     *slog.Logger
}

func NewExchange(db repository.Storer, logger *slog.Logger) *Exchange {
	return &Exchange{
		db:         db,
		orderBooks: make(map[string]*OrderBook),
		logger:     logger,
	}
}

func (e *Exchange) AddOrderBook(symbol string) error {
	if _, ok := e.orderBooks[symbol]; ok {
		e.logger.Error("Order book already exists")
		return errors.New("Order book already exists")
	}
	e.orderBooks[symbol] = NewOrderBook(e.logger)
	e.logger.Info("OrderBook created successfully", "symbol", symbol)
	return nil
}

func (e *Exchange) DeleteOrderBook(symbol string) error {
	if _, ok := e.orderBooks[symbol]; !ok {
		e.logger.Error("Order book not found")
		return errors.New("Order book not found")
	}

	e.logger.Info("Order book deleted successfully", "symbol", symbol)
	return nil
}

func (e *Exchange) PlaceOrder(input models.PlaceOrderReq) ([]models.Order, error) {
	ob, ok := e.orderBooks[input.Symbol]
	if !ok {
		e.logger.Error("Order book not found")
		return nil, errors.New("Order book not found")
	}

	orderID, err := e.db.CreateOrder(input)
	if err != nil {
		return nil, err
	}

	priceDecimal, err := decimal.NewFromString(input.Price)
	if err != nil {
		e.logger.Error("Error converting price to decimal", "error", err)
		return nil, err
	}

	qtyDecimal, err := decimal.NewFromString(input.Qty)
	if err != nil {
		e.logger.Error("Error converting qty to decimal", "error", err)
		return nil, err
	}

	var (
		matches *[]Match
		order   = &Order{
			ID:        orderID,
			isBid:     input.IsBid,
			orderType: input.Type,
			price:     priceDecimal,
			qty:       qtyDecimal,
		}
	)

	switch order.orderType {
	case "limit":
		e.logger.Info(
			"Placing limit Order",
			"userID", input.UserID,
			"orderID", orderID,
			"symbol", input.Symbol,
			"isBid", input.IsBid,
			"price", input.Price,
			"qty", input.Qty,
		)
		matches, err = ob.placeLimitOrder(input.Price, order)
	case "market":
		e.logger.Info(
			"Placing market Order",
			"userID", input.UserID,
			"orderID", orderID,
			"symbol", input.Symbol,
			"isBid", input.IsBid,
			"price", input.Price,
			"qty", input.Qty,
		)
		matches, err = ob.placeMarketOrder(order)
	}
	if err != nil {
		return nil, err
	}

	var addMatchesReq = repository.AddMatchesReq{
		OrderID:         order.ID,
		OrderSizeFilled: order.sizeFilled.String(),
	}

	if matches != nil {
		for _, match := range *matches {
			var newMatch = repository.Match{
				Qty:                    match.qty.String(),
				Price:                  match.price.String(),
				CounterOrderID:         match.counterOrderID,
				CounterOrderSizeFilled: match.counterOrderSizeFilled.String(),
			}
			addMatchesReq.Matches = append(addMatchesReq.Matches, newMatch)
		}
	}

	updatedOrders, err := e.db.AddMatches(addMatchesReq)
	if err != nil {
		return nil, err
	}

	e.logger.Info("Order filled successfully", "orderID", orderID)
	return updatedOrders, nil
}

func (e *Exchange) CancelOrder(orderID int64) (models.Order, error) {
	order, err := e.db.GetOrderByOrderID(orderID)
	if err != nil {
		return models.Order{}, err
	}

	ob, ok := e.orderBooks[order.Symbol]
	if !ok {
		e.logger.Error("Order book not found")
		return models.Order{}, errors.New("Order book not found")
	}

	err = ob.cancelLimitOrder(orderID, order.Price, order.IsBid)
	if err != nil {
		return models.Order{}, err
	}

	order, err = e.db.SetStatusToCancel(orderID)
	if err != nil {
		return models.Order{}, err
	}
	return order, nil
}

func (e *Exchange) GetCurrentOrders(userID int64) ([]models.Order, error) {
	return e.db.GetNotFilledOrdersByUser(userID)
}

func (e *Exchange) GetOrders(userID int64) ([]models.Order, error) {
	return e.db.GetOrdersByUser(userID)
}
