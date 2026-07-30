package atlas

import (
	"fmt"
	"os"
	"strings"
)

const maxRunLinkEvidenceBytes int64 = 16 << 20

func BuildEvidenceBoundRunLink(taskID, status string, evidence map[string]string, evidenceRoot, evidenceRootID string) (RunLink, error) {
	link := RunLink{
		ContractVersion: RunLinkContract,
		TaskID:          taskID,
		Status:          status,
		Evidence:        evidence,
		EvidenceDigests: map[string]string{},
		EvidenceRootID:  evidenceRootID,
	}
	if err := validateEvidenceRoot(evidenceRoot); err != nil {
		return RunLink{}, err
	}
	for key, relativePath := range evidence {
		digest, err := digestRunLinkEvidenceFile(evidenceRoot, relativePath)
		if err != nil {
			return RunLink{}, fmt.Errorf("evidence %s: %w", key, err)
		}
		link.EvidenceDigests[key] = digest
	}
	link.Digest = digestRunLink(link)
	if err := ValidateRunLink(link); err != nil {
		return RunLink{}, err
	}
	return link, nil
}

func VerifyRunLinkEvidence(link RunLink, evidenceRoot, evidenceRootID string) error {
	if err := ValidateRunLink(link); err != nil {
		return err
	}
	if len(link.EvidenceDigests) == 0 {
		return fmt.Errorf("evidence-bound run-link is required")
	}
	if evidenceRootID != link.EvidenceRootID {
		return fmt.Errorf("evidence root identity mismatch: got %q want %q", evidenceRootID, link.EvidenceRootID)
	}
	if err := validateEvidenceRoot(evidenceRoot); err != nil {
		return err
	}
	for key, relativePath := range link.Evidence {
		actual, err := digestRunLinkEvidenceFile(evidenceRoot, relativePath)
		if err != nil {
			return fmt.Errorf("evidence %s: %w", key, err)
		}
		expected := link.EvidenceDigests[key]
		if actual != expected {
			return fmt.Errorf("evidence digest mismatch for %s: got %s want %s", key, actual, expected)
		}
	}
	return nil
}

func validateEvidenceRoot(root string) error {
	if strings.TrimSpace(root) == "" {
		return fmt.Errorf("evidence root is required")
	}
	info, err := os.Lstat(root)
	if err != nil {
		return fmt.Errorf("evidence root: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return fmt.Errorf("evidence root must be a non-symlink directory")
	}
	return nil
}
