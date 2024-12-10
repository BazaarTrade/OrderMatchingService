package converter

import (
	"github.com/BazaarTrade/MatchingEngineProtoGen/pbM"
	"github.com/BazaarTrade/OrderMatchingService/internal/models"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ModelsOrderToProtoOrder(order models.Order) *pbM.Order {
	return &pbM.Order{
		ID:         int64(order.ID),
		UserID:     int64(order.UserID),
		IsBid:      order.IsBid,
		Pair:       order.Pair,
		Price:      order.Price,
		Qty:        order.Qty,
		SizeFilled: order.SizeFilled,
		Status:     order.Status,
		Type:       order.Type,
		CreatedAt:  timestamppb.New(order.CreatedAt),
		ClosedAt:   timestamppb.New(order.ClosedAt.Time),
	}
}

func ProtoPlaceOrderReqToModelsPlaceOrderReq(pbPlaceOrderReq *pbM.PlaceOrderReq) models.PlaceOrderReq {
	return models.PlaceOrderReq{
		UserID: int(pbPlaceOrderReq.UserID),
		IsBid:  pbPlaceOrderReq.IsBid,
		Pair:   pbPlaceOrderReq.Pair,
		Price:  pbPlaceOrderReq.Price,
		Qty:    pbPlaceOrderReq.Qty,
		Type:   pbPlaceOrderReq.Type,
	}
}

func ModelsPairsParamsToProtoPairParams(pairPrecisions models.PairParams) *pbM.PairParams {
	return &pbM.PairParams{
		Pair:            pairPrecisions.Pair,
		PricePrecisions: pairPrecisions.PricePrecisions,
		QtyPrecision:    pairPrecisions.QtyPecision,
	}
}
