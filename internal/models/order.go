package models

type Order struct {
	ID        int
	UserID    int
	Pair      string
	OrderTypr string
	Price     string
	Qty       string
	IsBid     bool
}

type Match struct {
	OrderIDBid int
	OrderIDAsk int
	Qty        string
	Price      string
}
