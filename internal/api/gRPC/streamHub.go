package server

import (
	"errors"

	"github.com/BazaarTrade/OrderMatchingService/internal/models"
	"github.com/BazaarTrade/OrderMatchingService/internal/service"
)

var (
	ErrStreamHubNotFound      = errors.New("failed to find stream hub")
	ErrOBSChanNil             = errors.New("failed to find OBS stream hub chan")
	ErrMatchesChanNil         = errors.New("failed to find matches stream hub chan")
	ErrStreamHubAlreadyExists = errors.New("stream hub already exists")
)

// StreamHub made for managing channels that are used for delivering data to stream cycles
type StreamHub struct {
	OrderBookSnapshot chan models.OrderBookSnapshot
	Matches           chan []service.Match
}

func (s *Server) NewStreamHub(pair string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.streams[pair]; exists {
		return ErrStreamHubAlreadyExists
	}

	streamHub := &StreamHub{
		OrderBookSnapshot: make(chan models.OrderBookSnapshot),
		Matches:           make(chan []service.Match),
	}
	s.streams[pair] = streamHub

	return nil
}

func (s *Server) deleteStreamHubs() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for pair, stream := range s.streams {
		close(stream.OrderBookSnapshot)
		close(stream.Matches)
		delete(s.streams, pair)
		s.logger.Info("deleted stream hub", "pair", pair)
	}
}

func (s *Server) DeleteStreamHubByPair(pair string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	streamHub, exists := s.streams[pair]
	if !exists {
		s.logger.Error("failed to find stream hub", "pair", pair)
		return ErrStreamHubNotFound
	}

	OBS := streamHub.OrderBookSnapshot
	if OBS == nil {
		s.logger.Error("failed to find OBS stream hub chan", "pair", pair)
		return ErrOBSChanNil
	}

	matches := streamHub.Matches
	if matches == nil {
		s.logger.Error("failed to find matches stream hub chan", "pair", pair)
		return ErrMatchesChanNil
	}

	close(OBS)
	close(matches)
	delete(s.streams, pair)

	return nil
}

func (s *Server) sendOBS(OBS models.OrderBookSnapshot) {
	s.mu.RLock()
	streamHub, exists := s.streams[OBS.Pair]
	s.mu.RUnlock()

	if !exists {
		s.logger.Error("failed to find stream hub", "pair", OBS.Pair)
		return
	}

	if streamHub.OrderBookSnapshot == nil {
		s.logger.Error("OBS stream hub channel is nil", "pair", OBS.Pair)
		return
	}

	select {
	case streamHub.OrderBookSnapshot <- OBS:
	default:
		s.logger.Warn("OBS channel is full", "pair", OBS.Pair)
	}
}

func (s *Server) sendMatches(matches []service.Match) {
	s.mu.RLock()
	streamHub, exists := s.streams[matches[0].Pair]
	s.mu.RUnlock()

	if !exists {
		s.logger.Error("failed to find stream hub", "pair", matches[0].Pair)
		return
	}

	if streamHub.Matches == nil {
		s.logger.Error("matches stream hub channel is nil", "pair", matches[0].Pair)
		return
	}

	select {
	case streamHub.Matches <- matches:
	default:
		s.logger.Warn("matches channel is full", "pair", matches[0].Pair)
	}
}

func (s *Server) GetOBSChan(pair string) (chan models.OrderBookSnapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	streamHub, exists := s.streams[pair]
	if !exists {
		s.logger.Error("failed to find stream hub", "pair", pair)
		return nil, ErrStreamHubNotFound
	}

	if streamHub.OrderBookSnapshot == nil {
		s.logger.Error("failed to find OBS stream hub chan", "pair", pair)
		return nil, ErrOBSChanNil
	}

	return streamHub.OrderBookSnapshot, nil
}

func (s *Server) GetMatchesChan(pair string) (chan []service.Match, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	streamHub, exists := s.streams[pair]
	if !exists {
		s.logger.Error("failed to find stream hub", "pair", pair)
		return nil, ErrStreamHubNotFound
	}

	if streamHub.Matches == nil {
		s.logger.Error("failed to find matches stream hub chan", "pair", pair)
		return nil, ErrMatchesChanNil
	}

	return streamHub.Matches, nil
}
