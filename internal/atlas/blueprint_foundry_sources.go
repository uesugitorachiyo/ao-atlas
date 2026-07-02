package atlas

func buildBlueprintFoundrySourceArtifacts(paths BlueprintImportPaths, contextPack ContextPack, packDigest string, authDigest string, digests map[string]string) []SourceRef {
	return []SourceRef{
		{Ref: publicArtifactRef(paths.PackPath), Digest: packDigest},
		{Ref: publicArtifactRef(paths.AuthorizationPath), Digest: authDigest},
		{Ref: "candidate-selection.json", Digest: digests["candidate_selection"]},
		{Ref: "context-packs/" + contextPack.ID + ".json", Digest: digests["context_pack"]},
		{Ref: "workgraph.json", Digest: digests["workgraph"]},
	}
}
