package kmipserver

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/openkcm/krypton/kmip"
	"github.com/openkcm/krypton/kmip/ttlv"
)

type httpHandler struct {
	inner       RequestHandler
	maxBodySize int
}

type HttpHandlerOption func(*httpHandler)

func WithHTTPMaxBodySize(maxBodySize int) HttpHandlerOption {
	return func(req *httpHandler) {
		if maxBodySize < DEFAULT_MAX_BODY_SIZE {
			maxBodySize = DEFAULT_MAX_BODY_SIZE
		}

		req.maxBodySize = maxBodySize
	}
}

func WithRequestHandler(hdl RequestHandler) HttpHandlerOption {
	return func(req *httpHandler) {
		req.inner = hdl
	}
}

// NewHTTPHandler creates a new HTTP handler that wraps the provided RequestHandler.
// It returns an http.Handler that can be used to serve HTTP requests using the given handler logic.
//
// The method ServeHTTP handles incoming HTTP requests for the KMIP server. It supports POST requests with
// Content-Type headers of "text/xml", "application/json", or "application/octet-stream", and
// unmarshals the request body accordingly. The function enforces a maximum body size and validates
// the Content-Length header. It processes the KMIP request message and marshals the response using
// the appropriate format. If an error occurs during unmarshalling, a KMIP error response is sent
// back. Only POST requests are allowed; other methods receive a 405 Method Not Allowed response.
// Unsupported Content-Type headers result in a 406 Not Acceptable response, and requests with
// invalid or missing Content-Length headers receive a 411 Length Required response.
func NewHTTPHandler(opts ...HttpHandlerOption) http.Handler {
	handler := &httpHandler{maxBodySize: DEFAULT_MAX_BODY_SIZE}
	for _, opt := range opts {
		opt(handler)
	}
	return handler
}

// ServeHTTP handles incoming HTTP requests for the KMIP server. It supports POST requests with
// Content-Type headers of "text/xml", "application/json", or "application/octet-stream", and
// unmarshals the request body accordingly. The function enforces a maximum body size and validates
// the Content-Length header. It processes the KMIP request message and marshals the response using
// the appropriate format. If an error occurs during unmarshalling, a KMIP error response is sent
// back. Only POST requests are allowed; other methods receive a 405 Method Not Allowed response.
// Unsupported Content-Type headers result in a 406 Not Acceptable response, and requests with
// invalid or missing Content-Length headers receive a 411 Length Required response.
func (hdl httpHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if body := req.Body; body != nil {
		defer body.Close()
	}
	if req.Method != http.MethodPost {
		rw.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = io.WriteString(rw, "Only POST method are allowed")
		return
	}

	var unmarshaller func(data []byte, ptr any) error
	var marshaller func(data any) []byte

	contentType := req.Header.Get("Content-Type")
	switch contentType {
	case "text/xml", "application/vnd.kmip+xml":
		unmarshaller = ttlv.UnmarshalXML
		marshaller = ttlv.MarshalXML
	case "application/json", "application/vnd.kmip+json":
		unmarshaller = ttlv.UnmarshalJSON
		marshaller = ttlv.MarshalJSON
	case "application/octet-stream", "application/vnd.kmip+ttl":
		unmarshaller = ttlv.UnmarshalTTLV
		marshaller = ttlv.MarshalTTLV
	default:
		rw.WriteHeader(http.StatusUnsupportedMediaType)
		_, _ = io.WriteString(rw, "Unsupported Content-Type header")
		return
	}

	rw.Header().Set("Accept", contentType)

	contentLen, err := strconv.Atoi(req.Header.Get("Content-Length"))
	if err != nil || contentLen <= 0 {
		rw.WriteHeader(http.StatusLengthRequired)
		return
	}
	if contentLen > DEFAULT_MAX_BODY_SIZE {
		rw.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(rw, "The request is too large")
		return
	}

	buf := make([]byte, contentLen)
	if _, err := io.ReadFull(req.Body, buf); err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(rw, "Amount of data is lower than Content-Length")
		return
	}

	msg := kmip.RequestMessage{}
	var resp *kmip.ResponseMessage
	if err := unmarshaller(buf, &msg); err != nil {
		// If encoding error, send back the kmip error response
		resp = hdl.handleError(req.Context(), err, &msg)
	} else {
		ctx := newConnContext(req.Context(), req.RemoteAddr, req.TLS, req.Header)
		resp = hdl.inner.HandleRequest(ctx, &msg)
	}

	buf = marshaller(resp)
	if _, err := rw.Write(buf); err != nil {
		//TODO: Use user provided logger maybe ?
		slog.Error("Failed to write HTTP response", "err", err)
	}
}

func (hdl httpHandler) handleError(ctx context.Context, err error, req *kmip.RequestMessage) *kmip.ResponseMessage {
	return handleMessageError(ctx, req, err)
}
