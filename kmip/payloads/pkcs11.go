package payloads

import (
	"github.com/openkcm/crypto/kmip"
)

func init() {
	kmip.RegisterOperationPayload[PKCS11RequestPayload, PKCS11ResponsePayload](kmip.OperationPKCS11)
}

var _ kmip.OperationPayload = (*PKCS11RequestPayload)(nil)

type PKCS11RequestPayload struct {
	UniqueIdentifier string `ttlv:",omitempty"`
}

// Operation implements kmip.OperationPayload.
func (a *PKCS11RequestPayload) Operation() kmip.Operation {
	return kmip.OperationPKCS11
}

type PKCS11ResponsePayload struct {
	// The Unique Identifier of the object.
	UniqueIdentifier string
}

var _ kmip.OperationPayload = (*PKCS11ResponsePayload)(nil)

// Operation implements kmip.OperationPayload.
func (a *PKCS11ResponsePayload) Operation() kmip.Operation {
	return kmip.OperationPKCS11
}
