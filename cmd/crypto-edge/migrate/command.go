package migrate

import (
	"context"
	"log/slog"
	"syscall"
	"time"

	"github.com/openkcm/common-sdk/pkg/health"
	"github.com/openkcm/common-sdk/pkg/logger"
	"github.com/openkcm/common-sdk/pkg/otlp"
	"github.com/openkcm/common-sdk/pkg/status"
	"github.com/samber/oops"
	"github.com/spf13/cobra"

	slogctx "github.com/veqryn/slog-context"

	"github.com/openkcm/crypto-edge/cmd/crypto-edge/common"
	"github.com/openkcm/crypto-edge/internal/config"
)

const (
	healthStatusTimeoutS = 5 * time.Second
)

// - Starts the status server
// - Starts the API Server
func run(ctx context.Context, cfg *config.Config) error {
	// LoggerConfig initialisation
	err := logger.InitAsDefault(cfg.Logger, cfg.Application)
	if err != nil {
		return oops.In("main").
			Wrapf(err, "Failed to initialise the logger")
	}

	slogctx.Debug(ctx, "Starting the application", slog.Any("config", cfg))

	// OpenTelemetry initialisation
	err = otlp.Init(ctx, &cfg.Application, &cfg.Telemetry, &cfg.Logger)
	if err != nil {
		return oops.In("main").
			Wrapf(err, "Failed to load the telemetry")
	}

	// Start status server
	startStatusServer(ctx, cfg)

	//Start Server Here
	//TODO: Add your code here

	<-ctx.Done()

	// Stop Server Here
	//TODO: Add your code here

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
		health.WithTimeout(healthStatusTimeoutS),
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

func Cmd(buildInfo string) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "migrate",
		Short: "Crypto Edge database migration",
		Long:  "Crypto Edge database migration command to handle database schema migrations.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := common.LoadConfig(buildInfo)
			if err != nil {
				return oops.In("main").Wrapf(err, "failed to load config")
			}

			err = run(cmd.Context(), cfg)
			if err != nil {
				return oops.In("main").Wrapf(err, "failed to run the api server")
			}

			return err
		},
	}

	return cmd
}
