package payloads

import (
	"github.com/openkcm/crypto/kmip"
)

func init() {
	kmip.RegisterOperationPayload[LoginRequestPayload, LoginResponsePayload](kmip.OperationLogin)
}

var _ kmip.OperationPayload = (*LoginRequestPayload)(nil)

type LoginRequestPayload struct {
	UniqueIdentifier string `ttlv:",omitempty"`
}

// Operation implements kmip.OperationPayload.
func (a *LoginRequestPayload) Operation() kmip.Operation {
	return kmip.OperationLogin
}

type LoginResponsePayload struct {
	// The Unique Identifier of the object.
	UniqueIdentifier string
}

var _ kmip.OperationPayload = (*LoginResponsePayload)(nil)

// Operation implements kmip.OperationPayload.
func (a *LoginResponsePayload) Operation() kmip.Operation {
	return kmip.OperationLogin
}
