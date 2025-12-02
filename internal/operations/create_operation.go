package operations

import (
	"context"

	"github.com/openkcm/crypto/internal/core"
	"github.com/openkcm/crypto/kmip"
	"github.com/openkcm/crypto/kmip/payloads"
)

type createAction struct {
}

var _ Operation = (*createAction)(nil)

func (c *createAction) Operation() kmip.Operation {
	return kmip.OperationCreate
}

func (c *createAction) Execute(ctx context.Context, serviceRegistry core.ServiceRegistry) (kmip.OperationPayload, error) {
	return &payloads.CreateResponsePayload{}, nil
}
