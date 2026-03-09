// Demonstrates how to allocate memory that is excluded from Linux core dumps.
//
// Two protection layers are applied to sensitive allocations:
//   - MADV_DONTDUMP: instructs the kernel to skip this region when writing a
//     core file. This is the primary defence against secrets leaking via dumps.
//   - mlock: pins the pages in RAM so the kernel never writes them to swap.
//
// An unprotected allocation is kept alongside the secret so that the
// accompanying dump script can verify it can find plaintext secrets in general,
// and that only the protected region is hidden.
package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

// secureAlloc copies data into a freshly mmap'd region and applies
// MADV_DONTDUMP + mlock so the contents are excluded from core dumps and
// never paged to disk.
//
// The caller must release the memory with secureZeroAndFree when done;
// ordinary GC will not reclaim mmap'd memory.
func secureAlloc(data []byte) []byte {
	size := len(data)

	// mmap gives us a page-aligned, anonymous (not backed by any file) region.
	// MAP_PRIVATE means writes are not visible to other processes.
	// Using mmap rather than a regular Go allocation lets us control the exact
	// virtual-memory region that we later pass to madvise/mlock.
	b, err := unix.Mmap(
		-1, 0, size,
		unix.PROT_READ|unix.PROT_WRITE,
		unix.MAP_PRIVATE|unix.MAP_ANONYMOUS,
	)
	if err != nil {
		panic(fmt.Sprintf("mmap failed: %v", err))
	}

	// MADV_DONTDUMP tells the kernel to exclude this VMA (virtual memory area)
	// from core dumps. This is the key protection: even if the process crashes
	// or is externally dumped, the contents of this region will not appear in
	// the core file.
	fmt.Println("applying MADV_DONTDUMP to the secret allocation")
	if err := unix.Madvise(b, unix.MADV_DONTDUMP); err != nil {
		panic(fmt.Sprintf("madvise(MADV_DONTDUMP) failed: %v", err))
	}

	// mlock pins the pages in physical RAM, preventing the kernel from paging
	// them out to swap. Without this, sensitive data could be written to the
	// swap partition and persist on disk after the process exits.
	// Requires CAP_IPC_LOCK or a sufficient RLIMIT_MEMLOCK; we treat failure as
	// non-fatal because MADV_DONTDUMP already covers the core-dump threat.
	if err := unix.Mlock(b); err != nil {
		fmt.Fprintf(os.Stderr, "warning: mlock failed (non-fatal): %v\n", err)
	}

	copy(b, data)
	return b
}

// secureZeroAndFree zeroes the region before unmapping it, ensuring the secret
// is not left readable in memory after the caller is done with it.
func secureZeroAndFree(b []byte) {
	for i := range b {
		b[i] = 0
	}

	// Read back one byte through an unsafe pointer. This creates an observable
	// side-effect that prevents the compiler from treating the loop above as a
	// dead store and optimising it away.
	_ = *(*byte)(unsafe.Pointer(&b[0]))

	if err := unix.Munmap(b); err != nil {
		fmt.Fprintf(os.Stderr, "munmap failed: %v\n", err)
	}
}

// keepInMem spins a goroutine that reads from b so the compiler and linker
// cannot prove the slice is unused and eliminate its backing allocation.
// This is only needed for the unprotected control allocation whose sole purpose
// is to be found in core dumps by the verification script.
func keepInMem(b []byte) {
	go func() {
		for {
			// The condition is never true at runtime; the loop just prevents
			// the slice from being optimised away entirely.
			if b[0] == 0xFF {
				panic("keepAlive: unreachable")
			}
		}
	}()
}

func main() {
	// Protected allocation: this secret should never appear in a core dump.
	secret := secureAlloc([]byte("THISISSECURE"))
	defer secureZeroAndFree(secret)

	// Unprotected allocation: ordinary Go heap memory. It exists solely as a
	// control to prove that the dump script is capable of finding secrets in
	// core files — only the mmap'd region above should be invisible.
	unprotected := []byte("thisIsInsecure")
	keepInMem(unprotected)

	fmt.Printf("PID %d: secret at %p (MADV_DONTDUMP + mlock), unprotected at %p\n",
		os.Getpid(), &secret[0], &unprotected[0])
	fmt.Println("Run the dump script with this PID, then Ctrl+C to exit.")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	fmt.Println("\nexiting — secret zeroed and freed.")
}
