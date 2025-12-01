package config

import (
	"time"

	"github.com/openkcm/common-sdk/pkg/commoncfg"
)

// Config is the root configuration structure for the Crypto service.
// It embeds BaseConfig (logging, application, telemetry, etc.) and
// adds configuration blocks for the KMIP server.
type Config struct {
	commoncfg.BaseConfig `mapstructure:",squash" yaml:",inline" json:",inline"`

	KMIPServer KMIPServer `mapstructure:"kmip" yaml:"kmip" json:"kmip"`
}

// KMIPServer describes configuration for KMIP over TCP and HTTP endpoints.
type KMIPServer struct {
	TCP  KMIPTCP  `yaml:"tcp" json:"tcp"`
	HTTP KMIPHTTP `yaml:"http" json:"http"`
}

// KMIPOperation defines filtering rules that control which KMIP operations
// may be processed by the server.
//
// **Rule precedence**
//  1. If `Only` is non-empty → ONLY these operations are allowed.
//     (Exclude is ignored.)
//  2. Else if `Exclude` is non-empty → all operations except these are allowed.
//  3. If both are empty → all operations are allowed.
//
// Operation codes use KMIP numerical identifiers,
// e.g., 0x00000001 (Create), 0x00000002 (Destroy), etc.
type KMIPOperation struct {

	// Exclude lists KMIP operation codes that must be explicitly blocked.
	// Ignored if Only list is non-empty.
	Exclude []uint32 `mapstructure:"exclude"  yaml:"exclude" json:"exclude"`

	// Only restricts allowed operations exclusively to this list.
	// When non-empty, Exclude is ignored.
	Only []uint32 `mapstructure:"only" yaml:"only" json:"only"`
}

// KMIPTCP configures the TCP-based KMIP endpoint (default port: 5696).
// This is the standard KMIP transport with binary TTLV framing.
type KMIPTCP struct {
	// Enabled controls whether the KMIP TCP server is started.
	// If false, the TCP endpoint will not bind or listen.
	Enabled bool `yaml:"enabled" json:"enabled"`

	// Address defines the TCP bind address for the KMIP service.
	// It follows the standard Go "host:port" format.
	Address string `yaml:"address" json:"address" default:":5696"`

	// TLS contains the mTLS configuration (certificates, CA, keys)
	// used to secure KMIP-over-TCP. If nil, the server operates
	// in plaintext mode (not recommended for production).
	TLS *commoncfg.MTLS `yaml:"tls" json:"tls"`

	// KMIPOperation specifies which KMIP operations are allowed
	// or denied on this TCP endpoint, based on allowlist or
	// denylist semantics.
	KMIPOperation KMIPOperation `mapstructure:"operation" yaml:"" json:"operation"`
}

// KMIPHTTP configures the HTTP-based KMIP endpoint (optional).
// HTTP KMIP is a vendor extension and not part of the KMIP core standard,
// but is useful for proxying, debugging, or integrating with web systems.
type KMIPHTTP struct {
	// Enabled controls whether the KMIP TCP server is started.
	// If false, the TCP endpoint will not bind or listen.
	Enabled bool `yaml:"enabled" json:"enabled"`

	// Address defines the TCP bind address for the KMIP service.
	// It follows the standard Go "host:port" format.
	Address string `yaml:"address" json:"address" default:":8080"`

	// TLS contains the mTLS configuration (certificates, CA, keys)
	// used to secure KMIP-over-TCP. If nil, the server operates
	// in plaintext mode (not recommended for production).
	TLS *commoncfg.MTLS `yaml:"tls" json:"tls"`

	// KMIPOperation specifies which KMIP operations are allowed
	// or denied on this TCP endpoint, based on allowlist or
	// denylist semantics.
	KMIPOperation KMIPOperation `mapstructure:"operation" yaml:"operation" json:"operation"`

	BasePath string `yaml:"basePath" json:"basePath" default:"/kmip"`

	// ReadTimeout is the maximum time allowed to read the full request,
	// including the body. Zero or negative disables the timeout.
	//
	// Note: Most users should prefer ReadHeaderTimeout for better control.
	ReadTimeout time.Duration `yaml:"readTimeout" json:"readTimeout" default:"0s"`

	// ReadHeaderTimeout is the time allowed to read HTTP headers.
	// If zero, falls back to ReadTimeout.
	// Zero or negative disables the timeout.
	ReadHeaderTimeout time.Duration `yaml:"readHeaderTimeout" json:"readHeaderTimeout" default:"0s"`

	// WriteTimeout limits how long the server will wait while writing
	// a response. Zero or negative disables the timeout.
	WriteTimeout time.Duration `yaml:"writeTimeout" json:"writeTimeout" default:"0s"`

	// IdleTimeout is the maximum time to wait for the next request
	// on a keep-alive connection. Zero → ReadTimeout. Negative disables timeout.
	IdleTimeout time.Duration `yaml:"idleTimeout" json:"idleTimeout" default:"0s"`

	// MaxHeaderBytes sets the limit for request header size.
	// Zero → http.DefaultMaxHeaderBytes.
	MaxHeaderBytes int `yaml:"maxHeaderBytes" json:"maxHeaderBytes" default:"0"`
}
