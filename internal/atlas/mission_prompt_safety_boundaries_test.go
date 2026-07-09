package atlas

import (
	"strings"
	"testing"
)

func TestPromptSafetyBoundaryTemplateRendersAuditedSharedText(t *testing.T) {
	var b strings.Builder
	writeAtlasPromptSafetyBoundaries(&b, AtlasPromptSafetyBoundaryOptions{
		PrefixLines: []string{"Keep exactly one executable mutation node active at a time."},
		SuffixLines: []string{"Use existing repo auth only for normal PR, CI, and merge if available without exposing credentials."},
	})

	want := strings.Join([]string{
		"Safety boundaries:",
		"- Keep exactly one executable mutation node active at a time.",
		"- No provider calls.",
		"- No credential or token inspection.",
		"- No direct main mutation.",
		"- No release, deploy, publish, upload, or tag.",
		"- No dependency updates unless separately authorized.",
		"- No auth, policy, or config widening.",
		"- No hidden instruction mutation.",
		"- No broad RSI claim.",
		"- RSI remains denied.",
		"- Use existing repo auth only for normal PR, CI, and merge if available without exposing credentials.",
		"",
	}, "\n")
	if b.String() != want {
		t.Fatalf("safety boundary template changed\nwant:\n%s\ngot:\n%s", want, b.String())
	}
}
