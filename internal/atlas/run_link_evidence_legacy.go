//go:build !go1.24

package atlas

import "fmt"

func openRunLinkEvidenceRoot(string) (runLinkEvidenceRoot, error) {
	return nil, fmt.Errorf("evidence-bound run links require Go 1.24 or later")
}
