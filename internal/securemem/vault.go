package securemem

import "errors"

type Vault struct {
	data map[string]*VaultData
}

func NewVault() *Vault {
	return &Vault{
		data: make(map[string]*VaultData),
	}
}

var ErrDestroyAll = errors.New("failed to destroy all vault data")

func (v *Vault) Put(name string, data []byte) error {
	err := v.Destroy(name)
	if err != nil {
		return err
	}

	vaultData, err := NewVaultDataFrom(name, data)
	if err != nil {
		return err
	}

	v.data[name] = vaultData
	return nil
}

func (v *Vault) Reserve(name string, size int) ([]byte, error) {
	err := v.Destroy(name)
	if err != nil {
		return nil, err
	}

	vaultData, err := NewVaultData(name, size)
	if err != nil {
		return nil, err
	}

	v.data[name] = vaultData
	return vaultData.Data(), nil
}

func (v *Vault) Get(name string) ([]byte, bool) {
	vaultData, ok := v.data[name]
	if !ok {
		return nil, false
	}

	return vaultData.Data(), true
}

func (v *Vault) DestroyAll() error {
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

func (v *Vault) Destroy(name string) error {
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
