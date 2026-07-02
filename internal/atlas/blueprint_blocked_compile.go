package atlas

import (
	"fmt"
	"strings"
)

func buildBlueprintBlockedCompileArtifacts(artifacts BlueprintCompileArtifacts, record BlueprintImport, sourceLoad blueprintCompileSourceLoad) BlueprintCompileArtifacts {
	blockedRequest := buildBlueprintBlockedRequest(record, sourceLoad.Missing, sourceLoad.Blockers)
	artifacts.Record = blockedRequest.Record
	artifacts.Request = blockedRequest.Request
	return artifacts
}

func blueprintBlockedCompileError(record BlueprintImport) error {
	return fmt.Errorf("blueprint import blocked: %s", strings.Join(record.BlockingNextActions, "; "))
}
