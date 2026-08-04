package atlas

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// beforeAOMissionNoFollowFinalOpen provides a deterministic final-entry swap
// point for the no-follow regression test. It is a no-op outside tests.
var beforeAOMissionNoFollowFinalOpen = func(string) {}

func readAOMissionBoundedRegularFileNoFollow(path string, limit int) ([]byte, error) {
	file, err := openAOMissionRegularFileNoFollow(path)
	if err != nil {
		return nil, err
	}
	return readAOMissionBoundedRegularFile(file, limit)
}

func readAOMissionBoundedRegularFileBeneathNoFollow(rootPath, relativePath string, limit int) ([]byte, error) {
	file, err := openAOMissionRegularFileBeneathNoFollow(rootPath, relativePath)
	if err != nil {
		return nil, err
	}
	return readAOMissionBoundedRegularFile(file, limit)
}

func readAOMissionBoundedRegularFile(file *os.File, limit int) ([]byte, error) {
	info, statErr := file.Stat()
	if statErr != nil {
		_ = file.Close()
		return nil, fmt.Errorf("stat regular input: %w", statErr)
	}
	if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		_ = file.Close()
		return nil, errors.New("input must be a regular non-symlink file")
	}
	if info.Size() > int64(limit) {
		_ = file.Close()
		return nil, fmt.Errorf("input exceeds %d-byte size limit", limit)
	}
	body, readErr := io.ReadAll(io.LimitReader(file, int64(limit)+1))
	closeErr := file.Close()
	if readErr != nil {
		return nil, fmt.Errorf("read regular input: %w", readErr)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("close regular input: %w", closeErr)
	}
	if len(body) > limit {
		return nil, fmt.Errorf("input exceeds %d-byte size limit", limit)
	}
	return body, nil
}

func aOMissionRelativePathParts(path string) ([]string, error) {
	clean := filepath.Clean(path)
	if clean == "." || filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return nil, fmt.Errorf("retained artifact path escapes root: %q", path)
	}
	parts := strings.Split(clean, string(filepath.Separator))
	for _, part := range parts {
		if part == "" || part == "." || part == ".." {
			return nil, fmt.Errorf("invalid retained artifact path: %q", path)
		}
	}
	return parts, nil
}
