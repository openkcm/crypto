package tenantmanager

import (
	"context"

	"github.com/samber/oops"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/openkcm/crypto/internal/config"
	"github.com/openkcm/crypto/internal/modules"
	"github.com/openkcm/crypto/pkg/concurrent"
	"github.com/openkcm/crypto/pkg/module"
)

const (
	moduleName = "manager"
)

// EmbeddedModule implements main.embeddedService interface.
type tenantManagerModule struct {
	config *config.Config
	fs     *pflag.FlagSet
}

var _ module.EmbeddedModule = (*tenantManagerModule)(nil)

func New() module.EmbeddedModule {
	return &tenantManagerModule{}
}

// Name implements main.embeddedService interface.
func (s *tenantManagerModule) Name() string { return moduleName }

// Init implements main.embeddedService interface.
func (s *tenantManagerModule) Init(cfg any, cmd, serveCmd *cobra.Command) error {
	//nolint: forcetypeassert
	s.config = cfg.(*config.Config)

	s.fs = serveCmd.Flags()
	return nil
}

// RunServe implements main.embeddedService interface.
func (s *tenantManagerModule) RunServe(ctxStartup, ctxShutdown context.Context, shutdown func()) (err error) {
	err = concurrent.Setup(ctxStartup, map[any]concurrent.SetupFunc{})
	if err != nil {
		return oops.In(moduleName).Wrapf(err, "failed to setup tenantManagerModule")
	}

	err = concurrent.Serve(ctxShutdown, shutdown,
		s.serveMetrics,
		s.serveStatusServer,
		s.serveHttpServer,
	)
	if err != nil {
		return oops.In(moduleName).Wrapf(err, "Failed to server kmip server")
	}
	return nil
}

func (s *tenantManagerModule) serveMetrics(_ context.Context) error {
	return nil
}

func (s *tenantManagerModule) serveHttpServer(ctx context.Context) error {
	return nil
}

func (s *tenantManagerModule) serveStatusServer(ctx context.Context) error {
	return modules.ServeStatus(ctx, &s.config.BaseConfig)
}
