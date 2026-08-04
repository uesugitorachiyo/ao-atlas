//go:build darwin || linux

package atlas

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

func openAOMissionRegularFileNoFollow(path string) (*os.File, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	clean := filepath.Clean(abs)
	beforeAOMissionNoFollowFinalOpen(clean)
	fd, err := syscall.Open(clean, syscall.O_RDONLY|syscall.O_NONBLOCK|syscall.O_CLOEXEC|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(fd), clean)
	if file == nil {
		_ = syscall.Close(fd)
		return nil, fmt.Errorf("open regular input")
	}
	return file, nil
}

func openAOMissionRegularFileBeneathNoFollow(rootPath, relativePath string) (*os.File, error) {
	parts, err := aOMissionRelativePathParts(relativePath)
	if err != nil {
		return nil, err
	}
	fd, err := syscall.Open(rootPath, syscall.O_RDONLY|syscall.O_DIRECTORY|syscall.O_CLOEXEC|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return nil, err
	}
	for index, part := range parts {
		final := index == len(parts)-1
		flags := syscall.O_RDONLY | syscall.O_CLOEXEC | syscall.O_NOFOLLOW
		if final {
			flags |= syscall.O_NONBLOCK
			beforeAOMissionNoFollowFinalOpen(filepath.Join(rootPath, relativePath))
		} else {
			flags |= syscall.O_DIRECTORY
		}
		next, openErr := openAOMissionUnixAt(fd, part, flags)
		closeErr := syscall.Close(fd)
		if openErr != nil {
			return nil, fmt.Errorf("open retained path component %q: %w", part, openErr)
		}
		if closeErr != nil {
			_ = syscall.Close(next)
			return nil, closeErr
		}
		fd = next
	}
	file := os.NewFile(uintptr(fd), filepath.Join(rootPath, relativePath))
	if file == nil {
		_ = syscall.Close(fd)
		return nil, errors.New("open retained artifact")
	}
	return file, nil
}
