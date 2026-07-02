package atlas

type blueprintReadyRecordInputs struct {
	Record        BlueprintImport
	Candidate     BlueprintCandidateSelection
	Digests       map[string]string
	FoundryDigest string
	HandoffDigest string
}

func buildBlueprintReadyRecord(inputs blueprintReadyRecordInputs) (BlueprintImport, error) {
	record := inputs.Record
	record.Status = "ready"
	record.Reason = "Blueprint authorization is ready and Atlas compiled digest-bound Foundry import material."
	record.CandidateSelection = inputs.Candidate
	record.DownstreamFoundryImport = SourceRef{Ref: "foundry-import/foundry-import.json", Digest: inputs.FoundryDigest}
	record.DownstreamFoundryContinuationHandoff = SourceRef{Ref: "foundry-import/foundry-continuation-handoff.json", Digest: inputs.HandoffDigest}
	record.Digests = inputs.Digests
	record.ReadyForFoundry = true
	if err := ValidateBlueprintImport(record); err != nil {
		return BlueprintImport{}, err
	}
	return record, nil
}
