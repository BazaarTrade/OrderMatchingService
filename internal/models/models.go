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

type Match struct {
	OrderID int
	Qty     string
	Price   string
}
