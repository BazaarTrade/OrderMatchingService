package gRPC

import (
	"context"
	"log"
	"net"

	"github.com/Moha192/OrderMatchingService/internal/models"
	pb "github.com/Moha192/OrderMatchingService/internal/proto"
	"github.com/Moha192/OrderMatchingService/internal/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type server struct {
	pb.UnimplementedMatchingEngineServer
	service service.Exchanger
}

func StartGRPCServer(service service.Exchanger) error {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return err
	}

	grpcServer := grpc.NewServer()

	s := &server{
		service: service,
	}

	pb.RegisterMatchingEngineServer(grpcServer, s)
	reflection.Register(grpcServer)
	log.Printf("Server is listening on port 50051...")

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
		return err
	}

	return nil
}

func (s *server) PlaceOrder(ctx context.Context, req *pb.PlaceOrderReq) (*pb.OrdersRes, error) {
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
		return nil, err
	}

	var res pb.OrdersRes
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

func (s *server) CancelOrder(ctx context.Context, req *pb.CancelOrderReq) (*pb.Order, error) {
	orderID := req.ID
	o, err := s.service.CancelOrder(orderID)
	if err != nil {
		return &pb.Order{}, err
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

func (s *server) GetCurrentOrders(ctx context.Context, req *pb.UserIDReq) (*pb.OrdersRes, error) {
	userID := req.UserID
	orders, err := s.service.GetCurrentOrders(userID)
	if err != nil {
		return nil, err
	}

	var res pb.OrdersRes
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

func (s *server) GetOrders(ctx context.Context, req *pb.UserIDReq) (*pb.OrdersRes, error) {
	userID := req.UserID
	orders, err := s.service.GetOrders(userID)
	if err != nil {
		return nil, err
	}

	var res pb.OrdersRes
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

func (s *server) CreateOrderBook(ctx context.Context, req *pb.OrderBookSymbol) (*emptypb.Empty, error) {
	err := s.service.AddOrderBook(req.Symbol)
	if err != nil {
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, nil
}

func (s *server) DeleteOrderBook(ctx context.Context, req *pb.OrderBookSymbol) (*emptypb.Empty, error) {
	err := s.service.DeleteOrderBook(req.Symbol)
	if err != nil {
		return &emptypb.Empty{}, err
	}
	return &emptypb.Empty{}, nil
}
