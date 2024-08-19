package models

import (
	"time"
)

type Order struct {
	ID         int64
	UserID     int64
	IsBid      bool
	Symbol     string
	Price      string
	Qty        string
	SizeFilled string
	Status     string
	Type       string
	CreatedAt  time.Time
	ClosedAt   time.Time
}

type Match struct {
	Qty   string
	Price string
}

type PlaceOrderReq struct {
	UserID int64
	IsBid  bool
	Symbol string
	Price  string
	Qty    string
	Type   string //Market or Limit
}
