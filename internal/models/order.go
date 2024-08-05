package models

type Order struct {
	ID           int
	UserID       int
	CurrencyPair string
	Price        string
	Qty          string
	OrderType    string
	IsBid        bool
	IsFilled     bool
}

type Match struct {
	OrderIDBid int
	OrderIDAsk int
	Qty        string
	Price      string
}
