package serve

import (
	"context"
	"errors"
	"net/http"

	"github.com/samber/oops"
)

// HTTP starts an HTTP server (with optional TLS) and blocks until:
//
//   - The server returns an error (e.g., port in use, fatal serve error), OR
//   - The provided context is cancelled — triggering a graceful shutdown.
//
// Behavior:
//   - If srv.TLSConfig is nil → HTTP server runs via ListenAndServe().
//   - If srv.TLSConfig is non-nil → HTTPS server runs via ListenAndServeTLS().
//   - When ctx is cancelled, the server is gracefully shut down using srv.Shutdown(ctx),
//     which allows in-flight requests to finish before closing listeners.
//
// Parameters:
//   - ctx: Cancellation context used for graceful shutdown.
//   - srv: Pre-configured http.Server instance (address, handlers, TLSConfig, timeouts, etc.).
//
// Returns:
//   - error: Serve or shutdown error, wrapped via oops, or nil if successfully stopped.
//
// Notes:
//   - Graceful shutdown does NOT terminate ongoing HTTP handlers abruptly.
//   - ListenAndServe* always returns a non-nil error — http.ErrServerClosed indicates
//     a NORMAL graceful shutdown and is not treated as a failure.
func HTTP(ctx context.Context, srv *http.Server) error {
	errc := make(chan error, 1)

	// Start the server in a separate goroutine.
	go func() {
		if srv.TLSConfig == nil {
			errc <- srv.ListenAndServe()
		} else {
			// Empty cert/key paths → Use srv.TLSConfig certificates only.
			errc <- srv.ListenAndServeTLS("", "")
		}
	}()

	var serveErr error

	select {
	case serveErr = <-errc:
		// The server stopped on its own (could be ErrServerClosed or fatal error).
	case <-ctx.Done():
		// Context cancelled → trigger graceful shutdown.
		shutdownErr := srv.Shutdown(ctx)
		if shutdownErr != nil {
			return oops.Wrapf(shutdownErr, "http graceful shutdown failed")
		}
		return nil
	}

	// If server exited normally via Shutdown(), serveErr == http.ErrServerClosed.
	if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
		return oops.Wrapf(serveErr, "failed to serve HTTP")
	}
	return nil
}
