package gRPC

import (
	"context"
	"log"
	"log/slog"
	"net"

	pb "github.com/BazaarTrade/GeneratedProto/pb"
	"github.com/BazaarTrade/OrderMatchingService/internal/models"
	"github.com/BazaarTrade/OrderMatchingService/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	service service.Exchanger
	logger  *slog.Logger

	pb.UnimplementedMatchingEngineServer
}

func NewServer(service service.Exchanger, logger *slog.Logger) *Server {
	return &Server{
		service: service,
		logger:  logger,
	}
}

func (s *Server) StartGRPCServer() error {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return err
	}

	grpcServer := grpc.NewServer()

	pb.RegisterMatchingEngineServer(grpcServer, s)
	reflection.Register(grpcServer)
	log.Printf("Server is listening on port 50051...")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
		return err
	}

	return nil
}

func (s *Server) PlaceOrder(ctx context.Context, req *pb.PlaceOrderReq) (*pb.Orders, error) {
	s.logger.Info("PlaceOrder request", "user_id", req.UserID)

	var placeOrder = models.PlaceOrderReq{
		UserID: req.UserID,
		IsBid:  req.IsBid,
		Symbol: req.Symbol,
		Price:  req.Price,
		Qty:    req.Qty,
		Type:   req.Type,
	}

	modifiedOrders, err := s.service.PlaceOrder(placeOrder)
	if err != nil {
		if err.Error() == "Order book not found" {
			return nil, status.Errorf(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "Failed to palce order: %v", err)
	}

	var res pb.Orders
	for _, o := range modifiedOrders {
		var order = pb.Order{
			ID:         o.ID,
			UserID:     o.UserID,
			IsBid:      o.IsBid,
			Symbol:     o.Symbol,
			Price:      o.Price,
			Qty:        o.Qty,
			SizeFilled: o.SizeFilled,
			Status:     o.Status,
			Type:       o.Type,
			CreatedAt:  timestamppb.New(o.CreatedAt),
			ClosedAt:   timestamppb.New(o.ClosedAt.Time),
		}

		res.Orders = append(res.Orders, &order)
	}
	return &res, nil
}

func (s *Server) CancelOrder(ctx context.Context, req *pb.OrderID) (*pb.Order, error) {
	s.logger.Info("CancelOrder request", "order_id", req.OrderID)

	o, err := s.service.CancelOrder(req.OrderID)
	if err != nil {
		if err.Error() == "Order book not found" {
			return nil, status.Errorf(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "Failed to cancel order: %v", err)
	}

	var order = pb.Order{
		ID:         o.ID,
		UserID:     o.UserID,
		IsBid:      o.IsBid,
		Symbol:     o.Symbol,
		Price:      o.Price,
		Qty:        o.Qty,
		SizeFilled: o.SizeFilled,
		Status:     o.Status,
		Type:       o.Type,
		CreatedAt:  timestamppb.New(o.CreatedAt),
		ClosedAt:   timestamppb.New(o.ClosedAt.Time),
	}
	return &order, nil
}

func (s *Server) GetCurrentOrders(ctx context.Context, req *pb.UserID) (*pb.Orders, error) {
	s.logger.Info("GetCurrentOrders request", "user_id", req.UserID)

	orders, err := s.service.GetCurrentOrders(req.UserID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get current orders: %v", err)
	}

	var res pb.Orders
	for _, o := range orders {
		res.Orders = append(res.Orders, &pb.Order{
			ID:         o.ID,
			UserID:     o.UserID,
			IsBid:      o.IsBid,
			Symbol:     o.Symbol,
			Price:      o.Price,
			Qty:        o.Qty,
			SizeFilled: o.SizeFilled,
			Status:     o.Status,
			CreatedAt:  timestamppb.New(o.CreatedAt),
			ClosedAt:   timestamppb.New(o.ClosedAt.Time),
		})
	}
	return &res, nil
}

func (s *Server) GetOrders(ctx context.Context, req *pb.UserID) (*pb.Orders, error) {
	s.logger.Info("GetOrders request", "user_id", req.UserID)

	orders, err := s.service.GetOrders(req.UserID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to get orders: %v", err)
	}

	var res pb.Orders
	for _, o := range orders {
		res.Orders = append(res.Orders, &pb.Order{
			ID:         o.ID,
			UserID:     o.UserID,
			IsBid:      o.IsBid,
			Symbol:     o.Symbol,
			Price:      o.Price,
			Qty:        o.Qty,
			SizeFilled: o.SizeFilled,
			Status:     o.Status,
			CreatedAt:  timestamppb.New(o.CreatedAt),
			ClosedAt:   timestamppb.New(o.ClosedAt.Time),
		})
	}
	return &res, nil
}

func (s *Server) CreateOrderBook(ctx context.Context, req *pb.OrderBookSymbol) (*emptypb.Empty, error) {
	s.logger.Info("CreateOrderBook request", "symbol", req.Symbol)

	err := s.service.AddOrderBook(req.Symbol)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to create orderbook: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) DeleteOrderBook(ctx context.Context, req *pb.OrderBookSymbol) (*emptypb.Empty, error) {
	s.logger.Info("DeleteOrderBook request", "symbol", req.Symbol)

	err := s.service.DeleteOrderBook(req.Symbol)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to delete orderbook: %v", err)
	}
	return &emptypb.Empty{}, nil
}
