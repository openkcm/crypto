package kmip

import "context"

type Handler func(context.Context, []byte) ([]byte, error)
