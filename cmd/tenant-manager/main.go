package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/openkcm/common-sdk/pkg/commoncfg"
	"github.com/openkcm/common-sdk/pkg/health"
	"github.com/openkcm/common-sdk/pkg/logger"
	"github.com/openkcm/common-sdk/pkg/otlp"
	"github.com/openkcm/common-sdk/pkg/status"
	"github.com/openkcm/common-sdk/pkg/utils"
	"github.com/samber/oops"
	"github.com/spf13/cobra"

	slogctx "github.com/veqryn/slog-context"

	"github.com/openkcm/crypto-edge/internal/config"
	"github.com/openkcm/crypto-edge/internal/tenantmanager"
)

const (
	defaultTimeout = 5
)

var (
	// BuildInfo will be set by the build system
	BuildInfo = "{}"

	isVersionCmd            bool
	gracefulShutdownSec     int64
	gracefulShutdownMessage string
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Tenant Manager Version",
	RunE: func(cmd *cobra.Command, _ []string) error {
		isVersionCmd = true

		value, err := utils.ExtractFromComplexValue(BuildInfo)
		if err != nil {
			return err
		}

		slog.InfoContext(cmd.Context(), value)

		return nil
	},
}

func rootCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "tenant-manager",
		Short: "Crypto Edge Tenant Manager",
		Long:  `Crypto Edge Tenant Manager - a service to manage tenants.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			// Load Configuration
			cfg, err := loadConfig()
			if err != nil {
				return oops.In("main").
					Wrapf(err, "Failed to load the configuration")
			}

			// Update Version
			err = commoncfg.UpdateConfigVersion(&cfg.BaseConfig, BuildInfo)
			if err != nil {
				return oops.In("main").
					Wrapf(err, "Failed to update the version configuration")
			}

			// LoggerConfig initialisation
			err = logger.InitAsDefault(cfg.Logger, cfg.Application)
			if err != nil {
				return oops.In("main").
					Wrapf(err, "Failed to initialise the logger")
			}

			// OpenTelemetry initialisation
			err = otlp.Init(ctx, &cfg.Application, &cfg.Telemetry, &cfg.Logger)
			if err != nil {
				return oops.In("main").
					Wrapf(err, "Failed to load the telemetry")
			}

			// Status Server Initialisation
			startStatusServer(ctx, cfg)

			// Create Server Here
			server := tenantmanager.NewServer(cfg)

			//Start Server Here
			err = server.Start()
			if err != nil {
				return oops.In("main").
					Wrapf(err, "Failed to start the server")
			}

			<-ctx.Done()

			// Stop Server Here
			err = server.Close()
			if err != nil {
				return oops.In("main").
					Wrapf(err, "Failed to close the server")
			}

			return nil
		},
	}

	cmd.AddCommand(versionCmd)

	return cmd
}

func execute() error {
	ctx, cancelOnSignal := signal.NotifyContext(
		context.Background(),
		os.Interrupt, syscall.SIGTERM,
	)
	defer cancelOnSignal()

	err := rootCmd().ExecuteContext(ctx)
	if err != nil {
		slogctx.Error(ctx, "Failed to start the application", "error", err)
		_, _ = fmt.Fprintln(os.Stderr, err)

		return err
	}

	// graceful shutdown so running goroutines may finish
	if !isVersionCmd {
		_, _ = fmt.Fprintln(os.Stderr, fmt.Sprintf(gracefulShutdownMessage, gracefulShutdownSec))
		time.Sleep(time.Duration(gracefulShutdownSec) * time.Second)
	}

	return nil
}

func startStatusServer(ctx context.Context, cfg *config.Config) {
	liveness := status.WithLiveness(
		health.NewHandler(
			health.NewChecker(health.WithDisabledAutostart()),
		),
	)

	healthOptions := make([]health.Option, 0)
	healthOptions = append(healthOptions,
		health.WithDisabledAutostart(),
		health.WithTimeout(defaultTimeout*time.Second),
		health.WithStatusListener(func(ctx context.Context, state health.State) {
			slogctx.Info(ctx, "readiness status changed", slog.String("status", string(state.Status)))
		}),
	)

	readiness := status.WithReadiness(
		health.NewHandler(
			health.NewChecker(healthOptions...),
		),
	)

	go func() {
		err := status.Start(ctx, &cfg.BaseConfig, liveness, readiness)
		if err != nil {
			slogctx.Error(ctx, "Failure on the status server", err)

			_ = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		}
	}()
}

func loadConfig() (*config.Config, error) {
	cfg := &config.Config{}

	err := commoncfg.LoadConfig(
		cfg,
		map[string]any{},
		"/etc/tenant-manager",
		"$HOME/.tenant-manager",
		".",
	)

	return cfg, err
}

func main() {
	err := execute()
	if err != nil {
		os.Exit(1)
	}
}
