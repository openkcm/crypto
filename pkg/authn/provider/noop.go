package provider

import (
	"context"

	"github.com/openkcm/krypton/pkg/authn"
)

type NoOp struct{}

var _ authn.Provider = &NoOp{}

func (n *NoOp) Verify(ctx context.Context, c *authn.Credentials) (*authn.Token, error) {
	return &authn.Token{}, nil
}

func (n *NoOp) Validate(ctx context.Context, s *authn.Token) (authn.ValidationResult, error) {
	return authn.ValidationResult{
		Status: authn.ValidStatus,
	}, nil
}
