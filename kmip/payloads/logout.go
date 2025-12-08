package payloads

import (
	"github.com/openkcm/crypto/kmip"
)

func init() {
	kmip.RegisterOperationPayload[LogoutRequestPayload, LogoutResponsePayload](kmip.OperationLogout)
}

var _ kmip.OperationPayload = (*LogoutRequestPayload)(nil)

type LogoutRequestPayload struct {
	UniqueIdentifier string `ttlv:",omitempty"`
}

// Operation implements kmip.OperationPayload.
func (a *LogoutRequestPayload) Operation() kmip.Operation {
	return kmip.OperationLogout
}

type LogoutResponsePayload struct {
	// The Unique Identifier of the object.
	UniqueIdentifier string
}

var _ kmip.OperationPayload = (*LogoutResponsePayload)(nil)

// Operation implements kmip.OperationPayload.
func (a *LogoutResponsePayload) Operation() kmip.Operation {
	return kmip.OperationLogout
}
