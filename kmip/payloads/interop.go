package payloads

import (
	"github.com/openkcm/crypto/kmip"
)

func init() {
	kmip.RegisterOperationPayload[InteropRequestPayload, InteropResponsePayload](kmip.OperationInterop)
}

var _ kmip.OperationPayload = (*InteropRequestPayload)(nil)

type InteropRequestPayload struct {
	UniqueIdentifier string `ttlv:",omitempty"`
}

// Operation implements kmip.OperationPayload.
func (a *InteropRequestPayload) Operation() kmip.Operation {
	return kmip.OperationInterop
}

type InteropResponsePayload struct {
	// The Unique Identifier of the object.
	UniqueIdentifier string
}

var _ kmip.OperationPayload = (*InteropResponsePayload)(nil)

// Operation implements kmip.OperationPayload.
func (a *InteropResponsePayload) Operation() kmip.Operation {
	return kmip.OperationInterop
}
