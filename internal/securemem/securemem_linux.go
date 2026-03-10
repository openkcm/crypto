package securemem

import (
	"errors"

	"golang.org/x/sys/unix"
)

func alloc(size int) ([]byte, error) {
	data, err := unix.Mmap(
		-1,                             // no file
		0,                              // offset
		size,                           // length
		unix.PROT_READ|unix.PROT_WRITE, // to have read and write permissions
		unix.MAP_ANON|unix.MAP_PRIVATE, // anonymous mapping, not backed by any file, and private to this process
	)
	if err != nil {
		return nil, err
	}

	err = unix.Madvise(data, unix.MADV_DONTDUMP)
	if err != nil {
		errUnmap := unix.Munmap(data)
		return nil, errors.Join(err, errUnmap)
	}

	err = unix.Mlock(data)
	if err != nil {
		errUnmap := unix.Munmap(data)
		return nil, errors.Join(err, errUnmap)
	}

	return data, nil
}
