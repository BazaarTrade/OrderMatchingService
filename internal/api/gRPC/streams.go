package server

import (
	"errors"

	"github.com/BazaarTrade/MatchingEngineProtoGen/pbM"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Server) StreamOrderBookSnapshot(req *pbM.Pair, stream pbM.MatchingEngine_StreamOrderBookSnapshotServer) error {
	OBSchan, exists := s.orderBookSnapshot[req.Pair]
	if !exists {
		s.logger.Error("failed to find order book snapshot chan", "pair", req.Pair)
		return errors.New("failed to find order book snapshot chan for: " + req.Pair)
	}

	for {
		select {
		case OBS, ok := <-OBSchan:
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

			s.logger.Info("Sent OBS", "pair", req.Pair)

		case <-stream.Context().Done():
			s.logger.Info("stream stopped by client", "pair", req.Pair)
			return nil
		}
	}
}

func (s *Server) StreamTrades(req *pbM.Pair, stream pbM.MatchingEngine_StreamTradesServer) error {
	tradesChan, exists := s.trades[req.Pair]
	if !exists {
		s.logger.Error("failed to find order book snapshot", "pair", req.Pair)
		return errors.New("failed to find order book snapshot for: " + req.Pair)
	}

	for {
		select {
		case trades, ok := <-tradesChan:
			if !ok {
				s.logger.Info("trades stream stopped manualy", "pair", req.Pair)
				return nil
			}

			var pbTrades = pbM.Trades{Pair: trades.Pair}

			for _, trade := range trades.Trades {
				pbTrades.Trades = append(pbTrades.Trades, &pbM.Trade{
					IsBid: trade.IsBid,
					Price: trade.Price,
					Qty:   trade.Qty,
					Time:  timestamppb.New(trade.Time),
				})
			}

			if err := stream.Send(&pbTrades); err != nil {
				s.logger.Error("failed to send trades", "Pair", trades.Pair, "error", err)
				continue
			}

			s.logger.Info("Sent trades", "pair", req.Pair)

		case <-stream.Context().Done():
			s.logger.Info("trades stream stopped by client", "pair", req.Pair)
			return nil
		}
	}
}
