package store

import "github.com/openkcm/krypton/pkg/auth"

type NoOp struct{}

var _ auth.Store = &NoOp{}

func (n *NoOp) Store(s *auth.Assertion) error {
	return nil
}

func (n *NoOp) Get() (*auth.Assertion, error) {
	return &auth.Assertion{}, nil
}

func (n *NoOp) Delete() error {
	return nil
}
