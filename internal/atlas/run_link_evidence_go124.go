//go:build go1.24

package atlas

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func digestRunLinkEvidenceFile(root, relativePath string) (string, error) {
	clean := filepath.Clean(filepath.FromSlash(relativePath))
	if filepath.IsAbs(clean) || clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path must remain beneath evidence root")
	}
	rootHandle, err := os.OpenRoot(root)
	if err != nil {
		return "", err
	}
	defer rootHandle.Close()
	current := rootHandle
	openedRoots := []*os.Root{}
	defer func() {
		for index := len(openedRoots) - 1; index >= 0; index-- {
			_ = openedRoots[index].Close()
		}
	}()
	parts := strings.Split(clean, string(filepath.Separator))
	for index, part := range parts {
		info, err := current.Lstat(part)
		if err != nil {
			return "", err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("path component %q must not be a symlink", part)
		}
		if index < len(parts)-1 {
			if !info.IsDir() {
				return "", fmt.Errorf("path component %q must be a directory", part)
			}
			next, err := current.OpenRoot(part)
			if err != nil {
				return "", err
			}
			openedInfo, err := next.Stat(".")
			if err != nil {
				_ = next.Close()
				return "", err
			}
			if !os.SameFile(info, openedInfo) {
				_ = next.Close()
				return "", fmt.Errorf("path component %q changed while opening", part)
			}
			openedRoots = append(openedRoots, next)
			current = next
			continue
		}
		if !info.Mode().IsRegular() {
			return "", fmt.Errorf("evidence must be a regular file")
		}
		if info.Size() > maxRunLinkEvidenceBytes {
			return "", fmt.Errorf("evidence exceeds %d-byte limit", maxRunLinkEvidenceBytes)
		}
		file, err := current.Open(part)
		if err != nil {
			return "", err
		}
		openedInfo, err := file.Stat()
		if err != nil {
			_ = file.Close()
			return "", err
		}
		if !openedInfo.Mode().IsRegular() || !os.SameFile(info, openedInfo) {
			_ = file.Close()
			return "", fmt.Errorf("evidence changed while opening")
		}
		hash := sha256.New()
		written, copyErr := io.Copy(hash, io.LimitReader(file, maxRunLinkEvidenceBytes+1))
		closeErr := file.Close()
		if copyErr != nil {
			return "", copyErr
		}
		if closeErr != nil {
			return "", closeErr
		}
		if written > maxRunLinkEvidenceBytes {
			return "", fmt.Errorf("evidence exceeds %d-byte limit", maxRunLinkEvidenceBytes)
		}
		return "sha256:" + hex.EncodeToString(hash.Sum(nil)), nil
	}
	return "", fmt.Errorf("evidence path is empty")
}
