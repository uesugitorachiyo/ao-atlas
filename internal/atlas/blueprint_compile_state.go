package atlas

type blueprintCompileState struct {
	Artifacts  BlueprintCompileArtifacts
	Record     BlueprintImport
	Digests    map[string]string
	PackDigest string
	PackErr    error
}

func newBlockedBlueprintCompileState(paths BlueprintImportPaths) blueprintCompileState {
	packDigest, packErr := digestDirectory(paths.PackPath)
	if packErr != nil {
		packDigest = DigestBytes([]byte("missing-blueprint-pack:" + paths.PackPath))
	}
	digests := map[string]string{"blueprint_pack": packDigest}
	record := BlueprintImport{
		ContractVersion:         BlueprintImportContract,
		ID:                      "blueprint-import-blocked",
		ProjectID:               "unknown-project",
		Status:                  "blocked",
		Reason:                  "AO Atlas cannot compile Blueprint material until authorization and candidate rules are ready.",
		BlueprintPack:           SourceRef{Ref: publicArtifactRef(paths.PackPath), Digest: packDigest},
		Digests:                 digests,
		ReadyForFoundry:         false,
		SafeToExecute:           false,
		LiveExecutionProven:     false,
		SchedulesWork:           false,
		ExecutesWork:            false,
		ApprovesWork:            false,
		MutatesRepositories:     false,
		CallsProviders:          false,
		ReleaseOrPublishAllowed: false,
	}
	return blueprintCompileState{
		Artifacts:  BlueprintCompileArtifacts{Record: record},
		Record:     record,
		Digests:    digests,
		PackDigest: packDigest,
		PackErr:    packErr,
	}
}
