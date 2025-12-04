package authorization

import (
	"crypto/x509"

	"github.com/openkcm/crypto/internal/config"
	"github.com/openkcm/crypto/kmip"
)

type certificateRequest struct {
	cfg *config.Config

	clientCertificates []*x509.Certificate
	operations         []kmip.Operation
}

func NewCertificateAuthorizationHandler(cfg *config.Config) AuthorizationHandler {
	return func(clientCertificates []*x509.Certificate, operations []kmip.Operation) Authorisation {
		return &certificateRequest{
			cfg:                cfg,
			clientCertificates: clientCertificates,
			operations:         operations,
		}
	}
}

func (req *certificateRequest) Check() *CheckResponse {
	return &CheckResponse{
		PerOperation: map[kmip.Operation]bool{},
		Result:       false,
	}
}
