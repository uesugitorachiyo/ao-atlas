package atlas

import (
	"errors"
	"strings"
)

func validateBlueprintImportInputs(paths BlueprintImportPaths) error {
	if strings.TrimSpace(paths.OutDir) == "" {
		return errors.New("--out is required")
	}
	return nil
}
