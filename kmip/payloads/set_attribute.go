package payloads

import (
	"github.com/openkcm/crypto/kmip"
)

func init() {
	kmip.RegisterOperationPayload[SetAttributeRequestPayload, SetAttributeResponsePayload](kmip.OperationSetAttribute)
}

var _ kmip.OperationPayload = (*SetAttributeRequestPayload)(nil)

type SetAttributeRequestPayload struct {
	UniqueIdentifier string `ttlv:",omitempty"`
}

// Operation implements kmip.OperationPayload.
func (a *SetAttributeRequestPayload) Operation() kmip.Operation {
	return kmip.OperationSetAttribute
}

type SetAttributeResponsePayload struct {
	// The Unique Identifier of the object.
	UniqueIdentifier string
}

var _ kmip.OperationPayload = (*SetAttributeResponsePayload)(nil)

// Operation implements kmip.OperationPayload.
func (a *SetAttributeResponsePayload) Operation() kmip.Operation {
	return kmip.OperationSetAttribute
}
