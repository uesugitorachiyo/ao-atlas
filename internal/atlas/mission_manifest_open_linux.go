//go:build linux

package atlas

import "syscall"

func openAOMissionUnixAt(dirfd int, path string, flags int) (int, error) {
	return syscall.Openat(dirfd, path, flags, 0)
}
