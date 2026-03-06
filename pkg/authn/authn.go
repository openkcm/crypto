// Package authn provides low-level authentication primitives and interfaces for verifying credentials, validating tokens, and managing authentication state.
// It offers a flexible and extensible framework that supports various authentication methods and storage mechanisms.
package authn

import (
	"context"
	"errors"
)

const (
	ValidStatus ValidationStatus = iota
	InvalidStatus
	ExpiredStatus
)

var (
	ErrTokenNotFound = errors.New("token not found")
	ErrTokenNil      = errors.New("token cannot be nil")
)

type (
	// Credentials represents something the claimant knows, has, or is that can be used to verify their identity, e.g. a password or a certificate.
	Credentials struct {
		// Type is a string that identifies the type of credentials.
		Type string
		// Value is the raw bytes of the credentials.
		Value []byte
	}

	// Token represents a statement from a provider to a relying party that a claimant is authenticated.
	Token struct {
		// Type is a string that identifies the type of token.
		Type string
		// Value is the raw bytes of the token.
		Value []byte
		// ExpiredAt is a Unix timestamp indicating when the token expires.
		ExpiredAt int64
		// Attributes is a map of additional information about the token.
		Attributes map[string]any
	}

	// ValidationResult represents the result of validating a token.
	ValidationResult struct {
		Status ValidationStatus
	}

	// ValidationStatus represents the status of validating an token.
	ValidationStatus int
)

// Provider is responsible for verifying credentials and validating an token.
type Provider interface {
	Verify(context.Context, *Credentials) (*Token, error)
	Validate(context.Context, *Token) (ValidationResult, error)
}

// Store is responsible for storing, retrieving and deleting token.
type Store interface {
	Store(context.Context, *Token) error
	Get(context.Context) (*Token, error)
	Delete(context.Context) error
}
