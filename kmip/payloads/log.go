package payloads

import (
	"github.com/openkcm/crypto/kmip"
)

func init() {
	kmip.RegisterOperationPayload[LogRequestPayload, LogResponsePayload](kmip.OperationLog)
}

var _ kmip.OperationPayload = (*LogRequestPayload)(nil)

type LogRequestPayload struct {
	UniqueIdentifier string `ttlv:",omitempty"`
}

// Operation implements kmip.OperationPayload.
func (a *LogRequestPayload) Operation() kmip.Operation {
	return kmip.OperationLog
}

type LogResponsePayload struct {
	// The Unique Identifier of the object.
	UniqueIdentifier string
}

var _ kmip.OperationPayload = (*LogResponsePayload)(nil)

// Operation implements kmip.OperationPayload.
func (a *LogResponsePayload) Operation() kmip.Operation {
	return kmip.OperationLog
}
