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

const maxRunLinkEvidenceBytes int64 = 16 << 20

func BuildEvidenceBoundRunLink(taskID, status string, evidence map[string]string, evidenceRoot string) (RunLink, error) {
	link := RunLink{
		ContractVersion: RunLinkContract,
		TaskID:          taskID,
		Status:          status,
		Evidence:        evidence,
		EvidenceDigests: map[string]string{},
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

func VerifyRunLinkEvidence(link RunLink, evidenceRoot string) error {
	if err := ValidateRunLink(link); err != nil {
		return err
	}
	if len(link.EvidenceDigests) == 0 {
		return fmt.Errorf("evidence-bound run-link is required")
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

func digestRunLinkEvidenceFile(root, relativePath string) (string, error) {
	clean := filepath.Clean(filepath.FromSlash(relativePath))
	if filepath.IsAbs(clean) || clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path must remain beneath evidence root")
	}
	current := root
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
		if index == len(parts)-1 {
			if !info.Mode().IsRegular() {
				return "", fmt.Errorf("evidence must be a regular file")
			}
			if info.Size() > maxRunLinkEvidenceBytes {
				return "", fmt.Errorf("evidence exceeds %d-byte limit", maxRunLinkEvidenceBytes)
			}
		}
	}
	file, err := os.Open(current)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	written, err := io.Copy(hash, io.LimitReader(file, maxRunLinkEvidenceBytes+1))
	if err != nil {
		return "", err
	}
	if written > maxRunLinkEvidenceBytes {
		return "", fmt.Errorf("evidence exceeds %d-byte limit", maxRunLinkEvidenceBytes)
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), nil
}
