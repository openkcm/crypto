package operations

import (
	"context"

	"github.com/openkcm/crypto/internal/core"
	"github.com/openkcm/crypto/kmip"
	"github.com/openkcm/crypto/kmip/payloads"
)

type create struct {
}

var _ Operation = (*create)(nil)

func (c *create) Operation() kmip.Operation {
	return kmip.OperationCreate
}

func (c *create) Execute(ctx context.Context, serviceRegistry core.ServiceRegistry) (kmip.OperationPayload, error) {
	return &payloads.CreateResponsePayload{}, nil
}
