package securemem

import (
	"errors"

	"golang.org/x/sys/unix"
)

func readonly(data []byte) error {
	return unix.Mprotect(data, unix.PROT_READ)
}

func readwrite(data []byte) error {
	return unix.Mprotect(data, unix.PROT_READ|unix.PROT_WRITE)
}

func unalloc(data []byte) error {
	zero(data)

	unlockErr := unix.Munlock(data)
	unmapErr := unix.Munmap(data)
	return errors.Join(unlockErr, unmapErr)
}

func zero(data []byte) {
	for i := range data {
		data[i] = 0
	}
}
