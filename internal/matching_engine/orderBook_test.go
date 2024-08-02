package engine

import (
	"testing"
)

func TestMarketTrades(t *testing.T) {
	ob := NewOrderBook()

	ob.PlaceLimitOrder("40", NewOrder(1, true, "30"))
	ob.PlaceLimitOrder("50", NewOrder(1, true, "20"))

	ob.PlaceMarketOrder(NewOrder(7, false, "40"))
}

func TestPlaceMarketOrder(t *testing.T) {

}

func TestPlaceLimitOrder(t *testing.T) {

}

func TestCancelLimitOrder(t *testing.T) {

}
