package exchange

import (
	"errors"

	"github.com/Moha192/OrderMatchingService/internal/models"
	"github.com/Moha192/OrderMatchingService/internal/repository"
	"github.com/shopspring/decimal"
)

type Exchange struct {
	db         repository.Storer
	OrderBooks map[string]*OrderBook
}

func NewExchange(db repository.Storer) *Exchange {
	return &Exchange{
		db:         db,
		OrderBooks: make(map[string]*OrderBook),
	}
}

func (e *Exchange) AddOrderBook(symbol string) error {
	if _, ok := e.OrderBooks[symbol]; ok {
		return errors.New("order book already exists")
	}
	e.OrderBooks[symbol] = NewOrderBook()
	return nil
}

func (e *Exchange) DeleteOrderBook(symbol string) error {
	if _, ok := e.OrderBooks[symbol]; !ok {
		return errors.New("order book does not exists")
	}

	return nil
}

func (e *Exchange) PlaceOrder(input models.PlaceOrderReq) ([]models.Order, error) {
	ob, ok := e.OrderBooks[input.Symbol]
	if !ok {
		return nil, errors.New("order book does not exists")
	}

	orderID, err := e.db.CreateOrder(input)
	if err != nil {
		return nil, err
	}

	priceDecimal, err := decimal.NewFromString(input.Price)
	if err != nil {
		return nil, err
	}
	qtyDecimal, err := decimal.NewFromString(input.Qty)
	if err != nil {
		return nil, err
	}
	var (
		matches *[]Match
		order   = &Order{
			ID:    orderID,
			IsBid: input.IsBid,
			Type:  input.Type,
			Price: priceDecimal,
			Qty:   qtyDecimal,
		}
	)

	switch order.Type {
	case "Limit":
		matches, err = ob.placeLimitOrder(input.Price, order)
	case "Market":
		matches, err = ob.placeMarketOrder(order)
	}
	if err != nil {
		return nil, err
	}

	var status string
	if order.Qty.IsZero() {
		status = "Filled"
	} else {
		status = "Filling"
	}
	var addMatchesReq = repository.AddMatchesReq{
		OrderID:         order.ID,
		OrderStatus:     status,
		OrderSizeFilled: order.SizeFilled.String(),
	}

	for _, match := range *matches {
		var newMatch = repository.Match{
			Qty:                    match.Qty.String(),
			Price:                  match.Price.String(),
			CounterOrderID:         match.CounterOrderID,
			CounterOrderStatus:     match.CounterOrderStatus,
			CounterOrderSizeFilled: match.CounterOrderSizeFilled.String(),
		}
		addMatchesReq.Matches = append(addMatchesReq.Matches, newMatch)
	}

	updatedOrders, err := e.db.AddMatches(addMatchesReq)
	if err != nil {
		return nil, err
	}

	return updatedOrders, nil
}

func (e *Exchange) CancelOrder(orderID int64) error {
	order, err := e.db.GetOrderByOrderID(orderID)
	if err != nil {
		return err
	}

	ob, ok := e.OrderBooks[order.Symbol]
	if !ok {
		return errors.New("order book not found")
	}

	err = ob.cancelLimitOrder(orderID, order.Price, order.IsBid)
	if err != nil {
		return err
	}
	return nil
}

func (e *Exchange) GetCurrentOrders(userID int64) {

}

func (e *Exchange) GetOrders(userID int64) {

}
