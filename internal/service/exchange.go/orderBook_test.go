package exchange

import (
	"testing"

	"github.com/Moha192/OrderMatchingService/internal/repository"
)

func TestMarketTrades(t *testing.T) {
	var db *repository.Database
	ob := NewOrderBook(db)

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
