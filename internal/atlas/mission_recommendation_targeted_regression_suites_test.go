package atlas

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRecommendationTargetedRegressionSuitesListAndRunFocusedBoundaries(t *testing.T) {
	root := repoRoot(t)
	script := filepath.Join(root, "scripts", "recommendation-targeted-regressions.sh")
	body, err := os.ReadFile(script)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{
		"TestLongRunHardeningWave.*",
		"TestFinalClosureConsolidation.*",
		"TestMissionRecommendations.*",
	} {
		if strings.Contains(string(body), forbidden) {
			t.Fatalf("targeted regression script should not use broad long-wave selector %q", forbidden)
		}
	}

	list := exec.Command("bash", script, "list")
	list.Dir = root
	var listOut bytes.Buffer
	list.Stdout = &listOut
	list.Stderr = &listOut
	if err := list.Run(); err != nil {
		t.Fatalf("targeted regression suite list failed: %v\n%s", err, listOut.String())
	}
	for _, suite := range []string{"validator-boundaries", "fixture-builders", "compact-dashboard", "command-ledger", "all"} {
		if !strings.Contains(listOut.String(), suite) {
			t.Fatalf("targeted regression suite list missing %s: %s", suite, listOut.String())
		}
	}

	cmd := exec.Command("bash", script, "validator-boundaries")
	cmd.Dir = root
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("validator-boundaries targeted suite failed: %v\n%s", err, out.String())
	}
	if !strings.Contains(out.String(), "suite=validator-boundaries") ||
		strings.Contains(out.String(), "TestLongRunHardeningWave") ||
		strings.Contains(out.String(), "TestFinalClosureConsolidation") {
		t.Fatalf("validator-boundaries suite did not stay focused: %s", out.String())
	}
}
