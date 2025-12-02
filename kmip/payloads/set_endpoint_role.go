package payloads

import (
	"github.com/openkcm/crypto/kmip"
)

func init() {
	kmip.RegisterOperationPayload[SetEndpointRoleRequestPayload, SetEndpointRoleResponsePayload](kmip.OperationSetEndpointRole)
}

var _ kmip.OperationPayload = (*SetEndpointRoleRequestPayload)(nil)

type SetEndpointRoleRequestPayload struct {
	UniqueIdentifier string `ttlv:",omitempty"`
}

// Operation implements kmip.OperationPayload.
func (a *SetEndpointRoleRequestPayload) Operation() kmip.Operation {
	return kmip.OperationSetEndpointRole
}

type SetEndpointRoleResponsePayload struct {
	// The Unique Identifier of the object.
	UniqueIdentifier string
}

var _ kmip.OperationPayload = (*SetEndpointRoleResponsePayload)(nil)

// Operation implements kmip.OperationPayload.
func (a *SetEndpointRoleResponsePayload) Operation() kmip.Operation {
	return kmip.OperationSetEndpointRole
}
