//go:build windows

package atlas

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

const (
	aOMissionWindowsObjectCaseInsensitive = 0x00000040
	aOMissionWindowsFileOpen              = 0x00000001
	aOMissionWindowsDirectoryFile         = 0x00000001
	aOMissionWindowsNonDirectoryFile      = 0x00000040
	aOMissionWindowsSynchronousIONonAlert = 0x00000020
	aOMissionWindowsOpenReparsePoint      = 0x00200000
)

var aOMissionWindowsNtCreateFile = syscall.NewLazyDLL("ntdll.dll").NewProc("NtCreateFile")

type aOMissionWindowsUnicodeString struct {
	Length        uint16
	MaximumLength uint16
	Buffer        *uint16
}

type aOMissionWindowsObjectAttributes struct {
	Length                   uintptr
	RootDirectory            syscall.Handle
	ObjectName               *aOMissionWindowsUnicodeString
	Attributes               uintptr
	SecurityDescriptor       uintptr
	SecurityQualityOfService uintptr
}

type aOMissionWindowsIOStatusBlock struct {
	Status      uintptr
	Information uintptr
}

func openAOMissionRegularFileNoFollow(path string) (*os.File, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	clean := filepath.Clean(abs)
	beforeAOMissionNoFollowFinalOpen(clean)
	return openAOMissionWindowsPathNoFollow(clean, false)
}

func openAOMissionRegularFileBeneathNoFollow(rootPath, relativePath string) (*os.File, error) {
	parts, err := aOMissionRelativePathParts(relativePath)
	if err != nil {
		return nil, err
	}
	root, err := openAOMissionWindowsPathNoFollow(rootPath, true)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	parent := syscall.Handle(root.Fd())
	parents := []syscall.Handle{}
	defer func() {
		for index := len(parents) - 1; index >= 0; index-- {
			_ = syscall.CloseHandle(parents[index])
		}
	}()
	for index, part := range parts {
		final := index == len(parts)-1
		if final {
			beforeAOMissionNoFollowFinalOpen(filepath.Join(rootPath, relativePath))
		}
		handle, err := openAOMissionWindowsRelativeNoFollow(parent, part, final)
		if err != nil {
			return nil, err
		}
		if final {
			file := os.NewFile(uintptr(handle), filepath.Join(rootPath, relativePath))
			if file == nil {
				_ = syscall.CloseHandle(handle)
				return nil, errors.New("open retained artifact")
			}
			return file, nil
		}
		parents = append(parents, handle)
		parent = handle
	}
	return nil, errors.New("retained artifact path is required")
}

func openAOMissionWindowsPathNoFollow(path string, directory bool) (*os.File, error) {
	pointer, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return nil, err
	}
	flags := uint32(syscall.FILE_FLAG_OPEN_REPARSE_POINT)
	if directory {
		flags |= syscall.FILE_FLAG_BACKUP_SEMANTICS
	}
	handle, err := syscall.CreateFile(pointer, syscall.GENERIC_READ,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE,
		nil, syscall.OPEN_EXISTING, flags, 0)
	if err != nil {
		return nil, err
	}
	if err := validateAOMissionWindowsNoFollowHandle(handle, directory); err != nil {
		_ = syscall.CloseHandle(handle)
		return nil, err
	}
	file := os.NewFile(uintptr(handle), path)
	if file == nil {
		_ = syscall.CloseHandle(handle)
		return nil, errors.New("open regular input")
	}
	return file, nil
}

func openAOMissionWindowsRelativeNoFollow(parent syscall.Handle, name string, final bool) (syscall.Handle, error) {
	encoded, err := syscall.UTF16FromString(name)
	if err != nil {
		return 0, err
	}
	objectName := aOMissionWindowsUnicodeString{
		Length:        uint16((len(encoded) - 1) * 2),
		MaximumLength: uint16(len(encoded) * 2),
		Buffer:        &encoded[0],
	}
	attributes := aOMissionWindowsObjectAttributes{
		Length:        unsafe.Sizeof(aOMissionWindowsObjectAttributes{}),
		RootDirectory: parent,
		ObjectName:    &objectName,
		Attributes:    aOMissionWindowsObjectCaseInsensitive,
	}
	options := uintptr(aOMissionWindowsSynchronousIONonAlert | aOMissionWindowsOpenReparsePoint)
	if final {
		options |= aOMissionWindowsNonDirectoryFile
	} else {
		options |= aOMissionWindowsDirectoryFile
	}
	var handle syscall.Handle
	var status aOMissionWindowsIOStatusBlock
	result, _, _ := aOMissionWindowsNtCreateFile.Call(
		uintptr(unsafe.Pointer(&handle)),
		uintptr(syscall.GENERIC_READ|syscall.SYNCHRONIZE),
		uintptr(unsafe.Pointer(&attributes)),
		uintptr(unsafe.Pointer(&status)),
		0,
		uintptr(syscall.FILE_ATTRIBUTE_NORMAL),
		uintptr(syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE),
		aOMissionWindowsFileOpen,
		options,
		0,
		0,
	)
	if result != 0 {
		return 0, fmt.Errorf("NtCreateFile %q failed with status %#x", name, result)
	}
	if err := validateAOMissionWindowsNoFollowHandle(handle, !final); err != nil {
		_ = syscall.CloseHandle(handle)
		return 0, err
	}
	return handle, nil
}

func validateAOMissionWindowsNoFollowHandle(handle syscall.Handle, directory bool) error {
	var info syscall.ByHandleFileInformation
	if err := syscall.GetFileInformationByHandle(handle, &info); err != nil {
		return err
	}
	if info.FileAttributes&syscall.FILE_ATTRIBUTE_REPARSE_POINT != 0 {
		return errors.New("regular input path contains a reparse point")
	}
	if directory != (info.FileAttributes&syscall.FILE_ATTRIBUTE_DIRECTORY != 0) {
		return errors.New("regular input path has an unexpected file type")
	}
	return nil
}
