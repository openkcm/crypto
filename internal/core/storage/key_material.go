package storage

import (
	"github.com/openkcm/crypto/internal/config"
)

type memoryKeyMaterialStorage struct {
}

func NewMemoryKeyMaterialStorage(cfg *config.Config) KeyMaterialStorage {
	return &memoryKeyMaterialStorage{}
}
