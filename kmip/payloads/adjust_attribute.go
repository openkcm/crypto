package payloads

import (
	"github.com/openkcm/crypto/kmip"
)

func init() {
	kmip.RegisterOperationPayload[AdjustAttributeRequestPayload, AdjustAttributeResponsePayload](kmip.OperationAdjustAttribute)
}

var _ kmip.OperationPayload = (*AdjustAttributeRequestPayload)(nil)

type AdjustAttributeRequestPayload struct {
	UniqueIdentifier   string              `ttlv:",omitempty"`
	AttributeReference kmip.Attribute      `ttlv:",omitempty"`
	AdjustmentType     kmip.AdjustmentType `ttlv:",omitempty"`
	AdjustmentValue    string              `ttlv:",omitempty"`
}

// Operation implements kmip.OperationPayload.
func (a *AdjustAttributeRequestPayload) Operation() kmip.Operation {
	return kmip.OperationAdjustAttribute
}

type AdjustAttributeResponsePayload struct {
	// The Unique Identifier of the object.
	UniqueIdentifier string
}

var _ kmip.OperationPayload = (*AdjustAttributeResponsePayload)(nil)

// Operation implements kmip.OperationPayload.
func (a *AdjustAttributeResponsePayload) Operation() kmip.Operation {
	return kmip.OperationAdjustAttribute
}
