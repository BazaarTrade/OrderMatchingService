package service

type Service interface {
	PlaceLimitOrder()
	CancelLimitOrder()
	PlaceMarketOrder()
}
