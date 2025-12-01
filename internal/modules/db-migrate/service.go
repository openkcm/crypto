package dbmigrate

import (
	"context"

	"github.com/samber/oops"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/openkcm/crypto/internal/config"
	"github.com/openkcm/crypto/pkg/concurrent"
	"github.com/openkcm/crypto/pkg/module"
)

var (
	moduleName = "dbmigrate"
)

// dbMigrateModule implements module.EmbeddedModule interface.
type dbMigrateModule struct {
	config *config.Config
	fs     *pflag.FlagSet
}

var _ module.EmbeddedModule = (*dbMigrateModule)(nil)

func New() module.EmbeddedModule {
	return &dbMigrateModule{}
}

// Name implements main.embeddedService interface.
func (s *dbMigrateModule) Name() string { return moduleName }

// Init implements main.embeddedService interface.
func (s *dbMigrateModule) Init(cfg any, serveCmd *cobra.Command) error {
	//nolint: forcetypeassert
	s.config = cfg.(*config.Config)

	s.fs = serveCmd.Flags()
	return nil
}

// RunServe implements main.embeddedService interface.
func (s *dbMigrateModule) RunServe(ctxStartup, ctxShutdown context.Context, shutdown func()) (err error) {
	err = concurrent.Setup(ctxStartup, map[any]concurrent.SetupFunc{})
	if err != nil {
		return oops.In(moduleName).Wrapf(err, "failed to setup dbMigrateModule")
	}

	err = concurrent.Serve(ctxShutdown, shutdown,
		s.serveDatabaseMigrate,
	)
	if err != nil {
		return oops.In(moduleName).Wrapf(err, "Failed to server kmip server")
	}
	return nil
}

func (s *dbMigrateModule) serveDatabaseMigrate(ctx context.Context) error {
	return nil
}
