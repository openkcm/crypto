package tenantmanager

import "github.com/openkcm/crypto-edge/internal/config"

type Server struct {
}

func NewServer(config *config.Config) *Server {
	return &Server{}
}

func (s *Server) Start() error {
	return nil
}

func (s *Server) Close() error {
	return nil
}
