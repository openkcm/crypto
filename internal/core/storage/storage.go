package storage

import (
	"context"
)

type KeyMaterial struct {
}

type KeyMaterialStorage interface {
	Store(ctx context.Context, key KeyMaterial) error
}

type KeyMetadata struct {
}
type KeyMetadataStorage interface {
	Store(ctx context.Context, key KeyMetadata) error
}
