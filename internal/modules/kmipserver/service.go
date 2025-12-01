package kmipserver

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"

	"github.com/openkcm/common-sdk/pkg/commoncfg"
	"github.com/openkcm/crypto/internal/actions"
	"github.com/openkcm/crypto/internal/kmiphandler"
	"github.com/openkcm/crypto/kmip"
	"github.com/openkcm/crypto/pkg/module/serve"
	"github.com/samber/oops"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	slogctx "github.com/veqryn/slog-context"

	"github.com/openkcm/crypto/internal/config"
	"github.com/openkcm/crypto/kmip/kmipserver"
	"github.com/openkcm/crypto/pkg/concurrent"
	"github.com/openkcm/crypto/pkg/module"
)

const (
	moduleName = "kmip"
)

// kmipServerModule implements module.EmbeddedModule interface.
type kmipServerModule struct {
	config *config.Config

	fs *pflag.FlagSet
}

var _ module.EmbeddedModule = (*kmipServerModule)(nil)

func New() module.EmbeddedModule {
	return &kmipServerModule{}
}

// Name implements main.embeddedService interface.
func (s *kmipServerModule) Name() string { return moduleName }

// Init implements main.embeddedService interface.
func (s *kmipServerModule) Init(cfg any, serveCmd *cobra.Command) error {
	//nolint: forcetypeassert
	s.config = cfg.(*config.Config)

	s.fs = serveCmd.Flags()
	return s.validate()
}

// RunServe implements main.embeddedService interface.
func (s *kmipServerModule) RunServe(ctxStartup, ctxShutdown context.Context, shutdown func()) (err error) {
	err = concurrent.Setup(ctxStartup, map[any]concurrent.SetupFunc{})
	if err != nil {
		return oops.In(moduleName).Wrapf(err, "failed to setup kmipServerModule")
	}

	svcs := []concurrent.ServiceFunc{
		s.serveMetrics,
	}
	if s.config.KMIPServer.TCP.Enabled {
		svcs = append(svcs, s.serveKMIPTCPServer)
	}
	if s.config.KMIPServer.HTTP.Enabled {
		svcs = append(svcs, s.serveKMIPHTTPServer)
	}
	err = concurrent.Serve(ctxShutdown, shutdown, svcs...)
	if err != nil {
		return oops.In(moduleName).Wrapf(err, "Failed to server kmip server")
	}
	return nil
}

func (s *kmipServerModule) serveMetrics(_ context.Context) error {
	return nil
}

func (s *kmipServerModule) serveKMIPTCPServer(ctx context.Context) error {
	cfg := s.config.KMIPServer.TCP

	address := cfg.Address
	tlsConfig, _ := commoncfg.LoadMTLSConfig(cfg.TLS)

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

	handler, err := kmiphandler.NewHandler(
		configureRegistry(actions.NewRegistry(), &cfg.KMIPOperation),
		s.config,
	)
	if err != nil {
		return oops.Wrapf(err, "failed to create handler")
	}

	// Create and start server
	srv, err := kmipserver.NewServer(ctx,
		kmipserver.WithListener(ln),
		kmipserver.WithHandler(handler),
	)
	if err != nil {
		return oops.Wrapf(err, "failed to start KMIP TCP server")
	}

	slogctx.Info(ctx, "Starting KMIP TCP server")
	go func() {
		err := srv.Serve()
		if err != nil && !errors.Is(err, kmipserver.ErrShutdown) {
			slogctx.Error(ctx, "KMIP TCP server failed to serve", "error", err)
		}
	}()

	slogctx.Info(ctx, "KMIP TCP server started")

	<-ctx.Done()

	slogctx.Info(ctx, "KMIP TCP server shutdown")
	err = srv.Shutdown()
	if err != nil {
		return oops.Wrapf(err, "failed to shutdown KMIP TCP server")
	}

	return nil
}

func (s *kmipServerModule) serveKMIPHTTPServer(ctx context.Context) error {
	cfg := s.config.KMIPServer.HTTP

	address := cfg.Address
	tlsConfig, _ := commoncfg.LoadMTLSConfig(cfg.TLS)

	handler, err := kmiphandler.NewHandler(
		configureRegistry(actions.NewRegistry(), &cfg.KMIPOperation),
		s.config,
	)
	if err != nil {
		return oops.Wrapf(err, "failed to create handler")
	}

	mux := http.NewServeMux()
	mux.Handle(cfg.BasePath, kmipserver.NewHTTPHandler(handler))

	slogctx.Info(ctx, "Starting KMIP HTTP server")
	err = serve.HTTP(ctx, &http.Server{
		Addr:              address,
		Handler:           mux,
		TLSConfig:         tlsConfig,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		MaxHeaderBytes:    cfg.MaxHeaderBytes,
	})
	slogctx.Info(ctx, "Stopped KMIP HTTP server")
	return err
}

func (s *kmipServerModule) validate() error {
	if s.config == nil {
		return errors.New("missing configuration")
	}

	return nil
}

func configureRegistry(registry actions.Registry, cfgOp *config.KMIPOperation) actions.Registry {
	if len(cfgOp.Only) > 0 {
		operations := make([]kmip.Operation, 0, len(cfgOp.Only))
		for _, op := range cfgOp.Only {
			operations = append(operations, kmip.Operation(op))
		}
		registry.KeepOnly(operations...)
	} else if len(cfgOp.Exclude) > 0 {
		operations := make([]kmip.Operation, 0, len(cfgOp.Exclude))
		for _, op := range cfgOp.Exclude {
			operations = append(operations, kmip.Operation(op))
		}
		registry.Remove(operations...)
	}
	return registry
}
