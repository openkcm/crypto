package kmiptcp

import (
	"context"
	"crypto/tls"
	"errors"
	"net"

	"github.com/openkcm/common-sdk/pkg/commoncfg"
	"github.com/openkcm/crypto/internal/actions"
	"github.com/samber/oops"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	slogctx "github.com/veqryn/slog-context"

	"github.com/openkcm/crypto/internal/config"
	"github.com/openkcm/crypto/internal/modules"
	"github.com/openkcm/crypto/kmip/kmipserver"
	"github.com/openkcm/crypto/pkg/concurrent"
	"github.com/openkcm/crypto/pkg/module"
)

const (
	moduleName = "kmip"
)

// EmbeddedModule implements main.embeddedService interface.
type kmipTCPModule struct {
	config *config.Config

	kmipActionRegistry actions.Registry
	fs                 *pflag.FlagSet
}

var _ module.EmbeddedModule = (*kmipTCPModule)(nil)

func New() module.EmbeddedModule {
	return &kmipTCPModule{}
}

// Name implements main.embeddedService interface.
func (s *kmipTCPModule) Name() string { return moduleName }

// Init implements main.embeddedService interface.
func (s *kmipTCPModule) Init(cfg any, cmd, serveCmd *cobra.Command) error {
	//nolint: forcetypeassert
	s.config = cfg.(*config.Config)
	s.kmipActionRegistry = actions.NewRegistry()

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
	address := s.config.KMIPServer.Address
	tlsConfig, _ := commoncfg.LoadMTLSConfig(s.config.KMIPServer.TLS)

	var ln net.Listener
	var err error
	if tlsConfig == nil {
		var lc net.ListenConfig
		ln, err = lc.Listen(ctx, "tcp", address)
	} else {
		ln, err = tls.Listen("tcp", address, tlsConfig)
	}
	if err != nil {
		return oops.Wrapf(err, "failed to listen on %s", address)
	}

	// Create and start server
	opts := []kmipserver.Option{
		kmipserver.WithListener(ln),
		kmipserver.WithHandler(NewHandler(s.kmipActionRegistry)),
	}
	srv, err := kmipserver.NewServer(ctx, opts...)
	if err != nil {
		return oops.Wrapf(err, "failed to start kmip TCP server")
	}

	slogctx.Info(ctx, "Starting KMIP server on "+address+"  ....")
	go func() {
		err := srv.Serve()
		if err != nil && !errors.Is(err, kmipserver.ErrShutdown) {
			slogctx.Error(ctx, "KMIP server failed to serve", "error", err)
		}
	}()

	slogctx.Info(ctx, "KMIP server started")

	<-ctx.Done()

	slogctx.Info(ctx, "KMIP server shutdown")
	err = srv.Shutdown()
	if err != nil {
		return oops.Wrapf(err, "failed to shutdown KMIP server")
	}

	//
	//kmipServer := server.NewKMIPServer(s.config)
	//
	////Start Server Here
	//err := kmipServer.Start(ctx)
	//if err != nil {
	//	return oops.In(moduleName).
	//		Wrapf(err, "Failed to start the KMIP Server")
	//}

	return nil
}

func (s *kmipTCPModule) serveStatusServer(ctx context.Context) error {
	return modules.ServeStatus(ctx, &s.config.BaseConfig)
}

func (s *kmipTCPModule) validate() error {
	if s.config == nil {
		return errors.New("missing configuration")
	}

	return nil
}
