package operations

import (
	"context"

	"github.com/openkcm/krypton/internal/core"
	"github.com/openkcm/krypton/kmip"
	"github.com/openkcm/krypton/kmip/payloads"
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
