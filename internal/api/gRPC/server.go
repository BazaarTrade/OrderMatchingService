package server

import (
	"log/slog"
	"net"
	"sync"

	"github.com/BazaarTrade/MatchingEngineProtoGen/pbM"
	"github.com/BazaarTrade/OrderMatchingService/internal/service"
	"google.golang.org/grpc"
)

type Server struct {
	pbM.UnimplementedMatchingEngineServer
	grpcServer *grpc.Server
	streams    map[string]*StreamHub
	service    *service.Service
	mu         sync.RWMutex
	logger     *slog.Logger
}

func New(service *service.Service, logger *slog.Logger) *Server {
	return &Server{
		streams: make(map[string]*StreamHub),
		service: service,
		logger:  logger,
	}
}

func (s *Server) Run(GRPC_PORT string) error {
	lis, err := net.Listen("tcp", GRPC_PORT)
	if err != nil {
		s.logger.Error("failed to listen", "err", err)
		return err
	}

	s.grpcServer = grpc.NewServer()

	pbM.RegisterMatchingEngineServer(s.grpcServer, s)
	s.logger.Info("server is listening on port " + GRPC_PORT)

	if err := s.grpcServer.Serve(lis); err != nil {
		s.logger.Error("failed to serve", "err", err)
		return err
	}
	return nil
}

func (s *Server) Stop() {
	s.deleteStreamHubs()
	s.grpcServer.GracefulStop()
}
