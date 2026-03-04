// Package auth defines low-level authentication primitives and interfaces for verifying, validating, storing, and injecting authentication assertions.
// It is designed to be flexible and extensible, allowing for a wide range of authentication methods and storage mechanisms.
// The core concepts and terminology are based on the NIST Digital Identity Guidelines (SP 800-63-3),
// which provide a framework for understanding and implementing authentication systems.
package auth

const (
	Valid ValidationResult = iota
	Invalid
	Expired
)

type (
	// Authenticator represents something the claimant knows, has, or is that can be used to verify their identity, e.g. a password or a certificate.
	Authenticator any

	// Assertion represents a statement from a provider to a relying party that a claimant is authenticated.
	Assertion struct {
		// Type is a string that identifies the type of assertion, e.g. basic auth.
		Type string
		// Value is the raw bytes of the assertion, e.g. the base64-encoded username and password for basic auth.
		Value []byte
		// ExpiredAt is a Unix timestamp indicating when the assertion expires.
		ExpiredAt int64
		// Attributes is a map of additional information about the assertion.
		Attributes map[string]any
	}

	// Request represents an outgoing request to a relying party that requires authentication.
	Request any

	// ValidationResult represents the result of validating an assertion.
	ValidationResult int
)

// Provider is responsible for verifying an authenticator and validating an assertion.
type Provider interface {
	Verify(Authenticator) (*Assertion, error)
	Validate(*Assertion) (ValidationResult, error)
}

// Store is responsible for storing, retrieving and deleting assertions.
type Store interface {
	Store(*Assertion) error
	Get() (*Assertion, error)
	Delete() error
}

// Injector is responsible for injecting an assertion into an outgoing request.
type Injector interface {
	Inject(*Request, *Assertion) error
}
