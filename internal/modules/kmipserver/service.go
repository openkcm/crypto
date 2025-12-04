package kmipserver

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"

	"github.com/openkcm/common-sdk/pkg/commoncfg"
	"github.com/openkcm/crypto/internal/core"
	"github.com/openkcm/crypto/internal/core/authorization"
	"github.com/openkcm/crypto/internal/core/kmiphandlers"
	"github.com/openkcm/crypto/internal/core/operations"
	"github.com/samber/oops"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	slogctx "github.com/veqryn/slog-context"

	"github.com/openkcm/crypto/internal/config"
	"github.com/openkcm/crypto/kmip/kmipserver"
	"github.com/openkcm/crypto/pkg/concurrent"
	"github.com/openkcm/crypto/pkg/module"
	"github.com/openkcm/crypto/pkg/module/serve"
)

const (
	moduleCryptoName = "kmip-crypto"
)

// kmipEdgeServerModule implements module.EmbeddedModule interface.
type kmipCryptoServerModule struct {
	config *config.Config

	fs *pflag.FlagSet
}

var _ module.EmbeddedModule = (*kmipCryptoServerModule)(nil)

func NewCrypto() module.EmbeddedModule {
	return &kmipCryptoServerModule{}
}

// Name implements main.embeddedService interface.
func (s *kmipCryptoServerModule) Name() string { return moduleCryptoName }

// Init implements main.embeddedService interface.
func (s *kmipCryptoServerModule) Init(cfg any, serveCmd *cobra.Command) error {
	//nolint: forcetypeassert
	s.config = cfg.(*config.Config)

	s.fs = serveCmd.Flags()
	return s.validate()
}

// RunServe implements main.embeddedService interface.
func (s *kmipCryptoServerModule) RunServe(ctxStartup, ctxShutdown context.Context, shutdown func()) (err error) {
	err = concurrent.Setup(ctxStartup, map[any]concurrent.SetupFunc{})
	if err != nil {
		return oops.In(moduleCryptoName).Wrapf(err, "failed to setup kmipEdgeServerModule")
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
		return oops.In(moduleCryptoName).Wrapf(err, "Failed to server kmip server")
	}
	return nil
}

func (s *kmipCryptoServerModule) serveMetrics(_ context.Context) error {
	return nil
}

func (s *kmipCryptoServerModule) serveKMIPTCPServer(ctx context.Context) error {
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

	var proxyHttpKMIP *kmiphandlers.HttpKMIP
	if s.config.KMIPServer.Proxy != nil {
		proxyHttpKMIP = &kmiphandlers.HttpKMIP{
			Endpoint: s.config.KMIPServer.Proxy.Endpoint,
		}
	}

	handler, err := kmiphandlers.NewCryptoHandler(
		configureRegistry(operations.NewRegistry(), &cfg.KMIPOperation),
		core.NewServiceRegistry(s.config),
		authorization.NewCertificateAuthorizationHandler(s.config),
		proxyHttpKMIP,
	)
	if err != nil {
		return oops.Wrapf(err, "failed to create handler")
	}

	return createStartKMIPTcpServer(ctx, kmipserver.WithListener(ln), kmipserver.WithHandler(handler))
}

func (s *kmipCryptoServerModule) serveKMIPHTTPServer(ctx context.Context) error {
	cfg := s.config.KMIPServer.HTTP

	tlsConfig, _ := commoncfg.LoadMTLSConfig(cfg.TLS)

	var proxyHttpKMIP *kmiphandlers.HttpKMIP
	if s.config.KMIPServer.Proxy != nil {
		proxyHttpKMIP = &kmiphandlers.HttpKMIP{
			Endpoint: s.config.KMIPServer.Proxy.Endpoint,
		}
	}

	handler, err := kmiphandlers.NewCryptoHandler(
		configureRegistry(operations.NewRegistry(), &cfg.KMIPOperation),
		core.NewServiceRegistry(s.config),
		authorization.NewCertificateAuthorizationHandler(s.config),
		proxyHttpKMIP,
	)
	if err != nil {
		return oops.Wrapf(err, "failed to create handler")
	}

	mux := http.NewServeMux()
	mux.Handle(cfg.BasePath, kmipserver.NewHTTPHandler(
		kmipserver.WithHTTPMaxBodySize(15*1024*1024),
		kmipserver.WithRequestHandler(handler)),
	)

	slogctx.Info(ctx, "Starting KMIP HTTP server")
	err = serve.HTTP(ctx, &http.Server{
		Addr:              cfg.Address,
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

func (s *kmipCryptoServerModule) validate() error {
	if s.config == nil {
		return errors.New("missing configuration")
	}

	return nil
}
