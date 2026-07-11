package atlas

import (
	"path/filepath"
	"testing"
)

func TestMonth3FinalClosureAO2ApprovalIntegrityChecklist(t *testing.T) {
	path := filepath.Join(repoRoot(t), "docs", "evidence", "ao-m3-final-closure-v01", "nodes", "mission-recommendation-month3-final-closure-22-ao2-approval-integrity", "ao2-approval-integrity-checklist.json")
	checklist := mustLoadJSON[map[string]any](t, path)
	for _, field := range []string{
		"proposed_bytes_digest_required",
		"base_commit_required",
		"approval_identity_hardcoded",
		"auto_approval_allowed",
		"provider_execution_allowed",
		"rsi_remains_denied",
	} {
		if _, ok := checklist[field]; !ok {
			t.Fatalf("AO2 approval integrity checklist missing %s: %#v", field, checklist)
		}
	}
	if checklist["proposed_bytes_digest_required"] != true ||
		checklist["base_commit_required"] != true ||
		checklist["approval_identity_hardcoded"] != false ||
		checklist["auto_approval_allowed"] != false ||
		checklist["provider_execution_allowed"] != false ||
		checklist["claims_authority_advance"] != false ||
		checklist["rsi_remains_denied"] != true {
		t.Fatalf("AO2 approval integrity checklist lost digest/base-commit or safety boundary: %#v", checklist)
	}
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.implementation-evidence.v0.1")
	if err != nil {
		t.Fatalf("AO2 approval integrity checklist evidence rejected: %v", err)
	}
	if validator != "generic:schema-marker" {
		t.Fatalf("unexpected checklist validator: %s", validator)
	}
}
