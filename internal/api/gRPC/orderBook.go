package server

import (
	"context"

	"github.com/BazaarTrade/MatchingEngineProtoGen/pbM"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) CreateOrderBook(ctx context.Context, req *pbM.Pair) (*emptypb.Empty, error) {
	err := s.service.CreateOrderBook(req.Pair)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create orderbook: %v", err)
	}

	s.InitChans(req.Pair)

	return &emptypb.Empty{}, nil
}

func (s *Server) DeleteOrderBook(ctx context.Context, req *pbM.Pair) (*emptypb.Empty, error) {
	s.RemoveChans(req.Pair)

	err := s.service.DeleteOrderBook(req.Pair)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete orderbook: %v", err)
	}
	return &emptypb.Empty{}, nil
}
