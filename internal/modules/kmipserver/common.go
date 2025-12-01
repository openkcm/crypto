package kmipserver

import (
	"context"
	"errors"

	"github.com/samber/oops"

	slogctx "github.com/veqryn/slog-context"

	"github.com/openkcm/crypto/internal/actions"
	"github.com/openkcm/crypto/internal/config"
	"github.com/openkcm/crypto/kmip"
	"github.com/openkcm/crypto/kmip/kmipserver"
)

func createStartKMIPTcpServer(ctx context.Context, options ...kmipserver.Option) error {
	// Create and start KMIP server
	srv, err := kmipserver.NewServer(ctx, options...)
	if err != nil {
		return oops.Wrapf(err, "failed to start KMIP TCP server")
	}

	slogctx.Info(ctx, "Starting KMIP TCP server")
	go func() {
		err := srv.Serve()
		if err != nil && !errors.Is(err, kmipserver.ErrShutdown) {
			slogctx.Error(ctx, "KMIP TCP server failed to serve", "error", err)
		}
	}()

	slogctx.Info(ctx, "KMIP TCP server started")

	<-ctx.Done()

	slogctx.Info(ctx, "KMIP TCP server shutdown")
	err = srv.Shutdown()
	if err != nil {
		return oops.Wrapf(err, "failed to shutdown KMIP TCP server")
	}
	return nil
}

func configureRegistry(registry actions.Registry, cfgOp *config.KMIPOperation) actions.Registry {
	if len(cfgOp.Only) > 0 {
		operations := make([]kmip.Operation, 0, len(cfgOp.Only))
		for _, op := range cfgOp.Only {
			operations = append(operations, kmip.Operation(op))
		}
		registry.KeepOnly(operations...)
	} else if len(cfgOp.Exclude) > 0 {
		operations := make([]kmip.Operation, 0, len(cfgOp.Exclude))
		for _, op := range cfgOp.Exclude {
			operations = append(operations, kmip.Operation(op))
		}
		registry.Remove(operations...)
	}
	return registry
}
