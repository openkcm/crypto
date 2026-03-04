package provider

import "github.com/openkcm/krypton/pkg/auth"

type NoOp struct{}

var _ auth.Provider = &NoOp{}

func (n *NoOp) Validate(s *auth.Assertion) (auth.ValidationResult, error) {
	return auth.Expired, nil
}

func (n *NoOp) Verify(c auth.Authenticator) (*auth.Assertion, error) {
	return &auth.Assertion{}, nil
}

func (n *NoOp) Refresh(s *auth.Assertion) error {
	return nil
}
