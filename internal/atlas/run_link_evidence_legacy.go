//go:build !go1.24

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
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", err
	}
	current := realRoot
	parts := strings.Split(clean, string(filepath.Separator))
	for index, part := range parts {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if err != nil {
			return "", err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("path component %q must not be a symlink", part)
		}
		if index < len(parts)-1 && !info.IsDir() {
			return "", fmt.Errorf("path component %q must be a directory", part)
		}
		if index == len(parts)-1 && !info.Mode().IsRegular() {
			return "", fmt.Errorf("evidence must be a regular file")
		}
	}
	file, err := os.Open(current)
	if err != nil {
		return "", err
	}
	openedInfo, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return "", err
	}
	expectedInfo, err := os.Lstat(current)
	if err != nil {
		_ = file.Close()
		return "", err
	}
	if !openedInfo.Mode().IsRegular() || !os.SameFile(expectedInfo, openedInfo) {
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
