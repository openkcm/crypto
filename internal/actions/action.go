package actions

import (
	"context"

	"github.com/openkcm/crypto/kmip"
)

type Action interface {
	Execute(ctx context.Context) (kmip.OperationPayload, error)
}
