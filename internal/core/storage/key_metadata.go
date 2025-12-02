package storage

import (
	"github.com/openkcm/crypto/internal/config"
)

type memoryKeyMetadataStorage struct {
}

func NewMemoryKeyMetadataStorage(cfg *config.Config) KeyMetadataStorage {
	return &memoryKeyMetadataStorage{}
}
