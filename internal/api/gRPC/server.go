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

func (s *server) PlaceOrder(ctx context.Context, req *pb.PlaceOrderReq) (*pb.PlaceOrderRes, error) {
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

	var res pb.PlaceOrderRes
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
		}

		res.UpdatedOrders = append(res.UpdatedOrders, &order)
	}
	return &res, nil
}

func (s *server) CancelOrder(ctx context.Context, req *pb.CancelOrderReq) (*pb.CancelOrderRes, error) {
	orderID := req.ID
	err := s.service.CancelOrder(orderID)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *server) GetCurrentOrders(ctx context.Context, req *pb.GetCurrentOrdersReq) (*pb.GetCurrentOrdersRes, error) {

	return nil, nil
}

func (s *server) GetOrderHistory(ctx context.Context, req *pb.GetOrderHistoryReq) (*pb.GetOrderHistoryRes, error) {
	return nil, nil
}

func (s *server) AddOrderBook(ctx context.Context, req *pb.AddOrderBookReq) (*pb.AddOrderBookRes, error) {
	err := s.service.AddOrderBook(req.Symbol)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *server) DeleteOrderBook(ctx context.Context, req *pb.DeleteOrderBookReq) (*pb.DeleteOrderBookRes, error) {
	return nil, nil
}
