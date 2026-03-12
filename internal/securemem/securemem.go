package securemem

import (
	"errors"
	"runtime"

	"golang.org/x/sys/unix"
)

// Zero securely wipes the contents of the provided byte slice by setting all bytes to zero.
func Zero(data []byte) {
	for i := range data {
		data[i] = 0
	}
}

// readonly changes the memory protection of the provided byte slice to
// read-only, preventing any modifications to its contents.
func readonly(data []byte) error {
	return unix.Mprotect(data, unix.PROT_READ)
}

// readwrite changes the memory protection of the provided byte slice to
// allow both reading and writing, enabling modifications to its contents.
func readwrite(data []byte) error {
	return unix.Mprotect(data, unix.PROT_READ|unix.PROT_WRITE)
}

// unalloc securely deallocates the memory associated with the provided byte slice by first zeroing its contents,
// then unlocking it to allow swapping, and finally unmapping it from the process's address space.
// The function ensures that the memory is securely wiped before being released, and it returns any
// errors encountered during the unlocking and unmapping processes.
func unalloc(data []byte) error {
	Zero(data)
	runtime.KeepAlive(data)

	unlockErr := unix.Munlock(data)
	unmapErr := unix.Munmap(data)
	return errors.Join(unlockErr, unmapErr)
}
