package serve

import (
	"context"
	"net"

	"github.com/samber/oops"
	"google.golang.org/grpc"
)

// GRPC starts a gRPC server and blocks until either:
//  1. The server returns an error, or
//  2. The provided context is canceled (e.g., shutdown signal).
//
// Behavior:
//   - Listens on the given TCP address.
//   - Runs srv.Serve(...) in a goroutine.
//   - If srv.Serve returns an error, the function returns that error.
//   - If ctx.Done() fires, the server is gracefully stopped:
//     srv.GracefulStop()
//     GracefulStop allows existing RPCs to finish but stops accepting new ones.
//
// Logging:
//   - Logs startup and successful shutdown messages using slogctx.
//
// Parameters:
//   - ctx: Context used to cancel and gracefully shut down the server.
//   - address: TCP address (e.g., ":8080") to listen on.
//   - srv: Fully configured *grpc.Server instance.
func GRPC(ctx context.Context, address string, srv *grpc.Server) error {
	// Start listening on TCP
	var lc net.ListenConfig
	ln, err := lc.Listen(ctx, "tcp", address)
	if err != nil {
		return oops.Wrapf(err, "failed to listen on %s", address)
	}

	// Capture Serve() error in a buffered channel so goroutine cannot deadlock.
	errChan := make(chan error, 1)

	// Start serving gRPC requests
	go func() {
		errChan <- srv.Serve(ln)
	}()

	select {
	case serveErr := <-errChan:
		// Serve returned an error (including normal closure)
		if serveErr != nil {
			return oops.Wrapf(serveErr, "failed to serve")
		}
	case <-ctx.Done():
		// Shutdown requested; gracefully stop the server
		srv.GracefulStop()
	}
	return nil
}
