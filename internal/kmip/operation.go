package kmip

import "context"

type Operation interface {
	Execute(ctx context.Context) error
}
