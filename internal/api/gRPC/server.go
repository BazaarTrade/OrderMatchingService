package server

import (
	"log/slog"
	"net"
	"sync"

	"github.com/BazaarTrade/MatchingEngineProtoGen/pbM"
	"github.com/BazaarTrade/OrderMatchingService/internal/models"
	"github.com/BazaarTrade/OrderMatchingService/internal/service"
	"google.golang.org/grpc"
)

type Server struct {
	pbM.UnimplementedMatchingEngineServer
	service           *service.Service
	orderBookSnapshot map[string]chan models.OrderBookSnapshot
	matches           map[string]chan []service.Match
	mu                sync.RWMutex
	logger            *slog.Logger
}

func New(serviceT *service.Service, logger *slog.Logger) *Server {
	return &Server{
		service:           serviceT,
		logger:            logger,
		orderBookSnapshot: make(map[string]chan models.OrderBookSnapshot),
		matches:           make(map[string]chan []service.Match),
	}
}

func (s *Server) Run() error {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		s.logger.Error("failed to listen", "err", err)
		return err
	}

	grpcServer := grpc.NewServer()

	pbM.RegisterMatchingEngineServer(grpcServer, s)
	s.logger.Info("server is listening on port 50051...")

	if err := grpcServer.Serve(lis); err != nil {
		s.logger.Error("failed to serve", "err", err)
		return err
	}
	return nil
}

func (s *Server) InitChans(pair string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.orderBookSnapshot[pair] = make(chan models.OrderBookSnapshot)
	s.matches[pair] = make(chan []service.Match)
}

func (s *Server) RemoveChans(pair string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if ch, exists := s.orderBookSnapshot[pair]; exists {
		close(ch)
		delete(s.orderBookSnapshot, pair)
	}

	if ch, exists := s.matches[pair]; exists {
		close(ch)
		delete(s.matches, pair)
	}
}
