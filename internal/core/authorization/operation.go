package authorization

import (
	"crypto/x509"

	"github.com/openkcm/crypto/kmip"
)

type Authorisation interface {
	Check() *CheckResponse
}

type CheckResponse struct {
	PerOperation map[kmip.Operation]bool
	Result       bool
}

type AuthorizationHandler func([]*x509.Certificate, []kmip.Operation) Authorisation
