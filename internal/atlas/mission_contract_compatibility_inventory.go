package atlas

import "fmt"

func BuildAtlasContractCompatibilityInventory() (AtlasContractCompatibilityInventory, error) {
	entries := []AtlasContractCompatibilityEntry{
		{
			ID:           "blueprint-build-authorization",
			SchemaName:   "ao.blueprint.build-authorization.v0.1",
			Owner:        "ao-blueprint",
			GateCritical: true,
			ConsumerTests: []string{
				"ao-atlas: blueprint-import compatibility tests",
				"ao-foundry: blueprint handoff compatibility tests",
			},
		},
		{
			ID:           "ao2-event-hash-vectors",
			SchemaName:   "ao2.event-hash-vectors.v1",
			Owner:        "ao2",
			GateCritical: true,
			ConsumerTests: []string{
				"ao-covenant: event policy digest replay tests",
				"ao-command: class decision readback digest tests",
			},
		},
		{
			ID:           "bounded-signer-contract",
			SchemaName:   AtlasBoundedSignerContractFixtureContract,
			Owner:        "ao-covenant",
			GateCritical: true,
			ConsumerTests: []string{
				"ao-promoter: signed assurance dry-run tests",
				"ao-sentinel: signed evidence freshness tests",
			},
		},
		{
			ID:           "foundry-import",
			SchemaName:   FoundryImportContract,
			Owner:        "ao-atlas",
			GateCritical: true,
			ConsumerTests: []string{
				"ao-foundry: atlas import validation tests",
				"ao-mission: continuation readback tests",
				"ao-command: compact timeline readback tests",
			},
		},
	}
	inventory := AtlasContractCompatibilityInventory{
		Schema:                 AtlasContractCompatibilityInventoryContract,
		Status:                 "compatibility_inventory_ready",
		Contracts:              entries,
		SchedulesWork:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
		ClaimsAuthorityAdvance: false,
		RSIRemainsDenied:       true,
	}
	summarizeContractCompatibilityInventory(&inventory)
	if err := ValidateAtlasContractCompatibilityInventory(inventory); err != nil {
		return AtlasContractCompatibilityInventory{}, err
	}
	return inventory, nil
}

func ValidateAtlasContractCompatibilityInventory(inventory AtlasContractCompatibilityInventory) error {
	var errs []string
	requireContract(&errs, "contract_compatibility_inventory", inventory.Schema, AtlasContractCompatibilityInventoryContract)
	if inventory.Status != "compatibility_inventory_ready" {
		errs = append(errs, "status must be compatibility_inventory_ready")
	}
	expected := inventory
	summarizeContractCompatibilityInventory(&expected)
	if inventory.ContractCount != expected.ContractCount {
		errs = append(errs, "contract_count must match contracts")
	}
	if inventory.GateCriticalCount != expected.GateCriticalCount {
		errs = append(errs, "gate_critical_count must match contracts")
	}
	if inventory.ConsumerTestCount != expected.ConsumerTestCount {
		errs = append(errs, "consumer_test_count must match contracts")
	}
	if inventory.MissingOwnerCount != expected.MissingOwnerCount {
		errs = append(errs, "missing_owner_count must match contracts")
	}
	if inventory.MissingConsumerTestCount != expected.MissingConsumerTestCount {
		errs = append(errs, "missing_consumer_test_count must match contracts")
	}
	if inventory.ContractCount == 0 {
		errs = append(errs, "contracts must not be empty")
	}
	for i, contract := range inventory.Contracts {
		prefix := fmt.Sprintf("contracts[%d]", i)
		requireField(&errs, prefix+".id", contract.ID)
		requireField(&errs, prefix+".schema_name", contract.SchemaName)
		if contract.GateCritical {
			requireField(&errs, prefix+".owner", contract.Owner)
			if len(contract.ConsumerTests) == 0 {
				errs = append(errs, prefix+".consumer_tests must not be empty")
			}
		}
		for j, test := range contract.ConsumerTests {
			requireField(&errs, fmt.Sprintf("%s.consumer_tests[%d]", prefix, j), test)
		}
	}
	if inventory.MissingOwnerCount != 0 {
		errs = append(errs, "missing_owner_count must be zero")
	}
	if inventory.MissingConsumerTestCount != 0 {
		errs = append(errs, "missing_consumer_test_count must be zero")
	}
	validateNoAuthorityEffects(&errs, inventory.SchedulesWork, inventory.ExecutesWork, inventory.ApprovesWork, inventory.ClaimsAuthorityAdvance, inventory.RSIRemainsDenied)
	return joinErrors(errs)
}

func summarizeContractCompatibilityInventory(inventory *AtlasContractCompatibilityInventory) {
	inventory.ContractCount = len(inventory.Contracts)
	inventory.GateCriticalCount = 0
	inventory.ConsumerTestCount = 0
	inventory.MissingOwnerCount = 0
	inventory.MissingConsumerTestCount = 0
	for _, contract := range inventory.Contracts {
		if !contract.GateCritical {
			continue
		}
		inventory.GateCriticalCount++
		if contract.Owner == "" {
			inventory.MissingOwnerCount++
		}
		if len(contract.ConsumerTests) == 0 {
			inventory.MissingConsumerTestCount++
		}
		inventory.ConsumerTestCount += len(contract.ConsumerTests)
	}
}
