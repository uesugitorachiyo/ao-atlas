package atlas

import "fmt"

func BuildAtlasSentinelHostedCIWorkflowFixture() (AtlasSentinelHostedCIWorkflowFixture, error) {
	commands := []string{
		"go test ./... -count=1",
		"go vet ./...",
		"go run ./cmd/sentinel fixture verify",
	}
	fixture := AtlasSentinelHostedCIWorkflowFixture{
		Schema:                       AtlasSentinelHostedCIWorkflowFixtureContract,
		Status:                       "least_privilege_workflow_ready",
		WorkflowName:                 "sentinel-fixture-verification",
		WorkflowPath:                 ".github/workflows/sentinel-fixture-verification.yml",
		Permissions:                  "contents:read",
		PermissionsReadOnly:          true,
		DeterministicFixtureCommands: commands,
		CommandCount:                 len(commands),
		UsesProviderCredentials:      false,
		UsesSecrets:                  false,
		TriggersRelease:              false,
		SchedulesWork:                false,
		ExecutesWork:                 false,
		ApprovesWork:                 false,
		ClaimsAuthorityAdvance:       false,
		RSIRemainsDenied:             true,
	}
	if err := ValidateAtlasSentinelHostedCIWorkflowFixture(fixture); err != nil {
		return AtlasSentinelHostedCIWorkflowFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasSentinelHostedCIWorkflowFixture(fixture AtlasSentinelHostedCIWorkflowFixture) error {
	var errs []string
	requireContract(&errs, "sentinel_hosted_ci_workflow_fixture", fixture.Schema, AtlasSentinelHostedCIWorkflowFixtureContract)
	if fixture.Status != "least_privilege_workflow_ready" {
		errs = append(errs, "status must be least_privilege_workflow_ready")
	}
	requireField(&errs, "workflow_name", fixture.WorkflowName)
	requireField(&errs, "workflow_path", fixture.WorkflowPath)
	if fixture.Permissions != "contents:read" {
		errs = append(errs, "permissions must be contents:read")
	}
	if !fixture.PermissionsReadOnly {
		errs = append(errs, "permissions_read_only must be true")
	}
	if fixture.CommandCount != len(fixture.DeterministicFixtureCommands) {
		errs = append(errs, "command_count must match deterministic_fixture_commands")
	}
	if fixture.CommandCount != 3 {
		errs = append(errs, "command_count must be 3")
	}
	for i, command := range fixture.DeterministicFixtureCommands {
		requireField(&errs, fmt.Sprintf("deterministic_fixture_commands[%d]", i), command)
	}
	if fixture.UsesProviderCredentials {
		errs = append(errs, "uses_provider_credentials must be false")
	}
	if fixture.UsesSecrets {
		errs = append(errs, "uses_secrets must be false")
	}
	if fixture.TriggersRelease {
		errs = append(errs, "triggers_release must be false")
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}
