package modules

import (
	"context"
	"time"

	"github.com/openkcm/common-sdk/pkg/commoncfg"
	"github.com/openkcm/common-sdk/pkg/health"
	"github.com/openkcm/common-sdk/pkg/status"
	"github.com/samber/oops"

	slogctx "github.com/veqryn/slog-context"
)

const (
	healthStatusTimeoutS = 5 * time.Second
)

func ServeStatus(ctx context.Context, baseConfig *commoncfg.BaseConfig, ops ...health.Option) error {
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
			subctx := slogctx.With(ctx, "status", state.Status)
			//nolint:fatcontext
			for name, substate := range state.CheckState {
				subctx = slogctx.WithGroup(subctx, name)
				subctx = slogctx.With(subctx,
					"status", substate.Status,
					"result", substate.Result,
				)
			}
			slogctx.Info(subctx, "readiness status changed")
		}),
	)
	healthOptions = append(healthOptions, ops...)

	readiness := status.WithReadiness(
		health.NewHandler(
			health.NewChecker(healthOptions...),
		),
	)

	err := status.Start(ctx, baseConfig, liveness, readiness)
	if err != nil {
		return oops.In(baseConfig.Application.Name).Wrapf(err, "Failed to server status server")
	}
	return err
}
