package kmipserver

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/openkcm/crypto/kmip"
	"github.com/openkcm/crypto/kmip/ttlv"
)

var ErrShutdown = errors.New("Server is shutting down")

// RequestHandler defines an interface for handling KMIP request messages.
// Implementations of this interface should process the provided RequestMessage
// and return an appropriate ResponseMessage. The context.Context parameter
// allows for request-scoped values, cancellation, and timeouts.
type RequestHandler interface {
	HandleRequest(ctx context.Context, req *kmip.RequestMessage) *kmip.ResponseMessage
}

// ConnectHook is a function that can be used to perform actions when a new connection is established.
// It takes a context.Context as input and returns a modified context.Context or an error
// that immediately terminates the connection, without calling any termination hook.
type ConnectHook func(context.Context) (context.Context, error)

// TerminateHook is a function that can be used to perform cleanup actions when a connection is terminated.
// It takes a context.Context as input.
//
// NOTE:  That context may have already been canceled. To pass it to context-cancellable function consider using WithoutCancel().
type TerminateHook func(context.Context)

// Server represents a KMIP server instance that manages incoming network connections,
// handles KMIP requests, and coordinates server lifecycle operations. It encapsulates
// the network listener, request handler, logging, context management for graceful
// shutdown, and a wait group for synchronizing goroutines. Additionally, it supports
// hooks for connect and terminate events, allowing customization of behavior when a client
// connects or disconnects.
type Server struct {
	listener net.Listener
	handler  RequestHandler

	ctx    context.Context
	cancel func()

	recvCtx    context.Context
	recvCancel func()

	wg *sync.WaitGroup

	onConnect ConnectHook
	onClose   TerminateHook
}

type Option func(*Server) error

// WithConnectHook sets the connect hook for the server, which is called when a new connection is established.
// This hook can be used to modify the context for the connection.
//
// Parameters:
//   - hook: The ConnectHook function to set.
//
// Returns:
//   - The Server instance with the connect hook set.
func WithConnectHook(hook ConnectHook) Option {
	return func(s *Server) error {
		s.onConnect = hook
		return nil
	}
}

// WithTerminateHook sets the terminate hook for the server, which is called when a connection is terminated.
// This hook can be used to perform any necessary cleanup or logging.
//
// Parameters:
//   - hook: The TerminateHook function to set.
//
// Returns:
//   - The Server instance with the terminate hook set.
func WithTerminateHook(hook TerminateHook) Option {
	return func(s *Server) error {
		s.onClose = hook
		return nil
	}
}

func WithHandler(handler RequestHandler) Option {
	return func(s *Server) error {
		s.handler = handler
		return nil
	}
}

func WithListener(listener net.Listener) Option {
	return func(s *Server) error {
		s.listener = listener
		return nil
	}
}

// ConnectHook wraps the configured connectHook function, calling it with the provided context.
// If no connectHook is set, it returns the original context without modification.
func (srv *Server) connectHook(ctx context.Context) (context.Context, error) {
	if srv.onConnect == nil {
		return ctx, nil
	}
	return srv.onConnect(ctx)
}

// TerminateHook wraps the configured terminateHook function, calling it with the provided context.
// If no terminateHook is set, it does nothing.
func (srv *Server) terminateHook(ctx context.Context) {
	if srv.onClose == nil {
		return
	}
	srv.onClose(ctx)
}

// NewServer creates and returns a new Server instance using the provided net.Listener and RequestHandler.
// It panics if the handler is nil. The function initializes internal contexts for server control and
// request reception, as well as a WaitGroup for managing goroutines.
func NewServer(ctx context.Context, options ...Option) (*Server, error) {
	rootCtx, cancel := context.WithCancel(ctx)
	recvCtx, recvCancel := context.WithCancel(rootCtx)

	srv := &Server{
		ctx:        rootCtx,
		cancel:     cancel,
		recvCtx:    recvCtx,
		recvCancel: recvCancel,
		wg:         &sync.WaitGroup{},
	}

	for _, opt := range options {
		if err := opt(srv); err != nil {
			return nil, err
		}
	}

	if err := srv.validate(); err != nil {
		return nil, err
	}

	return srv, nil
}

// Serve starts the KMIP server and listens for incoming client connections.
// It accepts connections in a loop, spawning a new goroutine to handle each connection.
// If the listener is closed, it returns ErrShutdown. Any other error encountered
// during Accept is returned immediately. The method blocks until the server is shut down.
func (srv *Server) Serve() error {
	slog.Info("KMIP server running", "bind", srv.listener.Addr())

	for {
		conn, err := srv.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return ErrShutdown
			}
			return fmt.Errorf("KMIP server shutting down: %w", err)
		}

		srv.wg.Add(1)
		go srv.handleConn(conn)
	}
}

// Shutdown gracefully shuts down the server by performing the following steps:
// 1. Logs a warning message indicating shutdown initiation.
// 2. Closes the listener to prevent new incoming connections.
// 3. Cancels the receive context to stop processing new requests.
// 4. Sets a timeout to force server context cancellation after 3 seconds.
// 5. Waits for all running requests to complete.
// 6. Cancels the server's root context.
// Returns any error encountered while closing the listener.
func (srv *Server) Shutdown() error {
	err := srv.listener.Close()

	srv.recvCancel()

	// Force-stop after grace period
	timer := time.AfterFunc(3*time.Second, srv.cancel)

	srv.wg.Wait()
	timer.Stop()

	srv.cancel()
	return err
}

func (srv *Server) handleConn(conn net.Conn) {
	defer srv.wg.Done()

	logger := slog.With("addr", conn.RemoteAddr())
	logger.Info("Connection established")

	// TLS handshake, if applicable
	var tlsState *tls.ConnectionState
	if tcon, ok := conn.(*tls.Conn); ok {
		if err := tcon.Handshake(); err != nil {
			_ = conn.Close()
			logger.Warn("TLS handshake failure", "err", err)
			return
		}
		cs := tcon.ConnectionState()
		tlsState = &cs
	}

	stream := newConn(srv.ctx, &connConfig{
		netCon:        conn,
		streamMaxSize: 15,
		logger:        logger,
	})
	defer stream.Close()

	// Per-connection context
	ctx := newConnContext(stream.ctx, conn.RemoteAddr().String(), tlsState)

	ctx, err := srv.connectHook(ctx)
	if err != nil {
		logger.Warn("Connect hook aborted connection", "err", err)
		return
	}
	defer srv.terminateHook(ctx)

	for {
		msg, err := stream.recv(srv.recvCtx)
		if err != nil {
			if errors.Is(err, io.EOF) {
				logger.Info("Client closed connection")
			} else {
				logger.Error("Failed to read from client", "err", err)

				// Return KMIP encoding error
				if ttlv.IsErrEncoding(err) {
					resp := srv.handleMessageError(ctx, msg, kmip.ResultReasonInvalidMessage, err.Error())
					_ = stream.send(resp)
				}
			}
			return
		}

		resp := srv.handleRequest(ctx, msg)
		if ctx.Err() != nil {
			logger.Warn("Request aborted", "err", ctx.Err())
			return
		}

		if err := stream.send(resp); err != nil {
			logger.Warn("Failed sending response", "err", err)
			return
		}
	}
}

func (srv *Server) handleMessageError(ctx context.Context, req *kmip.RequestMessage, reason kmip.ResultReason, message string) *kmip.ResponseMessage {
	return handleMessageError(ctx, req, Errorf(reason, "%s", message))
}

func (srv *Server) handleRequest(ctx context.Context, req *kmip.RequestMessage) (resp *kmip.ResponseMessage) {
	defer func() {
		if err := recover(); err != nil {
			resp = srv.handleMessageError(ctx, req, kmip.ResultReasonIllegalOperation, "")
		}
	}()
	resp = srv.handler.HandleRequest(ctx, req)
	return resp
}

func (srv *Server) validate() error {
	errs := make([]error, 0)
	if srv.handler == nil {
		errs = append(errs, errors.New("kmip handler is nil"))
	}
	if srv.listener == nil {
		errs = append(errs, errors.New("listener is nil"))
	}

	return errors.Join(errs...)
}
