package securemem

import "errors"

type VaultData struct {
	name       string
	data       []byte
	isReadOnly bool
}

var ErrInvalidSize = errors.New("invalid size: must be greater than 0")

func NewVaultDataFrom(name string, data []byte) (*VaultData, error) {
	vault, err := NewVaultData(name, len(data))
	if err != nil {
		return nil, err
	}

	copy(vault.data, data)

	return vault, nil
}

func NewVaultData(name string, size int) (*VaultData, error) {
	if size <= 0 {
		return nil, ErrInvalidSize
	}

	aBytes, err := alloc(size)
	if err != nil {
		return nil, err
	}

	return &VaultData{
		name: name,
		data: aBytes,
	}, nil
}

func (m *VaultData) Data() []byte {
	if m.data == nil {
		return nil
	}

	return m.data
}

func (m *VaultData) Destroy() error {
	if m.data == nil {
		return nil
	}

	defer func() {
		m.data = nil
	}()

	if m.isReadOnly {
		err := readwrite(m.data)
		if err != nil {
			return err
		}
		m.isReadOnly = false
	}

	return unalloc(m.data)
}

func (m *VaultData) Readonly() error {
	if m.data == nil {
		return nil
	}
	if m.isReadOnly {
		return nil
	}

	err := readonly(m.data)
	m.isReadOnly = err == nil
	return err
}

func (m *VaultData) Name() string {
	return m.name
}
