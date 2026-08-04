//go:build darwin

package atlas

import (
	"syscall"
	"unsafe"
)

const aOMissionDarwinOpenat = 463

func openAOMissionUnixAt(dirfd int, path string, flags int) (int, error) {
	pointer, err := syscall.BytePtrFromString(path)
	if err != nil {
		return -1, err
	}
	result, _, errno := syscall.Syscall6(aOMissionDarwinOpenat, uintptr(dirfd), uintptr(unsafe.Pointer(pointer)), uintptr(flags), 0, 0, 0)
	if errno != 0 {
		return -1, errno
	}
	return int(result), nil
}
