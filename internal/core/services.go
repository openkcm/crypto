package core

import (
	"github.com/openkcm/crypto/internal/config"
	"github.com/openkcm/crypto/internal/core/storage"
)

type ServiceRegistry interface {
	KeyMaterialStorage() storage.KeyMaterialStorage
	KeyMetadataStorage() storage.KeyMetadataStorage
}

type serviceRegistry struct {
	cfg *config.Config

	keyMaterialStorage storage.KeyMaterialStorage
	keyMetadataStorage storage.KeyMetadataStorage
}

func NewServiceRegistry(cfg *config.Config) ServiceRegistry {
	return &serviceRegistry{
		cfg: cfg,

		keyMaterialStorage: storage.NewMemoryKeyMaterialStorage(cfg),
		keyMetadataStorage: storage.NewMemoryKeyMetadataStorage(cfg),
	}
}

func (s *serviceRegistry) KeyMaterialStorage() storage.KeyMaterialStorage {
	return s.keyMaterialStorage
}

func (s *serviceRegistry) KeyMetadataStorage() storage.KeyMetadataStorage {
	return s.keyMetadataStorage
}
