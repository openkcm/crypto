package server

import (
	"context"

	"github.com/openkcm/crypto/internal/config"
	"github.com/openkcm/crypto/pkg/module/serve"
	slogctx "github.com/veqryn/slog-context"
)

type KMIPServer struct {
	config *config.Config
}

func NewKMIPServer(config *config.Config) *KMIPServer {
	return &KMIPServer{
		config: config,
	}
}

func (s *KMIPServer) Start(ctx context.Context) error {
	address := ":36444"
	slogctx.Info(ctx, "KMIP TCP server started", "address", address)
	err := serve.TCP(ctx, address, nil, s.connectionHandler, s)
	if err != nil {
		return err
	}
	slogctx.Info(ctx, "KMIP TCP server stopped", "address", address)
	return nil
}

func (s *KMIPServer) Close() error {
	return nil
}

func (s *KMIPServer) connectionHandler(bytes []byte) error {
	return nil
}
