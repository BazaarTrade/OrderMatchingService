package server

import (
	"errors"

	"github.com/BazaarTrade/MatchingEngineProtoGen/pbM"
	"github.com/BazaarTrade/OrderMatchingService/internal/converter.go"
)

func (s *Server) StreamOrderBookSnapshot(req *pbM.Pair, stream pbM.MatchingEngine_StreamOrderBookSnapshotServer) error {
	OBSChan, exists := s.orderBookSnapshot[req.Pair]
	if !exists {
		s.logger.Error("failed to find order book snapshot chan", "pair", req.Pair)
		return errors.New("failed to find order book snapshot chan for: " + req.Pair)
	}

	for {
		select {
		case OBS, ok := <-OBSChan:
			if !ok {
				s.logger.Info("OBS stream stopped manualy", "pair", req.Pair)
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
			s.logger.Info("stream stopped by client", "pair", req.Pair)
			return nil
		}
	}
}

func (s *Server) StreamTrades(req *pbM.Pair, stream pbM.MatchingEngine_StreamTradesServer) error {
	matchesChan, exists := s.matches[req.Pair]
	if !exists {
		s.logger.Error("failed to find matches chan", "pair", req.Pair)
		return errors.New("failed to find matches chan for: " + req.Pair)
	}

	for {
		select {
		case matches, ok := <-matchesChan:
			if !ok {
				s.logger.Info("trades stream stopped manualy", "pair", req.Pair)
				return nil
			}

			if err := stream.Send(converter.ServiceMatchesToPbMTrades(matches)); err != nil {
				s.logger.Error("failed to send trade", "Pair", req.Pair, "error", err)
				continue
			}

			s.logger.Debug("sent trades", "pair", req.Pair)

		case <-stream.Context().Done():
			s.logger.Info("trades stream stopped by client", "pair", req.Pair)
			return nil
		}
	}
}
