package models

import (
	"database/sql"
	"time"
)

type Order struct {
	ID         int
	UserID     int
	IsBid      bool
	Pair       string
	Price      string
	Qty        string
	SizeFilled string
	Status     string
	Type       string
	CreatedAt  time.Time
	ClosedAt   sql.NullTime
}

type PlaceOrderReq struct {
	UserID int
	IsBid  bool
	Pair   string
	Price  string
	Qty    string
	Type   string //Market or Limit
}

type OrderBookSnapshot struct {
	Pair    string
	Bids    []Limit
	Asks    []Limit
	BidsQty string
	AsksQty string
}

type Limit struct {
	Price string
	Qty   string
}

type Trades struct {
	Pair   string
	Trades []Trade
}

type Trade struct {
	IsBid bool
	Price string
	Qty   string
	Time  time.Time
}

type Match struct {
	OrderID int
	Qty     string
	Price   string
}

type PairParams struct {
	Pair            string
	PricePrecisions []int32
	QtyPecision     int32
}
