package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/openkcm/common-sdk/pkg/logger"
	"github.com/openkcm/common-sdk/pkg/otlp"
	"github.com/spf13/cobra"

	slogctx "github.com/veqryn/slog-context"

	"github.com/openkcm/crypto/cmd"
	"github.com/openkcm/crypto/internal/config"
	dbmigrate "github.com/openkcm/crypto/internal/modules/db-migrate"
	"github.com/openkcm/crypto/internal/modules/kmipserver"
	"github.com/openkcm/crypto/internal/modules/status"
	"github.com/openkcm/crypto/pkg/cmds"
	"github.com/openkcm/crypto/pkg/module"
)

var (
	// BuildInfo will be set by the build system
	BuildInfo = "{}"
)

var (
	serveModules = []module.EmbeddedModule{
		kmipserver.NewCrypto(),
		status.New(),
	}
	migrateModules = []module.EmbeddedModule{
		dbmigrate.New(),
	}

	serve = &cobra.Command{
		Use:   "serve",
		Short: "Start all enabled Crypto service modules and run the server.",
		Args:  cobra.NoArgs,
		RunE:  cmds.RunServeWithGracefulShutdown(serveModules),
	}
	migrate = &cobra.Command{
		Use:   "dbmigrate",
		Short: "Run database schema migrations required by the Crypto service.",
		Args:  cobra.NoArgs,
		RunE:  cmds.RunServeWithGracefulShutdown(migrateModules),
	}
)

func main() {
	ctx := context.Background()

	cfg, err := config.LoadConfig(BuildInfo,
		"/etc/crypto",
		"$HOME/.crypto",
		".",
	)
	if err != nil {
		slogctx.Error(ctx, "Failed to load config", "error", err)
		os.Exit(1)
	}

	// LoggerConfig initialisation
	err = logger.InitAsDefault(cfg.Logger, cfg.Application)
	if err != nil {
		slogctx.Error(ctx, "Failed to init the logger", "error", err)
		os.Exit(1)
	}

	// OpenTelemetry initialisation
	err = otlp.Init(ctx, &cfg.Application, &cfg.Telemetry, &cfg.Logger)
	if err != nil {
		slogctx.Error(ctx, "Failed to init the logger", "error", err)
		os.Exit(1)
	}

	cmd.RootCmd.Version = cfg.Application.BuildInfo.Version
	err = cmds.SetupRootCommand(cmd.RootCmd, cfg, map[*cobra.Command][]module.EmbeddedModule{
		serve:   serveModules,
		migrate: migrateModules,
	})
	if err != nil {
		slogctx.Error(ctx, "Failed to setup root command", "error", err)
		os.Exit(1)
	}

	err = cmd.RootCmd.ExecuteContext(ctx)
	if err != nil {
		slog.Error("Failed to start the application", "error", err)
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
}
