// Package memvault provides secure in-memory storage for sensitive data.
// It ensures that data is locked in memory and wiped after use to prevent leakage.
package memvault

import (
	"errors"
	"fmt"
	"sync"

	"golang.org/x/sys/unix"
)

type MemVault struct {
	data []byte
	mux  sync.Mutex
}

var (
	ErrVaultWiped   = errors.New("vault is wiped")
	ErrInvalidInput = errors.New("invalid input: data cannot be nil or empty")
)

func NewWithSecret(data []byte) (*MemVault, error) {
	vault, err := NewWithCapacity(len(data))
	if err != nil {
		return nil, err
	}

	copy(vault.data, data)
	clearBytes(data)

	return vault, nil
}

func NewWithCapacity(size int) (*MemVault, error) {
	if size <= 0 {
		return nil, ErrInvalidInput
	}

	lockedBytes, err := initLockedMem(size)
	if err != nil {
		return nil, err
	}

	return &MemVault{data: lockedBytes}, nil
}

func (v *MemVault) Read(fn func(data []byte) error) error {
	v.mux.Lock()
	defer v.mux.Unlock()

	if len(v.data) == 0 {
		return ErrVaultWiped
	}
	return fn(v.data)
}

func (v *MemVault) Bytes() []byte {
	return v.data
}

func (v *MemVault) ReadAndWipe(fn func(data []byte) error) error {
	v.mux.Lock()
	defer v.mux.Unlock()

	if len(v.data) == 0 {
		return ErrVaultWiped
	}

	fnErr := fn(v.data)
	wipeErr := v.wipe()

	return errors.Join(fnErr, wipeErr)
}

func (v *MemVault) Wipe() error {
	v.mux.Lock()
	defer v.mux.Unlock()

	return v.wipe()
}

func (v *MemVault) wipe() error {
	if len(v.data) == 0 {
		return nil
	}

	defer func() {
		v.data = nil
	}()

	clearBytes(v.data)

	return freeLockedMem(v.data)
}

func initLockedMem(size int) ([]byte, error) {
	b, err := unix.Mmap(
		-1,                             // no file
		0,                              // offset
		size,                           // length
		unix.PROT_READ|unix.PROT_WRITE, // to have read and write permissions
		unix.MAP_ANON|unix.MAP_PRIVATE, // anonymous mapping, not backed by any file, and private to this process
	)
	if err != nil {
		return nil, fmt.Errorf("mmap: %w", err)
	}

	// this will prevent the memory from being swapped out to disk, which is important for sensitive data
	err = unix.Mlock(b) // lock the memory to prevent it from being swapped out
	if err != nil {
		_ = unix.Munmap(b) // unmap the memory if locking fails
		return nil, fmt.Errorf("mlock: %w", err)
	}

	return b, nil
}

func freeLockedMem(b []byte) error {
	// unlock the memory to allow it to be swapped out again
	err := unix.Munlock(b)
	if err != nil {
		return fmt.Errorf("munlock: %w", err)
	}

	err = unix.Munmap(b) // unmap the memory to free it
	if err != nil {
		return fmt.Errorf("munmap: %w", err)
	}

	return nil
}

func clearBytes(b []byte) {
	for i := range b {
		b[i] = 0 // zero out the original data to clear sensitive information
	}
}
