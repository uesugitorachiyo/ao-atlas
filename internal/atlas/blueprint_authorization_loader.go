package atlas

import "strings"

type blueprintAuthorizationLoadResult struct {
	Record     BlueprintImport
	AuthDigest string
	Missing    []string
	Blockers   []string
}

func loadBlueprintAuthorization(paths BlueprintImportPaths, record BlueprintImport, rules BlueprintCandidateRules, packDigest string, digests map[string]string) blueprintAuthorizationLoadResult {
	result := blueprintAuthorizationLoadResult{Record: record}
	var authorization BlueprintBuildAuthorization
	if strings.TrimSpace(paths.AuthorizationPath) == "" {
		result.Missing = append(result.Missing, "build_authorization")
		result.Blockers = append(result.Blockers, "provide AO Blueprint build authorization")
		return result
	}
	if err := readJSONIfPossible(paths.AuthorizationPath, &authorization); err != nil {
		result.Missing = append(result.Missing, "build_authorization")
		result.Blockers = append(result.Blockers, "provide readable AO Blueprint build authorization")
		return result
	}
	authDigest, _ := digestFile(paths.AuthorizationPath)
	digests["build_authorization"] = authDigest
	record.BuildAuthorization = SourceRef{Ref: publicArtifactRef(paths.AuthorizationPath), Digest: authDigest}
	authMissing, authBlockers := validateBlueprintAuthorization(authorization, rules, packDigest)
	result.Record = record
	result.AuthDigest = authDigest
	result.Missing = append(result.Missing, authMissing...)
	result.Blockers = append(result.Blockers, authBlockers...)
	return result
}
