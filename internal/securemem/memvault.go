package securemem

import "errors"

type MemVault struct {
	data map[string]*MemVaultData
}

func NewMemVault() *MemVault {
	return &MemVault{
		data: make(map[string]*MemVaultData),
	}
}

var ErrDestroyAll = errors.New("failed to destroy all vault data")

func (v *MemVault) Put(name string, data []byte) error {
	err := v.Destroy(name)
	if err != nil {
		return err
	}

	vaultData, err := NewMemVaultDataFrom(name, data)
	if err != nil {
		return err
	}

	v.data[name] = vaultData
	return nil
}

func (v *MemVault) Reserve(name string, size int) ([]byte, error) {
	err := v.Destroy(name)
	if err != nil {
		return nil, err
	}

	vaultData, err := NewMemVaultData(name, size)
	if err != nil {
		return nil, err
	}

	v.data[name] = vaultData
	return vaultData.Data(), nil
}

func (v *MemVault) Get(name string) ([]byte, bool) {
	vaultData, ok := v.data[name]
	if !ok {
		return nil, false
	}

	return vaultData.Data(), true
}

func (v *MemVault) DestroyAll() error {
	isError := false
	for name := range v.data {
		err := v.Destroy(name)
		if err != nil {
			isError = true
		}
	}
	if isError {
		return ErrDestroyAll
	}
	return nil
}

func (v *MemVault) Destroy(name string) error {
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
