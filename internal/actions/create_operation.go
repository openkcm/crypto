package actions

import (
	"context"

	"github.com/openkcm/crypto/kmip"
	"github.com/openkcm/crypto/kmip/payloads"
)

type createAction struct {
}

var _ Action = (*createAction)(nil)

func (*createAction) Operation() kmip.Operation {
	return kmip.OperationCreate
}

func (c createAction) Execute(ctx context.Context) (kmip.OperationPayload, error) {
	return &payloads.CreateResponsePayload{}, nil
}
