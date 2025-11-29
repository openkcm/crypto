package serve

import (
	"context"
	"crypto/tls"
	"io"
	"net"

	"github.com/samber/oops"

	slogctx "github.com/veqryn/slog-context"
)

// TCP starts a TCP or TCP+TLS server, listens on the given address,
// and dispatches incoming byte streams to a provided handler.
//
// Parameters:
//
//	ctx       – cancellation context for graceful shutdown
//	address   – "host:port" binding address
//	tlsConfig – optional TLS configuration; nil → plain TCP
//	handler   – function invoked for each received message
//	closable  – optional resource that is closed on shutdown (DB, KV, module)
//
// Returns:
//
//	error – if listener fails or shutdown operations fail
func TCP(
	ctx context.Context,
	address string,
	tlsConfig *tls.Config,
	connectionHandler func(ctx context.Context, conn net.Conn),
	closable io.Closer,
) error {
	var ln net.Listener
	var err error

	if tlsConfig == nil {
		var lc net.ListenConfig
		ln, err = lc.Listen(ctx, "tcp", address)
	} else {
		ln, err = tls.Listen("tcp", address, tlsConfig)
	}
	if err != nil {
		return oops.Wrapf(err, "failed to listen on %s", address)
	}
	defer ln.Close()

	slogctx.Info(ctx, "TCP server started", "address", address)

	// Accept loop
	go acceptLoop(ctx, ln, connectionHandler)

	// Wait for context cancellation
	<-ctx.Done()

	if closable != nil {
		if err := closable.Close(); err != nil {
			return oops.Wrapf(err, "failed to close service")
		}
	}

	slogctx.Info(ctx, "TCP server shutdown", "address", address)
	return nil
}

// acceptLoop continuously accepts new connections until ctx is canceled.
func acceptLoop(
	ctx context.Context,
	ln net.Listener,
	connectionHandler func(ctx context.Context, conn net.Conn),
) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return // graceful shutdown
			}
			slogctx.Warn(ctx, "accept error", "error", err)
			continue
		}
		go connectionHandler(ctx, conn)
	}
}
