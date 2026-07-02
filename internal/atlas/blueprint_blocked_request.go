package atlas

type blueprintBlockedRequestResult struct {
	Record  BlueprintImport
	Request BlueprintRequest
}

func buildBlueprintBlockedRequest(record BlueprintImport, missing []string, blockers []string) blueprintBlockedRequestResult {
	request := BlueprintRequest{
		ContractVersion: BlueprintRequestContract,
		IntakeID:        firstNonEmpty(record.ProjectID, "blueprint-import") + "-intake",
		Status:          "blueprint_required",
		Missing:         uniqueStrings(missing),
		Reason:          "AO Atlas cannot emit a ready workgraph until Blueprint authorization is present, current, digest-bound, and scoped to this work.",
	}
	record.BlockingNextActions = uniqueStrings(blockers)
	if len(record.BlockingNextActions) == 0 {
		record.BlockingNextActions = []string{"return to AO Blueprint for build authorization"}
	}
	return blueprintBlockedRequestResult{
		Record:  record,
		Request: request,
	}
}
