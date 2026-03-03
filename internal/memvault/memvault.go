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

// New creates a new MemVault by copying the provided data into locked memory.
// The original data slice is securely cleared after copying.
// Returns a pointer to the MemVault or an error if input is invalid or memory allocation fails.
func New(data []byte) (*MemVault, error) {
	if len(data) == 0 {
		return nil, ErrInvalidInput
	}

	lockedBytes, err := initLockedMem(len(data))
	if err != nil {
		return nil, err
	}

	copy(lockedBytes, data)
	clearBytes(data)

	return &MemVault{data: lockedBytes}, nil
}

// Read provides read-only access to the vault's data by invoking the given function
// with the current data slice. If the vault has been wiped, it returns ErrVaultWiped.
// The provided function should not modify the data slice.
func (v *MemVault) Read(fn func(data []byte) error) error {
	v.mux.Lock()
	defer v.mux.Unlock()

	if len(v.data) == 0 {
		return ErrVaultWiped
	}
	return fn(v.data)
}

// ReadAndWipe provides access to the vault's data by invoking the given function
// with the current data slice, then securely wipes the data from memory.
// If the vault has already been wiped, it returns ErrVaultWiped.
// Returns a combined error from the provided function and the wipe operation, if any.
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

// initLockedMem allocates a slice of memory of the given size using mmap,
// and locks it into RAM to prevent it from being swapped to disk.
// Returns the allocated byte slice or an error if allocation or locking fails.
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

// freeLockedMem unlocks and frees a previously locked memory region.
// It first unlocks the memory to allow it to be swapped out, then unmaps it to release the memory.
// Returns an error if either operation fails.
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

// clearBytes securely zeroes out the contents of the given byte slice.
// This is used to clear sensitive information from memory.
func clearBytes(b []byte) {
	for i := range b {
		b[i] = 0 // zero out the original data to clear sensitive information
	}
}
