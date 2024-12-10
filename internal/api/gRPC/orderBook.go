package server

import (
	"context"

	"github.com/BazaarTrade/MatchingEngineProtoGen/pbM"
	"github.com/BazaarTrade/OrderMatchingService/internal/converter.go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) CreateOrderBook(ctx context.Context, req *pbM.PairParams) (*emptypb.Empty, error) {
	s.logger.Info("CreateOrderBook request", "symbol", req.Pair)

	err := s.service.CreateOrderBook(req.Pair, req.PricePrecisions, req.QtyPrecision)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create orderbook: %v", err)
	}

	s.InitChans(req.Pair)

	return &emptypb.Empty{}, nil
}

func (s *Server) DeleteOrderBook(ctx context.Context, req *pbM.Pair) (*emptypb.Empty, error) {
	s.logger.Info("DeleteOrderBook request", "symbol", req.Pair)

	s.RemoveChans(req.Pair)

	err := s.service.DeleteOrderBook(req.Pair)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete orderbook: %v", err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) GetPairsParams(ctx context.Context, req *emptypb.Empty) (*pbM.PairsParams, error) {
	s.logger.Info("GetPairsParams request")

	pairsParams, err := s.service.GetPairsParams()
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get pairs precisions")
	}

	var pbPairsParams = &pbM.PairsParams{PairParams: make([]*pbM.PairParams, len(pairsParams))}
	for i, pairParams := range pairsParams {
		pbPairsParams.PairParams[i] = converter.ModelsPairsParamsToProtoPairParams(pairParams)
	}
	return pbPairsParams, nil
}

func (s *Server) GetPairPricePrecisions(ctx context.Context, req *pbM.Pair) (*pbM.PairPricePrecisions, error) {
	s.logger.Info("GatPairPricePrecisions request")

	pairPricePrecisions, err := s.service.GetPairPricePrecisions(req.Pair)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to get pair price precisions")
	}
	return &pbM.PairPricePrecisions{Precisions: pairPricePrecisions}, nil
}
