package securemem

import "errors"

type MemVaultData struct {
	name       string
	data       []byte
	isReadOnly bool
}

var ErrInvalidSize = errors.New("invalid size: must be greater than 0")

func NewMemVaultDataFrom(name string, data []byte) (*MemVaultData, error) {
	vault, err := NewMemVaultData(name, len(data))
	if err != nil {
		return nil, err
	}

	copy(vault.data, data)

	return vault, nil
}

func NewMemVaultData(name string, size int) (*MemVaultData, error) {
	if size <= 0 {
		return nil, ErrInvalidSize
	}

	aBytes, err := alloc(size)
	if err != nil {
		return nil, err
	}

	return &MemVaultData{
		name: name,
		data: aBytes,
	}, nil
}

func (m *MemVaultData) Data() []byte {
	if m.data == nil {
		return nil
	}

	return m.data
}

func (m *MemVaultData) Destroy() error {
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

func (m *MemVaultData) Readonly() error {
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

func (m *MemVaultData) Name() string {
	return m.name
}
