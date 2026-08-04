//go:build !darwin && !linux && !windows

package atlas

import (
	"errors"
	"os"
)

func openAOMissionRegularFileNoFollow(string) (*os.File, error) {
	return nil, errors.New("no-follow Mission artifact imports are unsupported on this platform")
}

func openAOMissionRegularFileBeneathNoFollow(string, string) (*os.File, error) {
	return nil, errors.New("no-follow Mission artifact imports are unsupported on this platform")
}
