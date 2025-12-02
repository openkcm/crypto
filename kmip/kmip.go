package kmip

import (
	"fmt"

	"github.com/openkcm/crypto/kmip/ttlv"
)

type ProtocolVersion struct {
	ProtocolVersionMajor int32
	ProtocolVersionMinor int32
}

func (v ProtocolVersion) Major() int {
	return int(v.ProtocolVersionMajor)
}

func (v ProtocolVersion) Minor() int {
	return int(v.ProtocolVersionMinor)
}

func (pv ProtocolVersion) String() string {
	return fmt.Sprintf("%d.%d", pv.ProtocolVersionMajor, pv.ProtocolVersionMinor)
}

type CredentialValue struct {
	UserPassword    *CredentialValueUserPassword
	Device          *CredentialValueDevice
	Attestation     *CredentialValueAttestation
	OneTimePassword *CredentialValueOneTimePassword
	HashedPassword  *CredentialValueHashedPassword
	Ticket          *CredentialValueTicket
}

func (cred *CredentialValue) TagEncodeTTLV(e *ttlv.Encoder, tag int) {
	e.TagAny(tag, cred.UserPassword)
	e.TagAny(tag, cred.Device)
	e.TagAny(tag, cred.Attestation)
	e.TagAny(tag, cred.OneTimePassword)
	e.TagAny(tag, cred.Ticket)
}

func (cred *CredentialValue) decode(d *ttlv.Decoder, tag int, cType CredentialType) error {
	switch cType {
	case CredentialTypeUsernameAndPassword:
		return d.TagAny(tag, &cred.UserPassword)
	case CredentialTypeDevice:
		return d.TagAny(tag, &cred.Device)
	case CredentialTypeAttestation:
		return d.TagAny(tag, &cred.Attestation)
	case CredentialTypeOneTimePassword:
		return d.TagAny(tag, &cred.OneTimePassword)
	case CredentialTypeHashedPassword:
		return d.TagAny(tag, &cred.HashedPassword)
	case CredentialTypeTicket:
		return d.TagAny(tag, &cred.Ticket)
	}
	return fmt.Errorf("Unsupported credential type %X", cType)
}

type CredentialValueUserPassword struct {
	Username string
	Password string `ttlv:",omitempty"`
}

type CredentialValueDevice struct {
	DeviceSerialNumber string `ttlv:",omitempty"`
	Password           string `ttlv:",omitempty"`
	DeviceIdentifier   string `ttlv:",omitempty"`
	NetworkIdentifier  string `ttlv:",omitempty"`
	MachineIdentifier  string `ttlv:",omitempty"`
	MediaIdentifier    string `ttlv:",omitempty"`
}

type CredentialValueAttestation struct {
	Nonce                  Nonce
	AttestationType        AttestationType
	AttestationMeasurement []byte `ttlv:",omitempty"`
	AttestationAssertion   []byte `ttlv:",omitempty"`
}

type CredentialValueOneTimePassword struct {
}

type CredentialValueHashedPassword struct {
}

type CredentialValueTicket struct {
}

type Credential struct {
	CredentialType  CredentialType
	CredentialValue CredentialValue
}

func (kb *Credential) TagDecodeTTLV(d *ttlv.Decoder, tag int) error {
	return d.Struct(tag, func(d *ttlv.Decoder) error {
		if err := d.Any(&kb.CredentialType); err != nil {
			return err
		}
		return kb.CredentialValue.decode(d, TagCredentialValue, kb.CredentialType)
	})
}

type Authentication struct {
	Credential Credential
	// Starting from KMIP 1.2, Credential can be repeated
	AdditionalCredential []Credential `ttlv:",version=v1.2.."`
}

type RevocationReason struct {
	RevocationReasonCode RevocationReasonCode `ttlv:",omitempty"`
	RevocationMessage    string               `ttlv:",omitempty"`
}

type MessageExtension struct {
	VendorIdentification string
	CriticalityIndicator bool
	VendorExtension      ttlv.Struct
}

type CryptographicParameters struct {
	BlockCipherMode  BlockCipherMode  `ttlv:",omitempty"`
	PaddingMethod    PaddingMethod    `ttlv:",omitempty"`
	HashingAlgorithm HashingAlgorithm `ttlv:",omitempty"`
	KeyRoleType      KeyRoleType      `ttlv:",omitempty"`

	DigitalSignatureAlgorithm DigitalSignatureAlgorithm `ttlv:",omitempty,version=v1.2.."`
	CryptographicAlgorithm    CryptographicAlgorithm    `ttlv:",omitempty,version=v1.2.."`
	RandomIV                  *bool                     `ttlv:",version=v1.2.."`
	IVLength                  int32                     `ttlv:",omitempty,version=v1.2.."`
	TagLength                 int32                     `ttlv:",omitempty,version=v1.2.."`
	FixedFieldLength          int32                     `ttlv:",omitempty,version=v1.2.."`
	InvocationFieldLength     int32                     `ttlv:",omitempty,version=v1.2.."`
	CounterLength             int32                     `ttlv:",omitempty,version=v1.2.."`
	InitialCounterValue       *int32                    `ttlv:",version=v1.2.."`

	SaltLength                    *int32           `ttlv:",version=v1.4.."`
	MaskGenerator                 MaskGenerator    `ttlv:",omitempty,version=v1.4.."`
	MaskGeneratorHashingAlgorithm HashingAlgorithm `ttlv:",omitempty,version=v1.4.."`
	PSource                       []byte           `ttlv:",omitempty,version=v1.4.."`
	TrailerField                  *int32           `ttlv:",version=v1.4.."`
}

type CryptographicDomainParameters struct {
	Qlength          int32            `ttlv:",omitempty"`
	RecommendedCurve RecommendedCurve `ttlv:",omitempty"`
}

type KeyWrappingSpecification struct {
	WrappingMethod             WrappingMethod
	EncryptionKeyInformation   *EncryptionKeyInformation
	MACSignatureKeyInformation *MACSignatureKeyInformation
	AttributeName              []AttributeName
	EncodingOption             EncodingOption `ttlv:",omitempty,version=v1.1.."`
}

type Link struct {
	LinkType               LinkType `ttlv:",omitempty"`
	LinkedObjectIdentifier string   `ttlv:",omitempty"`
}

type Digest struct {
	HashingAlgorithm HashingAlgorithm
	DigestValue      []byte
	KeyFormatType    KeyFormatType `ttlv:",omitempty,version=1.1.."`
}

// Deprecated: deprecated as of kmip 1.1.
type CertificateIdentifier struct {
	Issuer       string `ttlv:",omitempty"`
	SerialNumber string `ttlv:",omitempty"`
}

// Deprecated: deprecated as of kmip 1.1.
type CertificateSubject struct {
	CertificateSubjectDistinguishedName string `ttlv:",omitempty"`
	CertificateSubjectAlternativeName   []string
}

// Deprecated: deprecated as of kmip 1.1.
type CertificateIssuer struct {
	CertificateIssuerDistinguishedName string `ttlv:",omitempty"`
	CertificateIssuerAlternativeName   []string
}

type ApplicationSpecificInformation struct {
	ApplicationNamespace string `ttlv:",omitempty"`
	ApplicationData      string `ttlv:",omitempty"` //TODO: Optional since kmip 1.3, not before.
}

type UsageLimits struct {
	UsageLimitsTotal int64
	UsageLimitsCount *int64
	UsageLimitsUnit  UsageLimitsUnit `ttlv:",omitempty"`
}

func (ul UsageLimits) Equals(other *UsageLimits) bool {
	return other != nil &&
		other.UsageLimitsTotal == ul.UsageLimitsTotal &&
		other.UsageLimitsUnit == ul.UsageLimitsUnit &&
		(other.UsageLimitsCount == nil && ul.UsageLimitsCount == nil ||
			*other.UsageLimitsCount == *ul.UsageLimitsCount)
}

// KMIP 1.1.

type ExtensionInformation struct {
	ExtensionName string
	ExtensionTag  int32 `ttlv:",omitempty"`
	ExtensionType int32 `ttlv:",omitempty"`
}

type X_509CertificateIdentifier struct {
	IssuerDistinguishedName []byte `ttlv:",omitempty"`
	CertificateSerialNumber []byte `ttlv:",omitempty"`
}

type X_509CertificateSubject struct {
	SubjectDistinguishedName []byte `ttlv:",omitempty"`
	SubjectAlternativeName   [][]byte
}

type X_509CertificateIssuer struct {
	IssuerDistinguishedName []byte `ttlv:",omitempty"`
	IssuerAlternativeName   [][]byte
}

// KMIP 1.2.

type Nonce struct {
	NonceID    []byte
	NonceValue []byte
}

type AlternativeName struct {
	AlternativeNameValue string              `ttlv:",omitempty"`
	AlternativeNameType  AlternativeNameType `ttlv:",omitempty"`
}

type KeyValueLocation struct {
	KeyValueLocationValue string               `ttlv:",omitempty"`
	KeyValueLocationType  KeyValueLocationType `ttlv:",omitempty"`
}

// KMIP 1.3.

type RNGParameters struct {
	RNGAlgorithm           RNGAlgorithm           `ttlv:",omitempty"`
	CryptographicAlgorithm CryptographicAlgorithm `ttlv:",omitempty"`
	CryptographicLength    int32                  `ttlv:",omitempty"`
	HashingAlgorithm       HashingAlgorithm       `ttlv:",omitempty"`
	DRBGAlgorithm          DRBGAlgorithm          `ttlv:",omitempty"`
	RecommendedCurve       RecommendedCurve       `ttlv:",omitempty"`
	FIPS186Variation       FIPS186Variation       `ttlv:",omitempty"`
	PredictionResistance   *bool
}

type ProfileInformation struct {
	ProfileName ProfileName
	ServerURI   string `ttlv:",omitempty"`
	ServerPort  int32  `ttlv:",omitempty"`
}

type ValidationInformation struct {
	ValidationAuthorityType         ValidationAuthorityType
	ValidationAuthorityCountry      string `ttlv:",omitempty"`
	ValidationAuthorityURI          string `ttlv:",omitempty"`
	ValidationVersionMajor          int32
	ValidationVersionMinor          *int32
	ValidationType                  ValidationType
	ValidationLevel                 int32
	ValidationCertificateIdentifier string `ttlv:",omitempty"`
	ValidationCertificateURI        string `ttlv:",omitempty"`
	ValidationVendorURI             string `ttlv:",omitempty"`
	ValidationProfile               []string
}

type CapabilityInformation struct {
	StreamingCapability     *bool
	AsynchronousCapability  *bool
	AttestationCapability   *bool
	BatchUndoCapability     *bool              `ttlv:",version=v1.4.."`
	BatchContinueCapability *bool              `ttlv:",version=v1.4.."`
	UnwrapMode              UnwrapMode         `ttlv:",omitempty"`
	DestroyAction           DestroyAction      `ttlv:",omitempty"`
	ShreddingAlgorithm      ShreddingAlgorithm `ttlv:",omitempty"`
	RNGMode                 RNGMode            `ttlv:",omitempty"`
}

type CertificateAttributes struct {
	CertificateSubjectCN           string `ttlv:",omitempty"`
	CertificateSubjectO            string `ttlv:",omitempty"`
	CertificateSubjectOU           string `ttlv:",omitempty"`
	CertificateSubjectEmail        string `ttlv:",omitempty"`
	CertificateSubjectC            string `ttlv:",omitempty"`
	CertificateSubjectST           string `ttlv:",omitempty"`
	CertificateSubjectL            string `ttlv:",omitempty"`
	CertificateSubjectUID          string `ttlv:",omitempty"`
	CertificateSubjectSerialNumber string `ttlv:",omitempty"`
	CertificateSubjectTitle        string `ttlv:",omitempty"`
	CertificateSubjectDC           string `ttlv:",omitempty"`
	CertificateSubjectDNQualifier  string `ttlv:",omitempty"`
	CertificateIssuerCN            string `ttlv:",omitempty"`
	CertificateIssuerO             string `ttlv:",omitempty"`
	CertificateIssuerOU            string `ttlv:",omitempty"`
	CertificateIssuerEmail         string `ttlv:",omitempty"`
	CertificateIssuerC             string `ttlv:",omitempty"`
	CertificateIssuerST            string `ttlv:",omitempty"`
	CertificateIssuerL             string `ttlv:",omitempty"`
	CertificateIssuerUID           string `ttlv:",omitempty"`
	CertificateIssuerSerialNumber  string `ttlv:",omitempty"`
	CertificateIssuerTitle         string `ttlv:",omitempty"`
	CertificateIssuerDC            string `ttlv:",omitempty"`
	CertificateIssuerDNQualifier   string `ttlv:",omitempty"`
}

type VendorAttribute struct {
	VendorIdentification string `ttlv:",omitempty"`
	AttributeName        string `ttlv:",omitempty"`
	AttributeValue       any    `ttlv:",omitempty"`
}

type Right struct {
	UsageLimits *UsageLimits `ttlv:",omitempty"`
	Operations  []Operation  `ttlv:",omitempty"`
	Objects     []Object     `ttlv:",omitempty"`
}

type Ticket struct {
	TicketType  TicketType `ttlv:",omitempty"`
	TicketValue []byte     `ttlv:",omitempty"`
}
