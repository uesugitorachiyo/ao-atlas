package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestFeatureDepthWaveFoundryTerminalStatusExamplesBindReadbackEnums(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v01")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-32")
	sourceReadbackPath := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-31", "recommendation-readback-after.json")
	recordedPath := filepath.Join(nodeDir, "foundry-terminal-status-examples.json")
	outPath := filepath.Join(t.TempDir(), "foundry-terminal-status-examples.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "foundry-terminal-status-examples",
		"--node-id", "mission-recommendation-feature-depth-next-wave-32",
		"--source-readback", sourceReadbackPath,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("foundry-terminal-status-examples command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=terminal_status_examples_validated") ||
		!strings.Contains(out.String(), "terminal_example_count=4") ||
		!strings.Contains(out.String(), "denied_example_count=3") {
		t.Fatalf("terminal status examples output missing validation state: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasFoundryTerminalStatusExamplesValidation](t, recordedPath)
	generated := mustLoadJSON[AtlasFoundryTerminalStatusExamplesValidation](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("Foundry terminal status examples fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasFoundryTerminalStatusExamplesValidation(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "terminal_status_examples_validated" ||
		recorded.TerminalStatusKeyCount != 4 ||
		recorded.TerminalExampleCount != 4 ||
		recorded.DeniedExampleCount != 3 ||
		!recorded.ExamplesMatchReadbackEnums ||
		!recorded.PromotedRequiresCommandPromoterAgreement ||
		!recorded.DeniedRequiresExactEvidence ||
		!recorded.BlockedRequiresRepairOrResume ||
		!recorded.DeniedExamplesSafe ||
		recorded.SchedulesWork ||
		recorded.ExecutesWork ||
		recorded.ApprovesWork ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("terminal status examples lost readback enum binding: %#v", recorded)
	}

	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, "ao.atlas.foundry-terminal-status-examples.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:foundry-terminal-status-examples" {
		t.Fatalf("expected typed Foundry terminal status examples validator, got %s", validator)
	}
}

func TestFeatureDepthWaveV02FoundryTerminalStatusExamplesBindReadbackEnums(t *testing.T) {
	root := repoRoot(t)
	waveRoot := filepath.Join(root, "docs", "evidence", "ao-atlas-feature-depth-wave-v02")
	nodeDir := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-32")
	sourceReadbackPath := filepath.Join(waveRoot, "nodes", "mission-recommendation-feature-depth-next-wave-31", "recommendation-readback-after.json")
	recordedPath := filepath.Join(nodeDir, "foundry-terminal-status-examples.json")
	outPath := filepath.Join(t.TempDir(), "foundry-terminal-status-examples.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "foundry-terminal-status-examples",
		"--node-id", "mission-recommendation-feature-depth-next-wave-32",
		"--source-readback", sourceReadbackPath,
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("foundry-terminal-status-examples command failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=terminal_status_examples_validated") ||
		!strings.Contains(out.String(), "terminal_example_count=4") ||
		!strings.Contains(out.String(), "denied_example_count=3") {
		t.Fatalf("terminal status examples output missing validation state: %s", out.String())
	}
	recorded := mustLoadJSON[AtlasFoundryTerminalStatusExamplesValidation](t, recordedPath)
	generated := mustLoadJSON[AtlasFoundryTerminalStatusExamplesValidation](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("v02 Foundry terminal status examples fixture changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if err := ValidateAtlasFoundryTerminalStatusExamplesValidation(recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Status != "terminal_status_examples_validated" ||
		recorded.TerminalStatusKeyCount != 4 ||
		recorded.TerminalExampleCount != 4 ||
		recorded.DeniedExampleCount != 3 ||
		!recorded.ExamplesMatchReadbackEnums ||
		!recorded.PromotedRequiresCommandPromoterAgreement ||
		!recorded.DeniedRequiresExactEvidence ||
		!recorded.BlockedRequiresRepairOrResume ||
		!recorded.DeniedExamplesSafe ||
		recorded.SchedulesWork ||
		recorded.ExecutesWork ||
		recorded.ApprovesWork ||
		recorded.ClaimsAuthorityAdvance ||
		!recorded.RSIRemainsDenied {
		t.Fatalf("v02 terminal status examples lost readback enum binding: %#v", recorded)
	}

	validator, err := validateRecommendationEvidenceTypedFile(recordedPath, "ao.atlas.foundry-terminal-status-examples.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:foundry-terminal-status-examples" {
		t.Fatalf("expected typed Foundry terminal status examples validator, got %s", validator)
	}
}
