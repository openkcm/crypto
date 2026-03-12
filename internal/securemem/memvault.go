package securemem

import (
	"errors"
	"log/slog"
	"sync"
)

type MemVault struct {
	mux  sync.RWMutex
	data map[string]*MemVaultData
}

func NewMemVault() *MemVault {
	return &MemVault{
		data: make(map[string]*MemVaultData),
	}
}

var (
	ErrDestroyAllFailed       = errors.New("failed to destroy all vault data")
	ErrVaultDataAlreadyExists = errors.New("vault data with the same name already exists")
	ErrMarkOnlyFailed         = errors.New("failed to mark all vault data as read-only")
)

func (v *MemVault) Reserve(name string, size int) ([]byte, error) {
	v.mux.Lock()
	defer v.mux.Unlock()

	_, ok := v.data[name]
	if ok {
		return nil, ErrVaultDataAlreadyExists
	}

	vaultData, err := NewMemVaultData(name, size)
	if err != nil {
		return nil, err
	}

	v.data[name] = vaultData
	return vaultData.Data(), nil
}

func (v *MemVault) Get(name string) ([]byte, bool) {
	v.mux.RLock()
	defer v.mux.RUnlock()

	vaultData, ok := v.data[name]
	if !ok {
		return nil, false
	}

	return vaultData.Data(), true
}

func (v *MemVault) DestroyAll() error {
	v.mux.Lock()
	defer v.mux.Unlock()

	isError := false
	for name, vaultData := range v.data {
		err := vaultData.Destroy()
		if err != nil {
			slog.Error("failed to destroy vault data for", "name", name, "error", err)
			isError = true
			continue
		}
		delete(v.data, name)
	}

	if isError {
		return ErrDestroyAllFailed
	}

	return nil
}

func (v *MemVault) Destroy(name string) error {
	v.mux.Lock()
	defer v.mux.Unlock()

	vaultData, ok := v.data[name]
	if !ok {
		return nil
	}
	err := vaultData.Destroy()
	if err != nil {
		return err
	}

	delete(v.data, name)
	return nil
}

func (v *MemVault) MarkAllReadOnly() error {
	v.mux.Lock()
	defer v.mux.Unlock()

	isError := false
	for name, vaultData := range v.data {
		err := vaultData.MarkReadOnly()
		if err != nil {
			slog.Error("failed to mark vault data as readonly for", "name", name, "error", err)
			isError = true
			continue
		}
	}
	if isError {
		return ErrMarkOnlyFailed
	}

	return nil
}
