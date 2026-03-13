package securemem

import (
	"errors"

	"golang.org/x/sys/unix"
)

// alloc allocates a memory region of the specified size, locks it to prevent swapping, and returns a byte slice that references this memory.
// The memory is allocated using the mmap system call with the MAP_ANON and MAP_PRIVATE flags,
// which means it is not backed by any file and is private to the process.
// Additionally, it uses the MADV_DONTDUMP flag to prevent the memory from being included in core dumps, enhancing security.
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
