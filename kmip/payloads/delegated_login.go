package payloads

import (
	"time"

	"github.com/openkcm/crypto/kmip"
)

func init() {
	kmip.RegisterOperationPayload[DelegatedLoginRequestPayload, DelegatedLoginResponsePayload](kmip.OperationDelegatedLogin)
}

var _ kmip.OperationPayload = (*DelegatedLoginRequestPayload)(nil)

type DelegatedLoginRequestPayload struct {
	LeaseTime    *time.Duration    `ttlv:",omitempty"`
	RequestCount *uint32           `ttlv:",omitempty"`
	UsageLimits  *kmip.UsageLimits `ttlv:",omitempty"`
	Rights       []kmip.Right      `ttlv:",omitempty"`
}

// Operation implements kmip.OperationPayload.
func (a *DelegatedLoginRequestPayload) Operation() kmip.Operation {
	return kmip.OperationDelegatedLogin
}

type DelegatedLoginResponsePayload struct {
	// The Unique Identifier of the object.
	TicketTicket kmip.Ticket `ttlv:",omitempty"`
}

var _ kmip.OperationPayload = (*DelegatedLoginResponsePayload)(nil)

// Operation implements kmip.OperationPayload.
func (a *DelegatedLoginResponsePayload) Operation() kmip.Operation {
	return kmip.OperationDelegatedLogin
}
