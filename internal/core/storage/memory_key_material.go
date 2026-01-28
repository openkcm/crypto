package storage

import (
	"context"
	"errors"

	"github.com/openkcm/krypton/internal/config"
)

type memoryKeyMaterialStorage struct {
}

var _ KeyMaterialStorage = (*memoryKeyMaterialStorage)(nil)

func NewMemoryKeyMaterialStorage(cfg *config.Config) KeyMaterialStorage {
	return &memoryKeyMaterialStorage{}
}

func (m memoryKeyMaterialStorage) Store(ctx context.Context, key KeyMaterial) error {
	return errors.New("store not implemented yet")
}
