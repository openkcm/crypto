package kmip

import (
	"strconv"
	"strings"

	"github.com/openkcm/crypto/kmip/ttlv"
)

// init registers the bitmask string representations for CryptographicUsageMask and StorageStatusMask
// with the KMIP TTLV package. This enables human-readable string formatting and parsing for these bitmask types.
func init() {

	ttlv.RegisterBitmask[CryptographicUsageMask](
		TagCryptographicUsageMask,
		"Sign",
		"Verify",
		"Encrypt",
		"Decrypt",
		"WrapKey",
		"UnwrapKey",
		"Export",
		"MACGenerate",
		"MACVerify",
		"DeriveKey",
		"ContentCommitment",
		"KeyAgreement",
		"CertificateSign",
		"CRLSign",
		"GenerateCryptogram",
		"ValidateCryptogram",
		"TranslateEncrypt",
		"TranslateDecrypt",
		"TranslateWrap",
		"TranslateUnwrap",
		"Authenticate",
		"Unrestricted",
		"FPEEncrypt",
		"FPEDecrypt",
	)

	ttlv.RegisterBitmask[StorageStatusMask](
		TagStorageStatusMask,
		"OnLineStorage",
		"ArchivalStorage",
	)

	ttlv.RegisterBitmask[ProtectionStorageMask](
		TagProtectionStorageMask,
		"Software",
		"Hardware",
		"OnProcessor",
		"OnSystem",
		"OffSystem",
		"Hypervisor",
		"OperatingSystem",
		"Container",
		"OnPremises",
		"OffPremises",
		"SelfManaged",
		"Outsourced",
		"Validated",
		"SameJurisdiction",
	)
}

// CryptographicUsageMask represents a set of bitmask flags indicating the permitted cryptographic operations
// that can be performed with a cryptographic object, such as encrypt, decrypt, sign, or verify.
// Each bit in the mask corresponds to a specific usage permission as defined by the KMIP specification.
// This type is used to restrict or allow certain cryptographic operations on keys and other objects.
type CryptographicUsageMask int32

const (
	CryptographicUsageSign CryptographicUsageMask = 1 << iota
	CryptographicUsageVerify
	CryptographicUsageEncrypt
	CryptographicUsageDecrypt
	CryptographicUsageWrapKey
	CryptographicUsageUnwrapKey
	CryptographicUsageExport
	CryptographicUsageMACGenerate
	CryptographicUsageMACVerify
	CryptographicUsageDeriveKey
	CryptographicUsageContentCommitment
	CryptographicUsageKeyAgreement
	CryptographicUsageCertificateSign
	CryptographicUsageCRLSign
	CryptographicUsageGenerateCryptogram
	CryptographicUsageValidateCryptogram
	CryptographicUsageTranslateEncrypt
	CryptographicUsageTranslateDecrypt
	CryptographicUsageTranslateWrap
	CryptographicUsageTranslateUnwrap
	// KMIP 2.0
	CryptographicUsageAuthenticate
	CryptographicUsageUnrestricted
	CryptographicUsageFPEEncrypt
	CryptographicUsageFPEDecrypt
)

// MarshalText returns a human-readable string representation of the CryptographicUsageMask.
// The string is a bitwise OR ("|") separated list of enabled usage flags.
// This method never returns an error.
func (mask CryptographicUsageMask) MarshalText() ([]byte, error) {
	return []byte(ttlv.BitmaskStr(mask, " | ")), nil
}

func (mask *CryptographicUsageMask) UnmarshalText(text []byte) error {
	return maskUnmarshalText(mask, TagCryptographicUsageMask, string(text))
}

// StorageStatusMask represents a bitmask for storage status flags.
// It is used to indicate various storage states using bitwise operations.
// Each bit corresponds to a specific storage status as defined by the KMIP specification.
type StorageStatusMask int32

const (
	// StorageStatusOnlineStorage indicates the object is in online storage.
	StorageStatusOnlineStorage StorageStatusMask = 1 << iota
	// StorageStatusArchivalStorage indicates the object is in archival storage.
	StorageStatusArchivalStorage

	// KMIP 2.0
	StorageStatusDestroyedStorage
)

// MarshalText returns a human-readable string representation of the StorageStatusMask.
// The string is a bitwise OR ("|") separated list of enabled storage status flags.
// This method never returns an error.
func (mask StorageStatusMask) MarshalText() ([]byte, error) {
	return []byte(ttlv.BitmaskStr(mask, " | ")), nil
}

func (mask *StorageStatusMask) UnmarshalText(text []byte) error {
	return maskUnmarshalText(mask, TagStorageStatusMask, string(text))
}

type ProtectionStorageMask int32

const (
	ProtectionStorageSoftware ProtectionStorageMask = 1 << iota
	ProtectionStorageHardware
	ProtectionStorageOnProcessor
	ProtectionStorageOnSystem
	ProtectionStorageOffSystem
	ProtectionStorageHypervisor
	ProtectionStorageOperatingSystem
	ProtectionStorageContainer
	ProtectionStorageOnPremises
	ProtectionStorageOffPremises
	ProtectionStorageSelfManaged
	ProtectionStorageOutsourced
	ProtectionStorageValidated
	ProtectionStorageSameJurisdiction
)

func (mask ProtectionStorageMask) MarshalText() ([]byte, error) {
	return []byte(ttlv.BitmaskStr(mask, " | ")), nil
}

func (mask *ProtectionStorageMask) UnmarshalText(text []byte) error {
	return maskUnmarshalText(mask, TagProtectionStorageMask, string(text))
}

func maskUnmarshalText[T ~int32](mask *T, tag int, text string) error {
	var parts []string
	if strings.ContainsRune(text, '|') {
		parts = strings.Split(text, "|")
	} else {
		parts = strings.Fields(text)
	}

	*mask = 0
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		var parsed int64
		var err error
		if strings.HasPrefix(part, "0x") || strings.HasPrefix(part, "0X") {
			parsed, err = strconv.ParseInt(part[2:], 16, 32)
		} else {
			parsed, err = strconv.ParseInt(part, 10, 32)
			if err != nil {
				// Look for the name
				var p int32
				p, err = ttlv.BitmaskByStr(tag, part)
				parsed = int64(p)
			}
		}
		if err != nil {
			return err
		}
		*mask |= T(parsed)
	}
	return nil
}
