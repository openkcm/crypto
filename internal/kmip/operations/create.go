package operations

import (
	"context"

	kmip14spec "github.com/gemalto/kmip-go/kmip14"
	"github.com/openkcm/crypto/internal/kmip"
)

const (
	createTag       = kmip14spec.OperationCreate
	nativeCreateTag = uint32(createTag)
)

type createOperation struct {
}

var _ kmip.Operation = (*createOperation)(nil)

func (c *createOperation) Execute(ctx context.Context) error {
	return nil
}
