package atlas

import (
	"fmt"
)

const maxRunLinkEvidenceBytes int64 = 16 << 20

type runLinkEvidenceRoot interface {
	Digest(relativePath string) (string, error)
	Close() error
}

func BuildEvidenceBoundRunLink(taskID, status string, evidence map[string]string, evidenceRoot, evidenceRootID string) (RunLink, error) {
	link := RunLink{
		ContractVersion: RunLinkContract,
		TaskID:          taskID,
		Status:          status,
		Evidence:        evidence,
		EvidenceDigests: map[string]string{},
		EvidenceRootID:  evidenceRootID,
	}
	root, err := openRunLinkEvidenceRoot(evidenceRoot)
	if err != nil {
		return RunLink{}, err
	}
	defer root.Close()
	for key, relativePath := range evidence {
		digest, err := root.Digest(relativePath)
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
	root, err := openRunLinkEvidenceRoot(evidenceRoot)
	if err != nil {
		return err
	}
	defer root.Close()
	for key, relativePath := range link.Evidence {
		actual, err := root.Digest(relativePath)
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
