package server

import (
	"context"

	"github.com/BazaarTrade/MatchingEngineProtoGen/pbM"
	"github.com/BazaarTrade/OrderMatchingService/internal/converter.go"
	"github.com/BazaarTrade/OrderMatchingService/internal/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) PlaceOrder(ctx context.Context, req *pbM.PlaceOrderReq) (*pbM.PlaceOrderRes, error) {
	s.logger.Info("PlaceOrder request", "userID", req.UserID)

	placeOrderReq := converter.ProtoPlaceOrderReqToModelsPlaceOrderReq(req)

	order, matchOrders, OBS, err := s.service.PlaceOrder(placeOrderReq)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	select {
	case s.orderBookSnapshot[OBS.Pair] <- OBS:
	default:
	}

	var (
		placeOrderRes = pbM.PlaceOrderRes{Order: converter.ModelsOrderToProtoOrder(order)}
		trades        = models.Trades{Pair: req.Pair}
	)

	if len(matchOrders) > 0 {
		if order.Status == "filled" {
			trades.Trades = append(trades.Trades, models.Trade{
				IsBid: order.IsBid,
				Price: order.Price,
				Qty:   order.SizeFilled,
				Time:  order.ClosedAt.Time,
			})
		}

		placeOrderRes.MatchOrders = make([]*pbM.Order, len(matchOrders))

		for i, order := range matchOrders {
			placeOrderRes.MatchOrders[i] = converter.ModelsOrderToProtoOrder(order)
			if order.Status == "filled" {
				trades.Trades = append(trades.Trades, models.Trade{
					IsBid: order.IsBid,
					Price: order.Price,
					Qty:   order.SizeFilled,
					Time:  order.ClosedAt.Time,
				})
			}
		}

		select {
		case s.trades[trades.Pair] <- trades:
		default:
		}
	}

	s.logger.Info("Order processed successfully", "orderID", order.ID)

	return &placeOrderRes, nil
}

func (s *Server) CancelOrder(ctx context.Context, req *pbM.OrderID) (*pbM.Order, error) {
	s.logger.Info("CancelOrder request", "orderID", req.OrderID)

	order, OBS, err := s.service.CancelOrder(int(req.OrderID))
	if err != nil {
		if err.Error() == "order book not found" {
			return nil, status.Errorf(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "failed to cancel order: %v", err)
	}

	s.orderBookSnapshot[OBS.Pair] <- OBS

	return converter.ModelsOrderToProtoOrder(order), nil
}

func (s *Server) GetCurrentOrders(ctx context.Context, req *pbM.UserID) (*pbM.Orders, error) {
	s.logger.Info("GetCurrentOrders request", "userID", req.UserID)

	orders, err := s.service.GetCurrentOrders(int(req.UserID))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get current orders: %v", err)
	}

	var pbOrders = pbM.Orders{Orders: make([]*pbM.Order, len(orders))}
	for i, order := range orders {
		pbOrders.Orders[i] = converter.ModelsOrderToProtoOrder(order)
	}
	return &pbOrders, nil
}

func (s *Server) GetOrders(ctx context.Context, req *pbM.UserID) (*pbM.Orders, error) {
	s.logger.Info("GetOrders request", "userID", req.UserID)

	orders, err := s.service.GetOrders(int(req.UserID))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get orders: %v", err)
	}

	var pbOrders = pbM.Orders{Orders: make([]*pbM.Order, len(orders))}
	for i, order := range orders {
		pbOrders.Orders[i] = converter.ModelsOrderToProtoOrder(order)
	}
	return &pbOrders, nil
}
