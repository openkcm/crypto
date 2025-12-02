package operations

import (
	"context"

	"github.com/openkcm/crypto/internal/core"
	"github.com/openkcm/crypto/kmip"
)

type Operation interface {
	Operation() kmip.Operation
	Execute(ctx context.Context, serviceRegistry core.ServiceRegistry) (kmip.OperationPayload, error)
}
