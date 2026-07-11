package atlas

import (
	"bytes"
	"strings"
	"testing"
)

func TestMissionRecommendationCommandRegistryDrivesDeterministicDispatchHelp(t *testing.T) {
	names := missionRecommendationCommandNames()
	if len(names) < 50 {
		t.Fatalf("expected command registry to cover the recommendation command catalog, got %d", len(names))
	}
	if names[0] != "import" || names[1] != "export-next-wave" || names[2] != "export-refactoring-wave" || names[len(names)-1] != "validate-evidence" {
		t.Fatalf("command registry order is not deterministic help order: %#v", names)
	}
	seen := map[string]bool{}
	for _, name := range names {
		if strings.TrimSpace(name) == "" {
			t.Fatalf("command registry contains blank command: %#v", names)
		}
		if seen[name] {
			t.Fatalf("command registry contains duplicate command %q: %#v", name, names)
		}
		seen[name] = true
	}
	for _, want := range []string{
		"next-track",
		"run-ledger-coverage-check",
		"mission-dashboard-compact-filters",
		"bounded-signer-contract-fixture",
		"canonical-contract-registry-manifest",
		"contract-compatibility-inventory",
		"canonical-json-vectors",
		"canonical-json-vector-smoke-checks",
		"sentinel-hosted-ci-workflow-fixture",
		"sentinel-signal-state-fixture",
		"signed-assurance-dry-run-fixture",
		"promoter-no-activation-boundary-fixture",
		"workspace-root-preflight-fixture",
		"bounded-execution-packet-fixture",
		"forge-goalrun-evidence-fixture",
		"execution-packet-regression-matrix",
		"durable-state-migration-metadata",
		"exactly-once-resume-accounting-fixture",
		"blueprint-canonical-preservation-fixture",
		"foundry-canonical-import-fixture",
		"command-covenant-field-parity-fixture",
		"complete-node",
		"resume",
	} {
		if !seen[want] {
			t.Fatalf("command registry missing %q: %#v", want, names)
		}
	}

	var out bytes.Buffer
	code := Run([]string{"mission", "recommendations", "does-not-exist"}, &out, &out)
	if code == 0 {
		t.Fatal("unknown mission recommendation command succeeded")
	}
	text := out.String()
	if !strings.Contains(text, "mission recommendations requires import, export-next-wave, export-refactoring-wave") ||
		!strings.Contains(text, "mission-dashboard-compact-filters, bounded-signer-contract-fixture, canonical-contract-registry-manifest, contract-compatibility-inventory, canonical-json-vectors, canonical-json-vector-smoke-checks, sentinel-hosted-ci-workflow-fixture, sentinel-signal-state-fixture, signed-assurance-dry-run-fixture, promoter-no-activation-boundary-fixture, workspace-root-preflight-fixture, bounded-execution-packet-fixture, forge-goalrun-evidence-fixture, execution-packet-regression-matrix, durable-state-migration-metadata, exactly-once-resume-accounting-fixture, blueprint-canonical-preservation-fixture, foundry-canonical-import-fixture, command-covenant-field-parity-fixture, complete-node, resume, or validate-evidence") {
		t.Fatalf("unknown command did not render registry-backed help: %s", text)
	}
}

func TestMissionRecommendationRunLedgerCommandCatalogIsRegistryBacked(t *testing.T) {
	fullCatalog := map[string]bool{}
	for _, name := range missionRecommendationCommandNames() {
		fullCatalog[name] = true
	}

	ledgerCommands := missionRecommendationRunLedgerCommandNames()
	want := []string{
		"export-refactoring-wave",
		"next-track",
		"consumed-ledger",
		"track-registry",
		"final-response-gates",
		"schema-registry",
		"schema-registry-coverage",
		"validate-evidence",
	}
	if strings.Join(ledgerCommands, ",") != strings.Join(want, ",") {
		t.Fatalf("run-ledger command catalog drifted: got %#v want %#v", ledgerCommands, want)
	}
	for _, command := range ledgerCommands {
		if !fullCatalog[command] {
			t.Fatalf("run-ledger command %q is not in the shared recommendation command registry", command)
		}
	}

	ledger := AtlasRecommendationCommandRunLedger{
		Schema:                 AtlasRecommendationCommandRunLedgerContract,
		Status:                 "recorded",
		Command:                "not-a-command",
		ArtifactPath:           "artifact.json",
		ArtifactDigest:         "sha256:" + strings.Repeat("0", 64),
		ArtifactSchema:         AtlasRecommendationNextTrackDecisionContract,
		TypedValidator:         "typed:recommendation-next-track-decision",
		OutputStatus:           "routed",
		RecordsInvocation:      true,
		NoPromotionRequested:   true,
		PromotionGranted:       false,
		ClaimsAuthorityAdvance: false,
		RSIRemainsDenied:       true,
		SafeToExecute:          false,
		SchedulesWork:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
		MutatesRepositories:    false,
	}
	err := ValidateAtlasRecommendationCommandRunLedger(ledger)
	if err == nil {
		t.Fatal("invalid run-ledger command was accepted")
	}
	if !strings.Contains(err.Error(), "command must be export-refactoring-wave, next-track, consumed-ledger, track-registry") ||
		!strings.Contains(err.Error(), "schema-registry-coverage, or validate-evidence") {
		t.Fatalf("run-ledger command error did not use the registry-backed catalog: %v", err)
	}
}

func TestMissionRecommendationCommandRegistrySeparatesPlanningOnlyAndMutationCapableCommands(t *testing.T) {
	planningOnly := map[string]bool{}
	for _, name := range missionRecommendationPlanningOnlyCommandNames() {
		planningOnly[name] = true
	}
	mutationCapable := map[string]bool{}
	for _, name := range missionRecommendationMutationCapableCommandNames() {
		mutationCapable[name] = true
	}

	for _, name := range []string{
		"export-next-wave",
		"export-refactoring-wave",
		"next-track",
		"consumed-ledger",
		"track-registry",
		"final-response-gates",
		"schema-registry",
		"schema-registry-coverage",
		"validate-evidence",
	} {
		if !planningOnly[name] {
			t.Fatalf("recommendation command %q must be classified as planning-only", name)
		}
		if mutationCapable[name] {
			t.Fatalf("planning-only recommendation command %q was also classified mutation-capable", name)
		}
	}
	for _, name := range []string{"import", "complete-node", "resume"} {
		if !mutationCapable[name] {
			t.Fatalf("recommendation lifecycle command %q must be classified mutation-capable", name)
		}
		if planningOnly[name] {
			t.Fatalf("mutation-capable recommendation command %q was also classified planning-only", name)
		}
	}
	for _, command := range missionRecommendationCommandRegistry() {
		if command.commandClass == "" {
			t.Fatalf("recommendation command %q has no command class", command.name)
		}
		if planningOnly[command.name] == mutationCapable[command.name] {
			t.Fatalf("recommendation command %q must be in exactly one class", command.name)
		}
	}
}
