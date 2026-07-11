package atlas

func BuildAtlasWorkspaceRootPreflightFixture() (AtlasWorkspaceRootPreflightFixture, error) {
	fixture := AtlasWorkspaceRootPreflightFixture{
		Schema:                      AtlasWorkspaceRootPreflightFixtureContract,
		Status:                      "preflight_ready",
		RepositoryIdentity:          "tiny-non-ao-repository",
		RepositoryIdentityValidated: true,
		NonAORepository:             true,
		ObjectiveDigest:             "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		ObjectiveDigestValidated:    true,
		WorkspaceRoot:               "fixtures/tiny-non-ao-repository",
		WorktreeBoundaryValidated:   true,
		SafeNextNode:                "bounded-diff-review",
		SafeNextNodeSelected:        true,
		SchedulesWork:               false,
		ExecutesWork:                false,
		ApprovesWork:                false,
		ClaimsAuthorityAdvance:      false,
		RSIRemainsDenied:            true,
	}
	if err := ValidateAtlasWorkspaceRootPreflightFixture(fixture); err != nil {
		return AtlasWorkspaceRootPreflightFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasWorkspaceRootPreflightFixture(fixture AtlasWorkspaceRootPreflightFixture) error {
	var errs []string
	requireContract(&errs, "workspace_root_preflight_fixture", fixture.Schema, AtlasWorkspaceRootPreflightFixtureContract)
	if fixture.Status != "preflight_ready" {
		errs = append(errs, "status must be preflight_ready")
	}
	requireField(&errs, "repository_identity", fixture.RepositoryIdentity)
	if !fixture.RepositoryIdentityValidated {
		errs = append(errs, "repository_identity_validated must be true")
	}
	if !fixture.NonAORepository {
		errs = append(errs, "non_ao_repository must be true")
	}
	validateRejectedTicketDigest(&errs, "objective_digest", fixture.ObjectiveDigest)
	if !fixture.ObjectiveDigestValidated {
		errs = append(errs, "objective_digest_validated must be true")
	}
	requireField(&errs, "workspace_root", fixture.WorkspaceRoot)
	if !fixture.WorktreeBoundaryValidated {
		errs = append(errs, "worktree_boundary_validated must be true")
	}
	requireField(&errs, "safe_next_node", fixture.SafeNextNode)
	if !fixture.SafeNextNodeSelected {
		errs = append(errs, "safe_next_node_selected must be true")
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}
