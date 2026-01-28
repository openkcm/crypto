package cmds

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/samber/oops"
	"github.com/spf13/cobra"

	slogctx "github.com/veqryn/slog-context"

	"github.com/openkcm/krypton/internal/config"
	"github.com/openkcm/krypton/pkg/concurrent"
	"github.com/openkcm/krypton/pkg/module"
)

var (
	startupTimeoutSec       int64
	gracefulShutdownSec     int64
	gracefulShutdownMessage string
)

func SetupRootCommand(rootCmd *cobra.Command, cfg *config.Config, modules map[*cobra.Command][]module.EmbeddedModule) error {
	seen := make(map[string]bool)
	for servedCmd, services := range modules {
		for _, s := range services {
			name := fmt.Sprintf("%s-%s", servedCmd.Use, s.Name())
			if seen[name] {
				return oops.Errorf("duplicate module %s", name)
			}
			seen[name] = true

			err := s.Init(cfg, servedCmd)
			if err != nil {
				return oops.Wrapf(err, "failed to init module %s", name)
			}
		}

		rootCmd.AddCommand(servedCmd)
	}

	rootCmd.PersistentFlags().Int64Var(&startupTimeoutSec, "startup-timeout", 30,
		"startup timeout seconds",
	)
	rootCmd.PersistentFlags().Int64Var(&gracefulShutdownSec, "graceful-shutdown", 5,
		"graceful shutdown seconds",
	)
	rootCmd.PersistentFlags().StringVar(&gracefulShutdownMessage, "graceful-shutdown-message",
		"Graceful shutdown in %d seconds",
		"graceful shutdown message",
	)

	return nil
}

func RunServeWithGracefulShutdown(modules []module.EmbeddedModule) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		ctx := slogctx.With(cmd.Context(), "command", cmd.Use)

		ctxStartup, cancel := context.WithTimeout(ctx, time.Duration(startupTimeoutSec)*time.Second)
		defer cancel()

		ctxShutdown, shutdown := context.WithCancel(ctx)
		ctxShutdown, _ = signal.NotifyContext(ctxShutdown, os.Interrupt, syscall.SIGTERM)
		defer shutdown()

		slogctx.Info(ctx, "Initializing command ...")

		services := make([]concurrent.ServiceFunc, len(modules))

		for i := range modules {
			name := modules[i].Name()
			runServe := modules[i].RunServe
			services[i] = func(ctxShutdown context.Context) error {
				return runServe(
					slogctx.With(ctxStartup, "module", name),
					slogctx.With(ctxShutdown, "module", name),
					shutdown,
				)
			}
		}
		slogctx.Info(ctx, "Initialization of command finalized with success.")

		slogctx.Info(ctx, "Application is serving")
		err := concurrent.Serve(ctxShutdown, shutdown, services...)
		if err != nil {
			slogctx.Error(ctx, "Failed to start the application", "error", err)
			return err
		}

		// Block here until context is done to ensure graceful shutdown
		<-ctxShutdown.Done()

		slogctx.Info(ctx, "Application is stopped")
		_, _ = fmt.Fprintln(os.Stderr, fmt.Sprintf(gracefulShutdownMessage, gracefulShutdownSec))
		time.Sleep(time.Duration(gracefulShutdownSec) * time.Second)
		return nil
	}
}
