package atlas

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAtlasReleaseFinalizeWorkflowStructure(t *testing.T) {
	path := filepath.Join(repoRoot(t), ".github", "workflows", "release-finalize.yml")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read release finalizer workflow: %v", err)
	}
	workflow := string(content)
	if err := validateReleaseFinalizeWorkflowStructure(workflow); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		mutate  func(string) string
		wantErr string
	}{
		{
			name: "missing protected environment",
			mutate: func(value string) string {
				return strings.Replace(value, "    environment: protected-release\n", "", 1)
			},
			wantErr: "publish-release must use protected-release environment",
		},
		{
			name: "top-level contents write",
			mutate: func(value string) string {
				return strings.Replace(value, "permissions:\n  actions: read\n  contents: read", "permissions:\n  actions: read\n  contents: write", 1)
			},
			wantErr: "workflow permissions must be exactly actions read and contents read",
		},
		{
			name: "missing live gate",
			mutate: func(value string) string {
				return strings.Replace(value, "inputs.dry_run == false", "inputs.dry_run", 1)
			},
			wantErr: "publish-release must require live mode",
		},
		{
			name: "wrong producer workflow",
			mutate: func(value string) string {
				return strings.ReplaceAll(value, ".github/workflows/release-rehearsal.yml", ".github/workflows/ci.yml")
			},
			wantErr: "producer workflow path must be exact",
		},
		{
			name: "missing expected plan digest",
			mutate: func(value string) string {
				return strings.Replace(value, "      expected_plan_digest:\n", "", 1)
			},
			wantErr: `missing required workflow input "expected_plan_digest:"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateReleaseFinalizeWorkflowStructure(tt.mutate(workflow))
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestAtlasReleaseFinalizeWorkflowAuthenticatesExactProducerArtifactInventory(t *testing.T) {
	content, err := os.ReadFile(filepath.Join(repoRoot(t), ".github", "workflows", "release-finalize.yml"))
	if err != nil {
		t.Fatal(err)
	}
	workflow := string(content)
	for _, required := range []string{
		`.total_count == 5`,
		`[.artifacts[].name] | unique | length) == 5`,
		`all(.artifacts[]; .expired == false)`,
		`"ao-atlas-release-input-binding-$SOURCE_SHA"`,
		`"ao-atlas-release-candidate-linux-x86_64-$SOURCE_SHA"`,
		`"ao-atlas-release-candidate-macos-aarch64-$SOURCE_SHA"`,
		`"ao-atlas-release-candidate-windows-x86_64-$SOURCE_SHA"`,
		`"ao-atlas-release-rehearsal-plan-$SOURCE_SHA"`,
		`cmp actual-artifacts.txt expected-artifacts.txt`,
	} {
		if !strings.Contains(workflow, required) {
			t.Fatalf("producer artifact authentication missing %q", required)
		}
	}
}

func validateReleaseFinalizeWorkflowStructure(workflow string) error {
	required := []string{
		"workflow_dispatch:", "producer_run_id:", "expected_source_sha:",
		"expected_version:", "expected_tag:", "expected_manifest_digest:",
		"expected_plan_digest:", "dry_run:", "live_confirmation:",
		"actions: read", "contents: read", "validate-imported-release:",
		"dry-run-boundary:", "publish-release:", "verify-public-release:",
		"environment: protected-release", "contents: write",
		"inputs.dry_run == false", "actions/runs/$PRODUCER_RUN_ID",
		".github/workflows/release-rehearsal.yml", `conclusion == "success"`,
		"scripts/verify-release-rehearsal-candidates.sh", "gh release create",
		"gh release download", "git/ref/tags/$TAG", "provider_credentials_used:false",
	}
	for _, value := range required {
		if !strings.Contains(workflow, value) {
			switch value {
			case "actions: read", "contents: read":
				return fmt.Errorf("workflow permissions must be exactly actions read and contents read")
			case "expected_plan_digest:":
				return fmt.Errorf("missing required workflow input %q", value)
			case "environment: protected-release":
				return fmt.Errorf("publish-release must use protected-release environment")
			case "inputs.dry_run == false":
				return fmt.Errorf("publish-release must require live mode")
			case ".github/workflows/release-rehearsal.yml":
				return fmt.Errorf("producer workflow path must be exact")
			default:
				return fmt.Errorf("release finalizer workflow missing %q", value)
			}
		}
	}

	triggers := yamlChildKeys(yamlTopLevelSection(workflow, "on:"), 2)
	if len(triggers) != 1 || triggers[0] != "workflow_dispatch" {
		return fmt.Errorf("workflow_dispatch must be the only trigger: %v", triggers)
	}
	permissions := strings.TrimSpace(yamlTopLevelSection(workflow, "permissions:"))
	if permissions != "permissions:\n  actions: read\n  contents: read" {
		return fmt.Errorf("workflow permissions must be exactly actions read and contents read")
	}

	jobs := yamlTopLevelSection(workflow, "jobs:")
	publish := yamlJobSection(jobs, "publish-release")
	if publish == "" || !strings.Contains(publish, "environment: protected-release") {
		return fmt.Errorf("publish-release must use protected-release environment")
	}
	if !strings.Contains(publish, "inputs.dry_run == false") {
		return fmt.Errorf("publish-release must require live mode")
	}
	nonPublish := strings.Replace(jobs, publish, "", 1)
	for _, capability := range []string{"contents: write", "gh release create", "gh release upload", "git tag", "git push"} {
		if strings.Contains(nonPublish, capability) {
			return fmt.Errorf("publication capability %q must be limited to publish-release", capability)
		}
	}
	for _, forbidden := range []string{"write-all", "pull-requests: write", "/environments", "gh release delete", "git push --force"} {
		if strings.Contains(workflow, forbidden) {
			return fmt.Errorf("forbidden release finalizer capability %q", forbidden)
		}
	}
	return nil
}

func yamlJobSection(jobs, name string) string {
	header := "  " + name + ":"
	lines := strings.Split(jobs, "\n")
	start := -1
	for i, line := range lines {
		if line == header {
			start = i
			break
		}
	}
	if start < 0 {
		return ""
	}
	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		if strings.HasPrefix(lines[i], "  ") && !strings.HasPrefix(lines[i], "    ") && strings.HasSuffix(lines[i], ":") {
			end = i
			break
		}
	}
	return strings.Join(lines[start:end], "\n")
}
