package payloads

import (
	"github.com/openkcm/crypto/kmip"
)

func init() {
	kmip.RegisterOperationPayload[ReProvisionRequestPayload, ReProvisionResponsePayload](kmip.OperationReProvision)
}

var _ kmip.OperationPayload = (*ReProvisionRequestPayload)(nil)

type ReProvisionRequestPayload struct {
	UniqueIdentifier string `ttlv:",omitempty"`
}

// Operation implements kmip.OperationPayload.
func (a *ReProvisionRequestPayload) Operation() kmip.Operation {
	return kmip.OperationReProvision
}

type ReProvisionResponsePayload struct {
	// The Unique Identifier of the object.
	UniqueIdentifier string
}

var _ kmip.OperationPayload = (*ReProvisionResponsePayload)(nil)

// Operation implements kmip.OperationPayload.
func (a *ReProvisionResponsePayload) Operation() kmip.Operation {
	return kmip.OperationReProvision
}
