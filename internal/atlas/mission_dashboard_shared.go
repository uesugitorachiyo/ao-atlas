package atlas

import (
	"path/filepath"
	"strings"
)

type missionDashboardEvidenceArtifact struct {
	SourcePath   string
	ResolvedPath string
	PublicPath   string
	Digest       string
}

func buildMissionDashboardEvidenceArtifact(path string) (missionDashboardEvidenceArtifact, error) {
	path = strings.TrimSpace(path)
	resolvedPath, err := resolveMissionDashboardEvidencePath(path)
	if err != nil {
		return missionDashboardEvidenceArtifact{}, err
	}
	digest, err := digestTextFileWithNormalizedLineEndings(resolvedPath)
	if err != nil {
		return missionDashboardEvidenceArtifact{}, err
	}
	return missionDashboardEvidenceArtifact{
		SourcePath:   path,
		ResolvedPath: resolvedPath,
		PublicPath:   missionDashboardEvidencePublicPath(path),
		Digest:       digest,
	}, nil
}

func missionDashboardEvidencePublicPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if filepath.IsAbs(path) || driveAbsPattern.MatchString(path) {
		return publicArtifactRef(path)
	}
	return filepath.ToSlash(filepath.Clean(path))
}

func buildMissionDashboardFreshnessCheckFromArtifact(name string, passed bool, artifact missionDashboardEvidenceArtifact) AtlasMissionDashboardFreshnessCheck {
	status := "failed"
	if passed {
		status = "passed"
	}
	return AtlasMissionDashboardFreshnessCheck{
		Name:           name,
		Status:         status,
		EvidencePath:   artifact.PublicPath,
		EvidenceDigest: artifact.Digest,
	}
}

func buildMissionDashboardProvenanceLinkFromArtifact(repo, role string, artifact missionDashboardEvidenceArtifact, expectedDigest string, finalResponseAllowed, rsiRemainsDenied bool) AtlasMissionDashboardProvenanceLink {
	digestMatches := strings.TrimSpace(expectedDigest) == artifact.Digest
	return AtlasMissionDashboardProvenanceLink{
		Repo:                         strings.TrimSpace(repo),
		Role:                         strings.TrimSpace(role),
		EvidencePath:                 artifact.PublicPath,
		EvidenceDigest:               artifact.Digest,
		ProvenanceLinkStatus:         "linked",
		DashboardRowMatched:          strings.TrimSpace(repo) != "" && strings.TrimSpace(role) != "",
		ClosureEvidenceDigestMatches: digestMatches,
		ArtifactDigestVerified:       digestMatches && digestPattern.MatchString(artifact.Digest),
		FinalResponseAllowed:         finalResponseAllowed,
		RSIRemainsDenied:             rsiRemainsDenied,
		AuthorityAdvanceClaimed:      false,
	}
}

func resolveMissionDashboardEvidencePath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if filepath.IsAbs(path) || driveAbsPattern.MatchString(path) {
		return filepath.Clean(path), nil
	}
	root, err := findRepoRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, path), nil
}
