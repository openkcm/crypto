package status

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

var (
	moduleName = "status"
)

// statusModule implements module.EmbeddedModule interface.
type statusModule struct {
	config *config.Config
	fs     *pflag.FlagSet
}

var _ module.EmbeddedModule = (*statusModule)(nil)

func New() module.EmbeddedModule {
	return &statusModule{}
}

// Name implements main.embeddedService interface.
func (s *statusModule) Name() string { return moduleName }

// Init implements main.embeddedService interface.
func (s *statusModule) Init(cfg any, serveCmd *cobra.Command) error {
	//nolint: forcetypeassert
	s.config = cfg.(*config.Config)

	s.fs = serveCmd.Flags()
	return nil
}

// RunServe implements main.embeddedService interface.
func (s *statusModule) RunServe(ctxStartup, ctxShutdown context.Context, shutdown func()) (err error) {
	err = concurrent.Setup(ctxStartup, map[any]concurrent.SetupFunc{})
	if err != nil {
		return oops.In(moduleName).Wrapf(err, "failed to setup statusModule")
	}

	err = concurrent.Serve(ctxShutdown, shutdown,
		s.serveStatusServer,
	)
	if err != nil {
		return oops.In(moduleName).Wrapf(err, "Failed to server kmip server")
	}
	return nil
}

func (s *statusModule) serveStatusServer(ctx context.Context) error {
	return modules.ServeStatus(ctx, &s.config.BaseConfig)
}
