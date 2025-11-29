package kmiptcp

import (
	"context"
	"errors"

	"github.com/openkcm/crypto/internal/kmip"
	"github.com/samber/oops"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/openkcm/crypto/internal/config"
	"github.com/openkcm/crypto/internal/modules"
	"github.com/openkcm/crypto/internal/modules/kmip-tcp/internal/server"
	"github.com/openkcm/crypto/pkg/concurrent"
	"github.com/openkcm/crypto/pkg/module"
)

const (
	moduleName = "kmip"
)

// EmbeddedModule implements main.embeddedService interface.
type kmipTCPModule struct {
	config *config.Config

	handler kmip.Handler
	fs      *pflag.FlagSet
}

var _ module.EmbeddedModule = (*kmipTCPModule)(nil)

func New() module.EmbeddedModule {
	return &kmipTCPModule{}
}

// Name implements main.embeddedService interface.
func (s *kmipTCPModule) Name() string { return moduleName }

// Init implements main.embeddedService interface.
func (s *kmipTCPModule) Init(cfg any, cmd, serveCmd *cobra.Command) error {
	s.config = cfg.(*config.Config)
	s.handler = server.KMIPMessagesHandler(s.config)

	s.fs = serveCmd.Flags()
	return s.validate()
}

// RunServe implements main.embeddedService interface.
func (s *kmipTCPModule) RunServe(ctxStartup, ctxShutdown context.Context, shutdown func()) (err error) {
	err = concurrent.Setup(ctxStartup, map[any]concurrent.SetupFunc{})
	if err != nil {
		return oops.In(moduleName).Wrapf(err, "failed to setup kmipTCPModule")
	}

	err = concurrent.Serve(ctxShutdown, shutdown,
		s.serveMetrics,
		s.serveStatusServer,
		s.serveKMIPTCPServer,
	)
	if err != nil {
		return oops.In(moduleName).Wrapf(err, "Failed to server kmip server")
	}
	return nil
}

func (s *kmipTCPModule) serveMetrics(_ context.Context) error {
	return nil
}

func (s *kmipTCPModule) serveKMIPTCPServer(ctx context.Context) error {
	kmipServer := server.NewKMIPServer(s.config, s.handler)

	//Start Server Here
	err := kmipServer.Start(ctx)
	if err != nil {
		return oops.In(moduleName).
			Wrapf(err, "Failed to start the KMIP Server")
	}

	return nil
}

func (s *kmipTCPModule) serveStatusServer(ctx context.Context) error {
	return modules.ServeStatus(ctx, &s.config.BaseConfig)
}

func (s *kmipTCPModule) validate() error {
	if s.config == nil {
		return errors.New("missing configuration")
	}

	if s.handler == nil {
		return errors.New("missing handler")
	}

	return nil
}
