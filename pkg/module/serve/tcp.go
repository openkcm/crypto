package serve

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	"github.com/samber/oops"

	slogctx "github.com/veqryn/slog-context"
)

// TCP starts a TCP or TCP+TLS server, listens on the given address,
// and dispatches incoming byte streams to a provided handler.
//
// Behavior:
//   - Creates a net.Listener (TLS if tlsConfig != nil)
//   - Spawns an accept-loop in a goroutine
//   - Each TCP connection is processed in its own goroutine
//   - Incoming bytes are read into a reusable buffer from sync.Pool
//   - Handler receives an immutable copy of the received bytes
//   - Server shuts down automatically when ctx is canceled
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
	handler func([]byte) error,
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

	// Shared buffer pool for all connections
	bufferPool := &sync.Pool{
		New: func() any { return new(bytes.Buffer) },
	}

	// Accept loop
	go acceptLoop(ctx, ln, bufferPool, handler)

	// Block until the server is asked to stop
	<-ctx.Done()

	if closable != nil {
		if err := closable.Close(); err != nil {
			return oops.Wrapf(err, "failed to close service")
		}
	}

	return nil
}

// acceptLoop continuously accepts new TCP connections until context is canceled.
func acceptLoop(
	ctx context.Context,
	ln net.Listener,
	pool *sync.Pool,
	handler func([]byte) error,
) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			if ctx.Err() != nil {
				// Graceful shutdown
				return
			}
			slogctx.Warn(ctx, "accept error", "error", err)
			continue
		}

		// Each connection is handled concurrently
		go handleConn(ctx, pool, conn, handler)
	}
}

// handleConn reads streamed TCP data from a single connection.
// It returns when the client disconnects or a read error occurs.
func handleConn(
	ctx context.Context,
	pool *sync.Pool,
	conn net.Conn,
	handler func([]byte) error,
) {
	defer conn.Close()

	//nolint: forcetypeassert
	buf := pool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		pool.Put(buf)
	}()

	buf.Reset()
	tmp := make([]byte, 4096) // working chunk

	for {
		// Optional: Protect against slow clients
		_ = conn.SetReadDeadline(time.Now().Add(30 * time.Second))

		n, err := conn.Read(tmp)
		if n > 0 {
			buf.Write(tmp[:n])

			// Make copy so handler cannot race with buffer reuse
			dataCopy := append([]byte(nil), buf.Bytes()...)

			if hErr := handler(dataCopy); hErr != nil {
				slogctx.Error(ctx, "handler returned error", "error", hErr)
			}
		}

		if err != nil {
			if !errors.Is(err, io.EOF) {
				slogctx.Warn(ctx, "tcp read error", "error", err)
			}
			return
		}
	}
}
