package atlas

import (
	"fmt"
	"strings"
)

type AtlasPromptSafetyBoundaryOptions struct {
	PrefixLines []string
	SuffixLines []string
}

func writeAtlasPromptSafetyBoundaries(b *strings.Builder, options AtlasPromptSafetyBoundaryOptions) {
	b.WriteString("Safety boundaries:\n")
	for _, line := range options.PrefixLines {
		writeAtlasPromptSafetyBoundaryLine(b, line)
	}
	for _, line := range atlasAuditedPromptSafetyBoundaryLines() {
		writeAtlasPromptSafetyBoundaryLine(b, line)
	}
	for _, line := range options.SuffixLines {
		writeAtlasPromptSafetyBoundaryLine(b, line)
	}
}

func atlasAuditedPromptSafetyBoundaryLines() []string {
	return []string{
		"No provider calls.",
		"No credential or token inspection.",
		"No direct main mutation.",
		"No release, deploy, publish, upload, or tag.",
		"No dependency updates unless separately authorized.",
		"No auth, policy, or config widening.",
		"No hidden instruction mutation.",
		"No broad RSI claim.",
		"RSI remains denied.",
	}
}

func writeAtlasPromptSafetyBoundaryLine(b *strings.Builder, line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}
	b.WriteString(fmt.Sprintf("- %s\n", line))
}
