package atlas

import (
	"fmt"
	"strings"
)

const ConsolidationRepositoryBaselineContract = "ao.atlas.consolidation-repository-baseline.v0.1"

type ConsolidationRepositoryBaseline struct {
	Schema              string                                 `json:"schema"`
	MissionID           string                                 `json:"mission_id"`
	TargetInstance      string                                 `json:"target_instance"`
	GeneratedAtUTC      string                                 `json:"generated_at_utc"`
	SourceKind          string                                 `json:"source_kind"`
	Status              string                                 `json:"status"`
	SafeToExecute       bool                                   `json:"safe_to_execute"`
	ExecutesWork        bool                                   `json:"executes_work"`
	ApprovesWork        bool                                   `json:"approves_work"`
	MutatesRepositories bool                                   `json:"mutates_repositories"`
	Repositories        []ConsolidationRepositoryBaselineEntry `json:"repositories"`
}

type ConsolidationRepositoryBaselineEntry struct {
	Repository            string   `json:"repository"`
	Branch                string   `json:"branch"`
	Head                  string   `json:"head"`
	Upstream              string   `json:"upstream"`
	Clean                 bool     `json:"clean"`
	DirtyFiles            string   `json:"dirty_files"`
	PreExistingDirtyFiles []string `json:"pre_existing_dirty_files"`
	WaveOwnedFiles        []string `json:"wave_owned_files"`
	StateClass            string   `json:"state_class"`
	AheadOfOriginMain     int      `json:"ahead_of_origin_main"`
	BehindOriginMain      int      `json:"behind_origin_main"`
}

func ValidateConsolidationRepositoryBaseline(baseline ConsolidationRepositoryBaseline) error {
	var errs []string
	requireContract(&errs, "repository_baseline", baseline.Schema, ConsolidationRepositoryBaselineContract)
	requireField(&errs, "mission_id", baseline.MissionID)
	requireField(&errs, "target_instance", baseline.TargetInstance)
	requireField(&errs, "generated_at_utc", baseline.GeneratedAtUTC)
	if baseline.Status != "passed" {
		errs = append(errs, "status must be passed")
	}
	if baseline.ExecutesWork || baseline.ApprovesWork || baseline.MutatesRepositories {
		errs = append(errs, "repository baseline must not execute, approve, or mutate repositories")
	}
	if len(baseline.Repositories) != 14 {
		errs = append(errs, fmt.Sprintf("repositories must contain exactly 14 active repositories, got %d", len(baseline.Repositories)))
	}
	seen := map[string]bool{}
	for i, repo := range baseline.Repositories {
		prefix := fmt.Sprintf("repositories[%d]", i)
		requireField(&errs, prefix+".repository", repo.Repository)
		requireField(&errs, prefix+".branch", repo.Branch)
		requireField(&errs, prefix+".head", repo.Head)
		if seen[repo.Repository] {
			errs = append(errs, prefix+" duplicates repository "+repo.Repository)
		}
		seen[repo.Repository] = true
		if len(repo.WaveOwnedFiles) != 0 {
			errs = append(errs, prefix+".wave_owned_files must be empty for a pre-mutation baseline")
		}
		if !repo.Clean && strings.TrimSpace(repo.DirtyFiles) == "" {
			errs = append(errs, prefix+".dirty_files must be present when clean is false")
		}
		expected := consolidationRepositoryStateClass(repo)
		if repo.StateClass != expected {
			errs = append(errs, prefix+fmt.Sprintf(".state_class must be %q for observed branch and sync state, got %q", expected, repo.StateClass))
		}
	}
	return joinErrors(errs)
}

func consolidationRepositoryStateClass(repo ConsolidationRepositoryBaselineEntry) string {
	codexBranch := strings.HasPrefix(repo.Branch, "codex/")
	outOfSync := repo.AheadOfOriginMain != 0 || repo.BehindOriginMain != 0
	switch {
	case !repo.Clean && repo.BehindOriginMain > 0:
		return "pre_existing_dirty_and_behind"
	case !repo.Clean:
		return "pre_existing_dirty"
	case codexBranch && outOfSync:
		return "pre_existing_codex_branch_and_out_of_sync"
	case codexBranch:
		return "pre_existing_codex_branch"
	case outOfSync:
		return "pre_existing_out_of_sync"
	default:
		return "clean_synced"
	}
}
