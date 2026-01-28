package authorization

import (
	"krypton/x509"

	"github.com/openkcm/krypton/kmip"
)

type Authorisation interface {
	Check() *CheckResponse
}

type CheckResponse struct {
	perOperation map[kmip.Operation]bool
}

func (resp *CheckResponse) Lookup(op kmip.Operation) (bool, bool) {
	value, found := resp.perOperation[op]
	return value, found
}

func (resp *CheckResponse) result() bool {
	if len(resp.perOperation) == 0 {
		return false
	}

	for _, ok := range resp.perOperation {
		if !ok {
			return false
		}
	}

	return true
}

func (resp *CheckResponse) Failed() bool {
	return !resp.result()
}

func (resp *CheckResponse) Succeeded() bool {
	return resp.result()
}

type AuthorizationHandler func([]*x509.Certificate, []kmip.Operation) Authorisation
