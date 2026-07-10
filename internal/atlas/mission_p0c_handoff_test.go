package atlas

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestP0CMissionFoundryHandoffPreservesRealCompletePathBoundaries(t *testing.T) {
	root := repoRoot(t)
	nodeDir := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01", "nodes", "mission-recommendation-p0b-contract-convergence-30")
	fixturePath := filepath.Join(nodeDir, "p0c-mission-foundry-handoff-check.json")
	promptPath := filepath.Join(nodeDir, "p0c-mission-foundry-handoff.md")
	fixture := mustLoadJSON[AtlasP0CMissionFoundryHandoffCheck](t, fixturePath)
	if err := ValidateAtlasP0CMissionFoundryHandoffCheck(fixture); err != nil {
		t.Fatal(err)
	}
	promptBytes, err := os.ReadFile(promptPath)
	if err != nil {
		t.Fatalf("read P0-C handoff prompt: %v", err)
	}
	prompt := string(promptBytes)
	for _, want := range []string{
		"Start a Codex goal and execute AO Mission supervised P0-C Mission-to-Foundry real complete-path readiness wave.",
		"Mission: mission-710327df54728420",
		"Minimum work budget: generate at least 30 bounded nodes and complete at least 20 before final response unless a true hard blocker remains.",
		"Use AO Mission as supervision owner and AO Atlas as workgraph owner.",
		"Keep exactly one executable mutation node active at a time.",
		"Do not use provider calls or credentials.",
		"Do not execute AO2 live mutation.",
		"RSI remains denied.",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("P0-C handoff prompt missing %q:\n%s", want, prompt)
		}
	}
	if fixture.CompletedNodesBefore != 29 ||
		fixture.ReadyNodesBefore != 1 ||
		fixture.MinimumGeneratedNodes != 30 ||
		fixture.MinimumCompletionBeforeFinalResponse != 20 ||
		!fixture.RequiresMissionFoundryCompletePath ||
		!fixture.RequiresSingleActiveNode ||
		fixture.PromotionRequested ||
		fixture.ClaimsAuthorityAdvance ||
		!fixture.RSIRemainsDenied {
		t.Fatalf("P0-C handoff check lost bounded continuation contract: %#v", fixture)
	}
}

func TestP0CMissionFoundryHandoffUsesTypedEvidenceValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "evidence", "ao-stack-p0b-contract-convergence-wave-v01", "nodes", "mission-recommendation-p0b-contract-convergence-30", "p0c-mission-foundry-handoff-check.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, AtlasP0CMissionFoundryHandoffCheckContract)
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:p0c-mission-foundry-handoff-check" {
		t.Fatalf("unexpected validator: %s", validator)
	}
}
