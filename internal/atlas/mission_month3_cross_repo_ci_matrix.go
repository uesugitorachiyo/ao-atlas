package atlas

import (
	"fmt"
	"strings"
)

func BuildAtlasMonth3CrossRepoCIMatrix(nodeID, sentinelSignalStatePath string) (AtlasMonth3CrossRepoCIMatrix, error) {
	nodeID = strings.TrimSpace(nodeID)
	if nodeID == "" {
		return AtlasMonth3CrossRepoCIMatrix{}, fmt.Errorf("node id is required")
	}
	sentinel, err := LoadJSON[AtlasSentinelSignalStateFixture](sentinelSignalStatePath)
	if err != nil {
		return AtlasMonth3CrossRepoCIMatrix{}, err
	}
	if err := ValidateAtlasSentinelSignalStateFixture(sentinel); err != nil {
		return AtlasMonth3CrossRepoCIMatrix{}, err
	}
	sentinelDigest, err := digestTextFileWithNormalizedLineEndings(sentinelSignalStatePath)
	if err != nil {
		return AtlasMonth3CrossRepoCIMatrix{}, err
	}
	repos := []AtlasMonth3CrossRepoCIRepo{
		{Name: "ao-mission", Runtime: "go", RequiredGate: "production-readiness"},
		{Name: "ao-atlas", Runtime: "go", RequiredGate: "production-readiness"},
		{Name: "ao-foundry", Runtime: "go", RequiredGate: "production-readiness"},
		{Name: "ao-covenant", Runtime: "go", RequiredGate: "production-readiness"},
		{Name: "ao-command", Runtime: "go", RequiredGate: "production-readiness"},
		{Name: "ao2", Runtime: "rust", RequiredGate: "workspace-ci"},
	}
	entries := make([]AtlasMonth3CrossRepoCIEntry, 0, len(repos)*len(sentinel.Signals))
	for _, repo := range repos {
		for _, signal := range sentinel.Signals {
			entries = append(entries, AtlasMonth3CrossRepoCIEntry{
				Repo:    repo.Name,
				Signal:  signal,
				State:   "pass",
				Verdict: "allow_continue",
			})
		}
	}
	matrix := AtlasMonth3CrossRepoCIMatrix{
		Schema:                    AtlasMonth3CrossRepoCIMatrixContract,
		NodeID:                    nodeID,
		Status:                    "cross_repo_ci_matrix_ready",
		SentinelSignalStatePath:   publicArtifactRef(sentinelSignalStatePath),
		SentinelSignalStateDigest: sentinelDigest,
		Repos:                     repos,
		RepoCount:                 len(repos),
		MatrixEntries:             entries,
		MatrixEntryCount:          len(entries),
		SentinelSignalStateBound:  sentinel.Status == "signal_states_ready",
		RequiresPassBeforeMerge:   true,
		BlocksOnFailure:           sentinelHasVerdict(sentinel, "failure", "block"),
		WaitsOnPending:            sentinelHasVerdict(sentinel, "pending", "wait"),
		SchedulesWork:             false,
		ExecutesWork:              false,
		ApprovesWork:              false,
		ClaimsAuthorityAdvance:    sentinel.ClaimsAuthorityAdvance,
		RSIRemainsDenied:          sentinel.RSIRemainsDenied,
	}
	if !matrix.SentinelSignalStateBound || !matrix.BlocksOnFailure || !matrix.WaitsOnPending || matrix.ClaimsAuthorityAdvance || !matrix.RSIRemainsDenied {
		matrix.Status = "cross_repo_ci_matrix_failed"
	}
	if err := ValidateAtlasMonth3CrossRepoCIMatrix(matrix); err != nil {
		return AtlasMonth3CrossRepoCIMatrix{}, err
	}
	return matrix, nil
}

func sentinelHasVerdict(fixture AtlasSentinelSignalStateFixture, state, verdict string) bool {
	for _, row := range fixture.Matrix {
		if row.State == state && row.Verdict == verdict {
			return true
		}
	}
	return false
}

func ValidateAtlasMonth3CrossRepoCIMatrix(matrix AtlasMonth3CrossRepoCIMatrix) error {
	var errs []string
	requireContract(&errs, "month3_cross_repo_ci_matrix", matrix.Schema, AtlasMonth3CrossRepoCIMatrixContract)
	requireField(&errs, "node_id", matrix.NodeID)
	checkPublicPath(&errs, "node_id", matrix.NodeID, true)
	if !oneOf(matrix.Status, "cross_repo_ci_matrix_ready", "cross_repo_ci_matrix_failed") {
		errs = append(errs, "status must be cross_repo_ci_matrix_ready or cross_repo_ci_matrix_failed")
	}
	requireField(&errs, "sentinel_signal_state_path", matrix.SentinelSignalStatePath)
	checkPublicPath(&errs, "sentinel_signal_state_path", matrix.SentinelSignalStatePath, true)
	if !digestPattern.MatchString(matrix.SentinelSignalStateDigest) {
		errs = append(errs, "sentinel_signal_state_digest must be sha256 digest")
	}
	if matrix.RepoCount != len(matrix.Repos) || matrix.RepoCount != 6 {
		errs = append(errs, "repo_count must match six repos")
	}
	if matrix.MatrixEntryCount != len(matrix.MatrixEntries) || matrix.MatrixEntryCount != matrix.RepoCount*4 {
		errs = append(errs, "matrix_entry_count must cover four signals per repo")
	}
	if !matrix.SentinelSignalStateBound {
		errs = append(errs, "sentinel_signal_state_bound must be true")
	}
	if !matrix.RequiresPassBeforeMerge {
		errs = append(errs, "requires_pass_before_merge must be true")
	}
	if !matrix.BlocksOnFailure {
		errs = append(errs, "blocks_on_failure must be true")
	}
	if !matrix.WaitsOnPending {
		errs = append(errs, "waits_on_pending must be true")
	}
	seenRepos := map[string]bool{}
	for i, repo := range matrix.Repos {
		prefix := fmt.Sprintf("repos[%d]", i)
		requireField(&errs, prefix+".name", repo.Name)
		requireField(&errs, prefix+".runtime", repo.Runtime)
		requireField(&errs, prefix+".required_gate", repo.RequiredGate)
		if seenRepos[repo.Name] {
			errs = append(errs, prefix+".name must be unique")
		}
		seenRepos[repo.Name] = true
	}
	for i, entry := range matrix.MatrixEntries {
		prefix := fmt.Sprintf("matrix_entries[%d]", i)
		requireField(&errs, prefix+".repo", entry.Repo)
		requireField(&errs, prefix+".signal", entry.Signal)
		requireField(&errs, prefix+".state", entry.State)
		requireField(&errs, prefix+".verdict", entry.Verdict)
		if !seenRepos[entry.Repo] {
			errs = append(errs, prefix+".repo must be declared in repos")
		}
		if entry.State != "pass" || entry.Verdict != "allow_continue" {
			errs = append(errs, prefix+".state/verdict must be pass/allow_continue")
		}
	}
	validateNoAuthorityEffects(&errs, matrix.SchedulesWork, matrix.ExecutesWork, matrix.ApprovesWork, matrix.ClaimsAuthorityAdvance, matrix.RSIRemainsDenied)
	return joinErrors(errs)
}

func WriteAtlasMonth3CrossRepoCIMatrix(path string, matrix AtlasMonth3CrossRepoCIMatrix) error {
	return WriteJSON(path, matrix)
}
