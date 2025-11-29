package operations

import (
	kmip14spec "github.com/gemalto/kmip-go/kmip14"
	"github.com/openkcm/crypto/internal/kmip"
)

var (
	Registry = map[uint32]kmip.Operation{
		uint32(kmip14spec.OperationCreate): &createOperation{},
	}
)
