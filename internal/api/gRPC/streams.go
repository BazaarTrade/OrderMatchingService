package server

import (
	"github.com/BazaarTrade/MatchingEngineProtoGen/pbM"
	"github.com/BazaarTrade/OrderMatchingService/internal/converter"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) StreamOrderBookSnapshot(req *pbM.Pair, stream pbM.MatchingEngine_StreamOrderBookSnapshotServer) error {
	OBSChan, err := s.GetOBSChan(req.Pair)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	s.logger.Info("client connected to OBS stream", "pair", req.Pair)

	for {
		select {
		case OBS, ok := <-OBSChan:
			if !ok {
				s.logger.Info("OBS stream stopped normally", "pair", req.Pair)
				return nil
			}

			var pbOBS = pbM.OrderBookSnapshot{
				Pair:    OBS.Pair,
				Bids:    make([]*pbM.Limit, len(OBS.Bids)),
				Asks:    make([]*pbM.Limit, len(OBS.Asks)),
				BidsQty: OBS.BidsQty,
				AsksQty: OBS.AsksQty,
			}

			for i, limit := range OBS.Bids {
				pbOBS.Bids[i] = &pbM.Limit{
					Price: limit.Price,
					Qty:   limit.Qty,
				}
			}

			for i, limit := range OBS.Asks {
				pbOBS.Asks[i] = &pbM.Limit{
					Price: limit.Price,
					Qty:   limit.Qty,
				}
			}

			if err := stream.Send(&pbOBS); err != nil {
				s.logger.Error("failed to send OBS", "Pair", OBS.Pair, "error", err)
				continue
			}

			s.logger.Debug("sent OBS", "pair", req.Pair)

		case <-stream.Context().Done():
			s.logger.Info("client disconnected from OBS stream", "pair", req.Pair)
			return nil
		}
	}
}

func (s *Server) StreamTrades(req *pbM.Pair, stream pbM.MatchingEngine_StreamTradesServer) error {
	matchesChan, err := s.GetMatchesChan(req.Pair)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	s.logger.Info("client connected to trades stream", "pair", req.Pair)

	for {
		select {
		case matches, ok := <-matchesChan:
			if !ok {
				s.logger.Info("trades stream stopped normally", "pair", req.Pair)
				return nil
			}

			if err := stream.Send(converter.ServiceMatchesToPbMTrades(matches)); err != nil {
				s.logger.Error("failed to send trade", "Pair", req.Pair, "error", err)
				continue
			}

			s.logger.Debug("sent trades", "pair", req.Pair)

		case <-stream.Context().Done():
			s.logger.Info("client disconnected from trades stream", "pair", req.Pair)
			return nil
		}
	}
}
