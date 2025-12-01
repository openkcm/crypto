package actions

import (
	"context"

	"github.com/openkcm/crypto/kmip"
)

type Action interface {
	Operation() kmip.Operation
	Execute(ctx context.Context) (kmip.OperationPayload, error)
}
