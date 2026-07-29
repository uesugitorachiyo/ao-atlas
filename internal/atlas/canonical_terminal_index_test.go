package atlas

import (
	"bytes"
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

func TestTerminalIndexCLISerializedPassingRoundTrip(t *testing.T) {
	root, manifest := writeTerminalIndexFixture(t, fixtureOptions{})
	firstPath := filepath.Join(root, "canonical-terminal-index.json")
	secondPath := filepath.Join(root, "canonical-terminal-index-repeat.json")

	firstBuild := runAtlasCLI(t, "terminal-index", "build", "--root", root, "--manifest", manifest, "--out", firstPath)
	if firstBuild.code != 0 {
		t.Fatalf("first build exit = %d, stderr = %q", firstBuild.code, firstBuild.stderr)
	}
	firstData, err := os.ReadFile(firstPath)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(firstData, &raw); err != nil {
		t.Fatal(err)
	}
	if got := bytes.TrimSpace(raw["conflict_codes"]); !bytes.Equal(got, []byte("[]")) {
		t.Fatalf("serialized conflict_codes = %s, want []", got)
	}
	var firstIndex CanonicalTerminalIndex
	if err := json.Unmarshal(firstData, &firstIndex); err != nil {
		t.Fatal(err)
	}

	secondBuild := runAtlasCLI(t, "terminal-index", "build", "--root", root, "--manifest", manifest, "--out", secondPath)
	if secondBuild.code != 0 {
		t.Fatalf("second build exit = %d, stderr = %q", secondBuild.code, secondBuild.stderr)
	}
	secondData, err := os.ReadFile(secondPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstData, secondData) {
		t.Fatal("repeated CLI builds with a fixed timestamp produced different bytes")
	}
	var secondIndex CanonicalTerminalIndex
	if err := json.Unmarshal(secondData, &secondIndex); err != nil {
		t.Fatal(err)
	}
	if firstIndex.Digest != secondIndex.Digest {
		t.Fatalf("repeated CLI build digest changed: %s != %s", firstIndex.Digest, secondIndex.Digest)
	}

	verify := runAtlasCLI(t, "terminal-index", "verify", "--root", root, "--index", firstPath)
	if verify.code != 0 {
		t.Fatalf("serialized verify exit = %d, stderr = %q", verify.code, verify.stderr)
	}
	if !strings.Contains(verify.stdout, "digest="+firstIndex.Digest) {
		t.Fatalf("verify output %q does not contain build digest %s", verify.stdout, firstIndex.Digest)
	}
}

func TestTerminalIndexCLISerializedValidation(t *testing.T) {
	t.Run("digest altered", func(t *testing.T) {
		root, manifest := writeTerminalIndexFixture(t, fixtureOptions{})
		indexPath := buildTerminalIndexWithCLI(t, root, manifest)
		index := readTerminalIndexFile(t, indexPath)
		index.Digest = "sha256:" + strings.Repeat("0", 64)
		if err := WriteJSON(indexPath, index); err != nil {
			t.Fatal(err)
		}
		verify := runAtlasCLI(t, "terminal-index", "verify", "--root", root, "--index", indexPath)
		if verify.code == 0 || !strings.Contains(verify.stderr, "index digest mismatch") {
			t.Fatalf("verify exit = %d, stderr = %q, want digest mismatch", verify.code, verify.stderr)
		}
	})

	t.Run("digest valid semantic contradiction", func(t *testing.T) {
		root, manifest := writeTerminalIndexFixture(t, fixtureOptions{})
		indexPath := buildTerminalIndexWithCLI(t, root, manifest)
		index := readTerminalIndexFile(t, indexPath)
		index.Lease.Status = "maximum_exceeded"
		unsigned := index
		unsigned.Digest = ""
		data, err := json.Marshal(unsigned)
		if err != nil {
			t.Fatal(err)
		}
		index.Digest = DigestBytes(data)
		if err := WriteJSON(indexPath, index); err != nil {
			t.Fatal(err)
		}
		verify := runAtlasCLI(t, "terminal-index", "verify", "--root", root, "--index", indexPath)
		if verify.code == 0 || !strings.Contains(verify.stderr, "lease status") {
			t.Fatalf("verify exit = %d, stderr = %q, want semantic lease rejection", verify.code, verify.stderr)
		}
	})

	t.Run("nonempty fail closed conflicts", func(t *testing.T) {
		root, manifest := writeTerminalIndexFixture(t, fixtureOptions{elapsed: 90})
		indexPath := buildTerminalIndexWithCLI(t, root, manifest)
		index := readTerminalIndexFile(t, indexPath)
		if len(index.Conflicts) == 0 {
			t.Fatal("fail-closed index has no conflict codes")
		}
		verify := runAtlasCLI(t, "terminal-index", "verify", "--root", root, "--index", indexPath)
		if verify.code != 0 {
			t.Fatalf("fail-closed serialized verify exit = %d, stderr = %q", verify.code, verify.stderr)
		}
	})
}

func TestDecodeStrictJSONPreservesNestedEmptyArrays(t *testing.T) {
	var decoded any
	if err := decodeStrictJSON([]byte(`{"outer":[[],{"inner":[]}]}`), &decoded); err != nil {
		t.Fatal(err)
	}
	object := decoded.(map[string]any)
	outer := object["outer"].([]any)
	if inner, ok := outer[0].([]any); !ok || inner == nil || len(inner) != 0 {
		t.Fatalf("first nested value = %#v, want non-nil empty array", outer[0])
	}
	nested := outer[1].(map[string]any)
	if inner, ok := nested["inner"].([]any); !ok || inner == nil || len(inner) != 0 {
		t.Fatalf("object nested value = %#v, want non-nil empty array", nested["inner"])
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
		`,"exact_next_action":"none","safety_boundaries":{"executes_work":false,"approves_work":false,` +
		`"mutates_repositories":false,"calls_providers":false,"publishes":false,"releases":false,` +
		`"deploys":false,"advances_authority":false}}`
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

type cliResult struct {
	code   int
	stdout string
	stderr string
}

func runAtlasCLI(t *testing.T, args ...string) cliResult {
	t.Helper()
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := Run(args, &stdout, &stderr)
	return cliResult{code: code, stdout: stdout.String(), stderr: stderr.String()}
}

func buildTerminalIndexWithCLI(t *testing.T, root, manifest string) string {
	t.Helper()
	indexPath := filepath.Join(root, "canonical-terminal-index.json")
	result := runAtlasCLI(t, "terminal-index", "build", "--root", root, "--manifest", manifest, "--out", indexPath)
	if result.code != 0 {
		t.Fatalf("build exit = %d, stderr = %q", result.code, result.stderr)
	}
	return indexPath
}

func readTerminalIndexFile(t *testing.T, path string) CanonicalTerminalIndex {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var index CanonicalTerminalIndex
	if err := json.Unmarshal(data, &index); err != nil {
		t.Fatal(err)
	}
	return index
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
