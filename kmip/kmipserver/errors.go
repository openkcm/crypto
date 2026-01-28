package kmipserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/openkcm/krypton/kmip"
	"github.com/openkcm/krypton/kmip/ttlv"
)

// Error represents a KMIP error with a specific reason and an optional message.
// It encapsulates the KMIP ResultReason and a human-readable error message.
type Error struct {
	Reason  kmip.ResultReason
	Message string
}

func (e Error) Error() string {
	if e.Message == "" {
		return ttlv.EnumStr(e.Reason)
	}
	return e.Message
}

// Errorf creates a new Error with the specified kmip.ResultReason and a formatted message.
// The message is formatted according to the given format specifier and arguments, similar to fmt.Sprintf.
//
// Parameters:
//   - reason: The kmip.ResultReason indicating the reason for the error.
//   - format: A format string compatible with fmt.Sprintf.
//   - args:   Variadic arguments to be formatted into the message.
//
// Returns:
//   - error: An Error instance containing the reason and formatted message.
func Errorf(reason kmip.ResultReason, format string, args ...any) error {
	return Error{
		Reason:  reason,
		Message: fmt.Sprintf(format, args...),
	}
}

var (
	ErrOperationNotSupported = Errorf(kmip.ResultReasonOperationNotSupported, "Operation not supported")
	ErrFeatureNotSupported   = Errorf(kmip.ResultReasonFeatureNotSupported, "Feature not supported")
	ErrMissingData           = Errorf(kmip.ResultReasonMissingData, "Missing data")
	ErrItemNotFound          = Errorf(kmip.ResultReasonItemNotFound, "Item not found")
	ErrPermissionDenied      = Errorf(kmip.ResultReasonPermissionDenied, "Permission denied")
	ErrInvalidMessage        = Errorf(kmip.ResultReasonInvalidMessage, "Invalid message")
	ErrInvalidField          = Errorf(kmip.ResultReasonInvalidField, "Invalid field")
)

func handleMessageError(ctx context.Context, req *kmip.RequestMessage, err error) *kmip.ResponseMessage {
	header := kmip.ResponseHeader{
		ProtocolVersion: kmip.V1_0,
		TimeStamp:       time.Now(),
		BatchCount:      1,
	}
	if req != nil {
		if req.Header.ProtocolVersion != (kmip.ProtocolVersion{}) {
			header.ProtocolVersion = req.Header.ProtocolVersion
		}
		header.ClientCorrelationValue = req.Header.ClientCorrelationValue
		header.ServerCorrelationValue = req.Header.ServerCorrelationValue
	}

	bi := kmip.ResponseBatchItem{}
	handleBatchItemError(ctx, &bi, err)

	return &kmip.ResponseMessage{
		Header:    header,
		BatchItem: []kmip.ResponseBatchItem{bi},
	}
}

func handleBatchItemError(ctx context.Context, bi *kmip.ResponseBatchItem, err error) {
	if err == nil {
		return
	}
	// Always clear the ID placeholder on error
	//TODO: Double check against the KMIP specification about this
	ClearIdPlaceholder(ctx)
	bi.ResultStatus = kmip.ResultStatusOperationFailed
	var e Error
	if errors.As(err, &e) {
		bi.ResultReason = e.Reason
	} else {
		bi.ResultReason = kmip.ResultReasonGeneralFailure
	}

	if errors.Is(err, ErrOperationNotSupported) ||
		errors.Is(err, ErrFeatureNotSupported) ||
		errors.Is(err, ErrMissingData) ||
		errors.Is(err, ErrItemNotFound) ||
		errors.Is(err, ErrPermissionDenied) ||
		errors.Is(err, ErrInvalidMessage) ||
		errors.Is(err, ErrInvalidField) {
		bi.ResultMessage = err.Error()
	} else {
		slog.Error("KMIP Batch Item Error", "error", err)
	}
}
