package server

import (
	"context"

	"github.com/BazaarTrade/MatchingEngineProtoGen/pbM"
	"github.com/BazaarTrade/OrderMatchingService/internal/converter.go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) PlaceOrder(ctx context.Context, req *pbM.PlaceOrderReq) (*pbM.PlaceOrderRes, error) {
	order, matchOrders, matches, OBS, err := s.service.PlaceOrder(converter.ProtoPlaceOrderReqToModelsPlaceOrderReq(req))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	select {
	case s.orderBookSnapshot[OBS.Pair] <- OBS:
	default:
	}

	var (
		placeOrderRes = pbM.PlaceOrderRes{
			Order:       converter.ModelsOrderToProtoOrder(order),
			MatchOrders: make([]*pbM.Order, len(matchOrders)),
		}
	)

	for i, matchOrder := range matchOrders {
		placeOrderRes.MatchOrders[i] = converter.ModelsOrderToProtoOrder(matchOrder)
	}

	if len(matches) > 0 {
		select {
		case s.matches[req.Pair] <- matches:
		default:
		}
	}

	return &placeOrderRes, nil
}

func (s *Server) CancelOrder(ctx context.Context, req *pbM.OrderID) (*pbM.Order, error) {
	order, OBS, err := s.service.CancelOrder(int(req.OrderID))
	if err != nil {
		if err.Error() == "order book not found" {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to cancel order: %v", err)
	}

	s.orderBookSnapshot[OBS.Pair] <- OBS

	return converter.ModelsOrderToProtoOrder(order), nil
}
