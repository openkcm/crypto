package encrypto

import "github.com/openkcm/crypto/internal/config"

type KMIPServer struct {
}

func NewKMIPServer(config *config.Config) *KMIPServer {
	return &KMIPServer{}
}

func (s *KMIPServer) Start() error {
	return nil
}

func (s *KMIPServer) Close() error {
	return nil
}
