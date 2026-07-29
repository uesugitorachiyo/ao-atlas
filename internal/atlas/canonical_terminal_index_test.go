package atlas

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestBuildCanonicalTerminalIndexReconcilesTerminalState(t *testing.T) {
	root, manifest := writeTerminalIndexFixture(t, fixtureOptions{})
	index, err := BuildCanonicalTerminalIndex(root, manifest)
	if err != nil {
		t.Fatal(err)
	}
	if !index.CompletionObserved || !index.ReadinessPassed || !index.FinalResponseAllowed {
		t.Fatalf("unexpected passing index: %+v", index)
	}
	if index.Lease.Status != "within_window" || index.Counts.Completed != 40 {
		t.Fatalf("unexpected lease or counts: %+v", index)
	}
	if err := VerifyCanonicalTerminalIndex(root, index); err != nil {
		t.Fatal(err)
	}
}

func TestBuildCanonicalTerminalIndexFailClosedStates(t *testing.T) {
	tests := []struct {
		name     string
		options  fixtureOptions
		conflict string
	}{
		{name: "below minimum", options: fixtureOptions{elapsed: 90}, conflict: "lease_minimum_not_met"},
		{name: "above maximum", options: fixtureOptions{elapsed: 1191}, conflict: "lease_maximum_exceeded"},
		{name: "mislabeled maximum", options: fixtureOptions{elapsed: 1191, leaseStatus: "minimum_minutes_met"}, conflict: "terminal_lease_status_mismatch"},
		{name: "ready with final true", options: fixtureOptions{terminalCompleted: 39, ready: 1, final: true}, conflict: "unfinished_work_final_response"},
		{name: "failed lifecycle", options: fixtureOptions{terminalCompleted: 39, failed: 1, final: true}, conflict: "failed_work_final_response"},
		{name: "stale duration", options: fixtureOptions{durationCompleted: 26}, conflict: "duration_state_stale"},
		{name: "missing terminal", options: fixtureOptions{missingTerminal: true}, conflict: "canonical_terminal_missing"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root, manifest := writeTerminalIndexFixture(t, test.options)
			index, err := BuildCanonicalTerminalIndex(root, manifest)
			if err != nil {
				t.Fatal(err)
			}
			if index.ReadinessPassed || index.FinalResponseAllowed {
				t.Fatalf("unsafe passing index: %+v", index)
			}
			if !containsString(index.Conflicts, test.conflict) {
				t.Fatalf("missing conflict %q in %v", test.conflict, index.Conflicts)
			}
		})
	}
}

func TestBuildCanonicalTerminalIndexRejectsInvalidEvidence(t *testing.T) {
	tests := []struct {
		name    string
		options fixtureOptions
		want    string
	}{
		{name: "altered digest", options: fixtureOptions{alterDigest: true}, want: "digest mismatch"},
		{name: "identity mismatch", options: fixtureOptions{terminalMission: "other"}, want: "mission identity mismatch"},
		{name: "root terminal disagreement without lineage", options: fixtureOptions{rootCompleted: 40, terminalCompleted: 39}, want: "non-monotonic"},
		{name: "non monotonic sequence", options: fixtureOptions{rootCompleted: 41}, want: "non-monotonic"},
		{name: "malformed json", options: fixtureOptions{malformed: true}, want: "invalid JSON"},
		{name: "duplicate key", options: fixtureOptions{duplicate: true}, want: "duplicate JSON key"},
		{name: "oversized input", options: fixtureOptions{oversized: true}, want: "size limit"},
		{name: "path traversal", options: fixtureOptions{traversal: true}, want: "unsafe evidence path"},
		{name: "symlink", options: fixtureOptions{symlink: true}, want: "regular file"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root, manifest := writeTerminalIndexFixture(t, test.options)
			_, err := BuildCanonicalTerminalIndex(root, manifest)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want substring %q", err, test.want)
			}
		})
	}
}

func TestCanonicalTerminalIndexIsDeterministic(t *testing.T) {
	root, manifest := writeTerminalIndexFixture(t, fixtureOptions{})
	first, err := BuildCanonicalTerminalIndex(root, manifest)
	if err != nil {
		t.Fatal(err)
	}
	second, err := BuildCanonicalTerminalIndex(root, manifest)
	if err != nil {
		t.Fatal(err)
	}
	if first.Digest != second.Digest {
		t.Fatalf("digest changed: %s != %s", first.Digest, second.Digest)
	}
}

func TestVerifyCanonicalTerminalIndexRejectsDigestValidSemanticContradiction(t *testing.T) {
	root, manifest := writeTerminalIndexFixture(t, fixtureOptions{})
	index, err := BuildCanonicalTerminalIndex(root, manifest)
	if err != nil {
		t.Fatal(err)
	}
	index.Lease.Status = "maximum_exceeded"
	unsigned := index
	unsigned.Digest = ""
	data, err := json.Marshal(unsigned)
	if err != nil {
		t.Fatal(err)
	}
	index.Digest = DigestBytes(data)
	if err := VerifyCanonicalTerminalIndex(root, index); err == nil || !strings.Contains(err.Error(), "lease status") {
		t.Fatalf("error = %v, want semantic lease rejection", err)
	}
}

type fixtureOptions struct {
	elapsed           int
	leaseStatus       string
	ready             int
	failed            int
	final             bool
	rootCompleted     int
	terminalCompleted int
	durationCompleted int
	terminalMission   string
	alterDigest       bool
	missingTerminal   bool
	malformed         bool
	duplicate         bool
	oversized         bool
	traversal         bool
	symlink           bool
}

func writeTerminalIndexFixture(t *testing.T, options fixtureOptions) (string, string) {
	t.Helper()
	root := t.TempDir()
	mission := "fixture-wave"
	elapsed := defaultInt(options.elapsed, 150)
	terminalCompleted := defaultInt(options.terminalCompleted, 40)
	durationCompleted := defaultInt(options.durationCompleted, terminalCompleted)
	terminalMission := options.terminalMission
	if terminalMission == "" {
		terminalMission = mission
	}
	final := true
	if options.ready > 0 || options.failed > 0 {
		final = options.final
	}
	leaseStatus := options.leaseStatus
	if leaseStatus == "" {
		leaseStatus = "within_window"
	}
	rootJSON := `{"contract_version":"fixture.v1","mission_id":"` + mission + `","counts":{"completed":` +
		itoa(options.rootCompleted) + `,"ready":40,"blocked":0,"failed":0},"final_response_allowed":false}`
	terminalJSON := `{"contract_version":"fixture.v1","mission_id":"` + terminalMission +
		`","counts":{"completed":` + itoa(terminalCompleted) + `,"ready":` + itoa(options.ready) +
		`,"blocked":0,"failed":` + itoa(options.failed) + `},"elapsed_minutes":` + itoa(elapsed) +
		`,"lease_time_status":"` + leaseStatus + `","final_response_allowed":` + boolString(final) +
		`,"exact_next_action":"none","safety_boundaries":{"inbound_network":false,"credential_changes":false,"release":false}}`
	durationJSON := `{"contract_version":"fixture.v1","mission_id":"` + mission + `","completed_nodes":` +
		itoa(durationCompleted) + `,"elapsed_minutes":` + itoa(elapsed) + `}`
	leaseJSON := `{"contract_version":"fixture.v1","mission_id":"` + mission +
		`","minimum_nodes":40,"minimum_minutes":120,"target_minutes":150,"maximum_minutes":180}`

	paths := []struct {
		role string
		name string
		data string
	}{
		{"lease", "lease.json", leaseJSON},
		{"root", "root.json", rootJSON},
		{"duration", "duration.json", durationJSON},
	}
	if !options.missingTerminal {
		paths = append(paths, struct {
			role string
			name string
			data string
		}{"terminal", "terminal.json", terminalJSON})
	}
	if options.malformed {
		paths[1].data = "{"
	}
	if options.duplicate {
		paths[1].data = `{"mission_id":"fixture-wave","mission_id":"fixture-wave"}`
	}
	if options.oversized {
		paths[1].data = `{"padding":"` + strings.Repeat("x", canonicalTerminalArtifactMaxBytes) + `"}`
	}

	manifestArtifacts := ""
	for i, item := range paths {
		path := filepath.Join(root, item.name)
		if err := os.WriteFile(path, []byte(item.data), 0o600); err != nil {
			t.Fatal(err)
		}
		relative := item.name
		if options.symlink && item.role == "root" {
			target := filepath.Join(root, "target.json")
			if err := os.Rename(path, target); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(target, path); err != nil {
				t.Fatal(err)
			}
		}
		if options.traversal && item.role == "root" {
			relative = "../root.json"
		}
		digest := DigestBytes([]byte(item.data))
		if options.alterDigest && item.role == "root" {
			digest = "sha256:" + strings.Repeat("0", 64)
		}
		if i > 0 {
			manifestArtifacts += ","
		}
		state := map[string]string{
			"lease": "lease_authority", "root": "initial_snapshot",
			"duration": "duration_snapshot", "terminal": "terminal_candidate",
		}[item.role]
		manifestArtifacts += `{"role":"` + item.role + `","sequence":` + itoa(i) +
			`,"state":"` + state + `","path":"` + relative + `","sha256":"` + digest + `"}`
	}
	manifestJSON := `{"contract_version":"ao.canonical-terminal-index-input.v1","mission_id":"` + mission +
		`","evidence_root":".","generated_at_utc":"` + time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC).Format(time.RFC3339) +
		`","artifacts":[` + manifestArtifacts + `]}`
	manifest := filepath.Join(root, "manifest.json")
	if err := os.WriteFile(manifest, []byte(manifestJSON), 0o600); err != nil {
		t.Fatal(err)
	}
	return root, manifest
}

func defaultInt(value, fallback int) int {
	if value == 0 {
		return fallback
	}
	return value
}

func itoa(value int) string {
	return fmt.Sprintf("%d", value)
}

func boolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
