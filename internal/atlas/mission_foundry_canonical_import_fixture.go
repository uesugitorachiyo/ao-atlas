package atlas

import (
	"encoding/json"
	"fmt"
)

func ValidateFoundryImportCanonicalEnvelope(raw []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return err
	}
	if _, ok := fields["contract_version"]; !ok {
		return fmt.Errorf("foundry import canonical envelope must use contract_version")
	}
	for _, alias := range []string{"schema", "schema_version", "contractVersion"} {
		if _, ok := fields[alias]; ok {
			return fmt.Errorf("foundry import %s alias is not accepted; use contract_version", alias)
		}
	}
	var foundryImport FoundryImport
	if err := json.Unmarshal(raw, &foundryImport); err != nil {
		return err
	}
	return ValidateFoundryImport(foundryImport)
}

func BuildAtlasFoundryCanonicalImportFixture(workgraphPath string, expectedNode string) (AtlasFoundryCanonicalImportFixture, error) {
	workgraph, err := LoadJSON[Workgraph](workgraphPath)
	if err != nil {
		return AtlasFoundryCanonicalImportFixture{}, err
	}
	foundryImport, err := BuildFoundryImportForNodes(workgraph, []string{expectedNode}, nil)
	if err != nil {
		return AtlasFoundryCanonicalImportFixture{}, err
	}
	if err := ValidateFoundryImportMatchesWorkgraph(workgraph, foundryImport); err != nil {
		return AtlasFoundryCanonicalImportFixture{}, err
	}
	raw, err := json.Marshal(foundryImport)
	if err != nil {
		return AtlasFoundryCanonicalImportFixture{}, err
	}
	if err := ValidateFoundryImportCanonicalEnvelope(raw); err != nil {
		return AtlasFoundryCanonicalImportFixture{}, err
	}
	rejected := []AtlasFoundryRejectedSchemaAlias{}
	for _, alias := range []string{"schema", "schema_version", "contractVersion"} {
		err := validateFoundryImportWithAlias(foundryImport, alias)
		if err == nil {
			return AtlasFoundryCanonicalImportFixture{}, fmt.Errorf("foundry import alias %s was not rejected", alias)
		}
		rejected = append(rejected, AtlasFoundryRejectedSchemaAlias{
			Alias:  alias,
			Status: "rejected",
			Reason: err.Error(),
		})
	}
	task := foundryImport.Tasks[0]
	fixture := AtlasFoundryCanonicalImportFixture{
		Schema:                           AtlasFoundryCanonicalImportFixtureContract,
		Status:                           "canonical_import_ready",
		WorkgraphID:                      workgraph.ID,
		TargetInstance:                   workgraph.TargetInstance,
		ExpectedNode:                     expectedNode,
		ExpectedTask:                     task.TaskID,
		CanonicalContractField:           "contract_version",
		AcceptedCanonicalImport:          true,
		CanonicalWorkgraphFieldsConsumed: foundryImport.WorkgraphID == workgraph.ID && foundryImport.TargetInstance == workgraph.TargetInstance && task.NodeID == expectedNode,
		RejectedAliases:                  rejected,
		TaskHash:                         task.TaskHash,
		SchedulesWork:                    false,
		ExecutesWork:                     false,
		ApprovesWork:                     false,
		ClaimsAuthorityAdvance:           false,
		RSIRemainsDenied:                 true,
	}
	fixture.RejectedAliasCount = len(fixture.RejectedAliases)
	if err := ValidateAtlasFoundryCanonicalImportFixture(fixture); err != nil {
		return AtlasFoundryCanonicalImportFixture{}, err
	}
	return fixture, nil
}

func validateFoundryImportWithAlias(foundryImport FoundryImport, alias string) error {
	body, err := json.Marshal(foundryImport)
	if err != nil {
		return err
	}
	var fields map[string]any
	if err := json.Unmarshal(body, &fields); err != nil {
		return err
	}
	fields[alias] = FoundryImportContract
	aliased, err := json.Marshal(fields)
	if err != nil {
		return err
	}
	return ValidateFoundryImportCanonicalEnvelope(aliased)
}

func ValidateAtlasFoundryCanonicalImportFixture(fixture AtlasFoundryCanonicalImportFixture) error {
	var errs []string
	requireContract(&errs, "foundry_canonical_import_fixture", fixture.Schema, AtlasFoundryCanonicalImportFixtureContract)
	if fixture.Status != "canonical_import_ready" {
		errs = append(errs, "status must be canonical_import_ready")
	}
	for field, value := range map[string]string{
		"workgraph_id":             fixture.WorkgraphID,
		"target_instance":          fixture.TargetInstance,
		"expected_node":            fixture.ExpectedNode,
		"expected_task":            fixture.ExpectedTask,
		"canonical_contract_field": fixture.CanonicalContractField,
		"task_hash":                fixture.TaskHash,
	} {
		requireField(&errs, field, value)
	}
	if fixture.CanonicalContractField != "contract_version" {
		errs = append(errs, "canonical_contract_field must be contract_version")
	}
	if !fixture.AcceptedCanonicalImport {
		errs = append(errs, "accepted_canonical_import must be true")
	}
	if !fixture.CanonicalWorkgraphFieldsConsumed {
		errs = append(errs, "canonical_workgraph_fields_consumed must be true")
	}
	if fixture.RejectedAliasCount != len(fixture.RejectedAliases) {
		errs = append(errs, "rejected_alias_count must match rejected_aliases")
	}
	if fixture.RejectedAliasCount < 3 {
		errs = append(errs, "rejected_alias_count must cover schema, schema_version, and contractVersion")
	}
	for i, alias := range fixture.RejectedAliases {
		prefix := fmt.Sprintf("rejected_aliases[%d]", i)
		requireField(&errs, prefix+".alias", alias.Alias)
		if alias.Status != "rejected" {
			errs = append(errs, prefix+".status must be rejected")
		}
		requireField(&errs, prefix+".reason", alias.Reason)
	}
	validateRejectedTicketDigest(&errs, "task_hash", fixture.TaskHash)
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}
