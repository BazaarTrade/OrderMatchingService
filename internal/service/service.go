package service

type OrderService interface {
	PlaceLimitOrder()
	CancelLimitOrder()
	PlaceMarketOrder()
}
