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
	tenantmanager "github.com/openkcm/crypto/internal/modules/tenant-manager"
	"github.com/openkcm/crypto/pkg/cmds"
	"github.com/openkcm/crypto/pkg/module"
)

var (
	// BuildInfo will be set by the build system
	BuildInfo = "{}"
)

var (
	tmModules = []module.EmbeddedModule{
		tenantmanager.New(),
	}

	tmCmd = &cobra.Command{
		Use:   "tenant-manager",
		Short: "Crypto tenant-manager",
		Args:  cobra.NoArgs,
		RunE:  cmds.RunServeWithGracefulShutdown(tmModules),
	}
)

func main() {
	ctx := context.Background()

	cfg, err := config.LoadConfig(BuildInfo,
		"/etc/tenant-manager",
		"$HOME/.tenant-manager",
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
		tmCmd: tmModules,
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
