package storage

import (
	"context"
	"errors"

	"github.com/openkcm/krypton/internal/config"
)

type memoryKeyMetadataStorage struct {
}

var _ KeyMetadataStorage = (*memoryKeyMetadataStorage)(nil)

func NewMemoryKeyMetadataStorage(cfg *config.Config) KeyMetadataStorage {
	return &memoryKeyMetadataStorage{}
}

func (m memoryKeyMetadataStorage) Store(ctx context.Context, key KeyMetadata) error {
	return errors.New("store not implemented yet")
}
