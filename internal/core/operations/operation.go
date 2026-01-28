package operations

import (
	"context"

	"github.com/openkcm/krypton/internal/core"
	"github.com/openkcm/krypton/kmip"
)

type Operation interface {
	Operation() kmip.Operation
	Execute(ctx context.Context, serviceRegistry core.ServiceRegistry) (kmip.OperationPayload, error)
}
