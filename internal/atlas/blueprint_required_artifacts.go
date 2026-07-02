package atlas

import "path/filepath"

type blueprintRequiredArtifactsResult struct {
	Missing  []string
	Blockers []string
}

func loadBlueprintRequiredArtifacts(paths BlueprintImportPaths, digests map[string]string) blueprintRequiredArtifactsResult {
	result := blueprintRequiredArtifactsResult{}
	for name, path := range map[string]string{
		"implementation_spec": filepath.Join(paths.PackPath, "implementation-spec.md"),
		"quality_profile":     filepath.Join(paths.PackPath, "quality-profile.md"),
	} {
		digest, err := digestFile(path)
		if err != nil {
			result.Missing = append(result.Missing, name)
			result.Blockers = append(result.Blockers, "add "+filepath.Base(path)+" to the Blueprint pack")
			continue
		}
		digests[name] = digest
	}
	return result
}
