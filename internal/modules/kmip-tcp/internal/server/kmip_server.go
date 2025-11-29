package server

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/openkcm/crypto/internal/config"
	"github.com/openkcm/crypto/internal/kmip"
	"github.com/openkcm/crypto/pkg/module/serve"
	slogctx "github.com/veqryn/slog-context"
)

var (
	defaultPool = &sync.Pool{
		New: func() any { return new(bytes.Buffer) },
	}
)

type KMIPServer struct {
	config *config.Config

	handler    kmip.Handler
	bufferPool *sync.Pool
}

func NewKMIPServer(config *config.Config, handler kmip.Handler) *KMIPServer {
	return &KMIPServer{
		config:     config,
		handler:    handler,
		bufferPool: defaultPool,
	}
}

func (s *KMIPServer) Start(ctx context.Context) error {
	address := ":36444"
	slogctx.Info(ctx, "KMIP TCP server started", "address", address)

	err := serve.TCP(ctx, address, nil, s.connectionHandler, s)
	if err != nil {
		return err
	}
	slogctx.Info(ctx, "KMIP TCP server stopped", "address", address)
	return nil
}

func (s *KMIPServer) Close() error {
	return nil
}

func (s *KMIPServer) connectionHandler(ctx context.Context, conn net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			slogctx.Error(ctx, "KMIP TCP connection handler panicked", "error", err)
		}
	}()
	s.handleByReadAndWriteFromToConnection(ctx, conn)
}

// handleConn reads multiple length-prefixed messages from a connection.
// Connection is reused until the client closes it or a read/write error occurs.
func (s *KMIPServer) handleByReadAndWriteFromToConnection(
	ctx context.Context,
	conn net.Conn,
) {
	defer conn.Close()

	buf := s.bufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		s.bufferPool.Put(buf)
	}()

	for {
		msg, err := readFramedMessage(ctx, conn, buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			slogctx.Warn(ctx, "read error", "error", err)
		}

		if len(msg) > 0 {
			customCtx := context.WithValue(ctx, "remote-address", conn.RemoteAddr().String())
			resp, hErr := s.handler(customCtx, msg)
			if hErr != nil {
				slogctx.Error(ctx, "failure on handling the request", "error", hErr)
				// Do not close connection; continue to read next message
			}

			if len(resp) > 0 {
				if wErr := writeFramedMessage(ctx, conn, resp); wErr != nil {
					slogctx.Error(ctx, "failure on writing the response", "error", wErr)
					return
				}
			}
		}

	}
}

// readFramedMessage reads a KMIP-compliant length-prefixed message from the connection.
//
// KMIP messages have an 8-byte header:
//
//	0-1: Protocol version (big-endian, e.g., 0x0200 for 2.0)
//	2-3: Reserved (0x0000)
//	4-7: Total message length (including header + payload)
func readFramedMessage(ctx context.Context, conn net.Conn, buf *bytes.Buffer) ([]byte, error) {
	_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	for {
		// If buffer has at least 8 bytes, we can parse header
		if buf.Len() >= 8 {
			header := buf.Bytes()[:8]
			protocolVersion := binary.BigEndian.Uint16(header[0:2])
			totalLength := binary.BigEndian.Uint32(header[4:8])

			if totalLength < 8 || totalLength > 16*1024*1024 {
				return nil, fmt.Errorf("invalid KMIP message length: %d", totalLength)
			}

			// Check if full message is in buffer
			if uint32(buf.Len()) >= totalLength {
				msg := buf.Next(int(totalLength))
				slogctx.Debug(ctx, "read KMIP message",
					"protocol_version", protocolVersion,
					"payload_length", totalLength-8,
				)
				return msg[8:], nil // return payload only
			}
		}

		// Read more data from connection
		tmp := make([]byte, 4096)
		n, err := conn.Read(tmp)
		if n > 0 {
			buf.Write(tmp[:n])
		}

		if err != nil {
			if errors.Is(err, io.EOF) && buf.Len() > 0 {
				return nil, fmt.Errorf("incomplete KMIP message on connection close")
			}
			return nil, err
		}
	}
}

// writeFramedMessage writes a KMIP-compliant message with retry on temporary errors.
//
// KMIP messages have an 8-byte header:
//
//	0-1: Protocol version (big-endian, e.g., 0x0200 for 2.0)
//	2-3: Reserved (0x0000)
//	4-7: Total message length (including header + payload)
func writeFramedMessage(ctx context.Context, conn net.Conn, data []byte) error {
	const (
		protocolVersion uint16 = 0x0200 // KMIP v2.0
		reserved        uint16 = 0x0000
	)

	// Total message length = header (8 bytes) + payload
	msgLen := 8 + len(data)
	frame := make([]byte, msgLen)

	// Write KMIP header
	binary.BigEndian.PutUint16(frame[0:2], protocolVersion)
	binary.BigEndian.PutUint16(frame[2:4], reserved)
	binary.BigEndian.PutUint32(frame[4:8], uint32(msgLen))

	// Copy payload
	copy(frame[8:], data)

	// Use retry logic for transient network errors
	return writeWithRetry(ctx, conn, frame, 3, 50*time.Millisecond)
}

// writeWithRetry writes data to the connection with retry for temporary network errors.
func writeWithRetry(ctx context.Context, conn net.Conn, data []byte, maxRetries int, delay time.Duration) error {
	var err error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		_ = conn.SetWriteDeadline(time.Now().Add(30 * time.Second))

		_, err = conn.Write(data)
		if err == nil {
			return nil
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}

		if !isTemporaryNetErr(err) {
			return err
		}

		time.Sleep(delay * time.Duration(1<<attempt))
	}

	return fmt.Errorf("write failed after %d retries: %w", maxRetries, err)
}

// isTemporaryNetErr checks if an error is a temporary network error.
func isTemporaryNetErr(err error) bool {
	var ne net.Error
	return errors.As(err, &ne) && (ne.Timeout() || ne.Temporary())
}
