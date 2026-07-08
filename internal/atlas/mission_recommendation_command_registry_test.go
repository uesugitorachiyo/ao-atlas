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
		!strings.Contains(text, "mission-dashboard-compact-filters, complete-node, resume, or validate-evidence") {
		t.Fatalf("unknown command did not render registry-backed help: %s", text)
	}
}
