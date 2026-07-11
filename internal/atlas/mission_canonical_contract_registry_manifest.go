package atlas

import "fmt"

func BuildAtlasCanonicalContractRegistryManifest() (AtlasCanonicalContractRegistryManifest, error) {
	manifest := AtlasCanonicalContractRegistryManifest{
		Schema:          AtlasCanonicalContractRegistryManifestContract,
		Status:          "canonical_contract_registry_ready",
		ManifestPurpose: "gate_critical_contract_owner_lifecycle_digest_consumer_index",
		Contracts: []AtlasCanonicalContractRegistryEntry{
			{
				ID:             "blueprint-build-authorization",
				Schema:         "ao.blueprint.build-authorization.v0.1",
				Owner:          "ao-blueprint",
				LifecycleClass: "stable",
				Digest:         "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				GateCritical:   true,
				Consumers:      []string{"ao-atlas", "ao-foundry"},
			},
			{
				ID:             "ao2-event-hash-vectors",
				Schema:         "ao2.event-hash-vectors.v1",
				Owner:          "ao2",
				LifecycleClass: "stable",
				Digest:         "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
				GateCritical:   true,
				Consumers:      []string{"ao-covenant", "ao-command"},
			},
			{
				ID:             "bounded-signer-contract",
				Schema:         AtlasBoundedSignerContractFixtureContract,
				Owner:          "ao-covenant",
				LifecycleClass: "experimental",
				Digest:         "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
				GateCritical:   true,
				Consumers:      []string{"ao-promoter", "ao-sentinel"},
			},
			{
				ID:             "foundry-import",
				Schema:         FoundryImportContract,
				Owner:          "ao-atlas",
				LifecycleClass: "stable",
				Digest:         "sha256:dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
				GateCritical:   true,
				Consumers:      []string{"ao-foundry", "ao-mission", "ao-command"},
			},
		},
		SchedulesWork:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
		ClaimsAuthorityAdvance: false,
		RSIRemainsDenied:       true,
	}
	manifest.ContractCount = len(manifest.Contracts)
	for _, contract := range manifest.Contracts {
		if contract.GateCritical {
			manifest.GateCriticalCount++
		}
		manifest.ConsumerCount += len(contract.Consumers)
	}
	if err := ValidateAtlasCanonicalContractRegistryManifest(manifest); err != nil {
		return AtlasCanonicalContractRegistryManifest{}, err
	}
	return manifest, nil
}

func ValidateAtlasCanonicalContractRegistryManifest(manifest AtlasCanonicalContractRegistryManifest) error {
	var errs []string
	requireContract(&errs, "canonical_contract_registry_manifest", manifest.Schema, AtlasCanonicalContractRegistryManifestContract)
	if manifest.Status != "canonical_contract_registry_ready" {
		errs = append(errs, "status must be canonical_contract_registry_ready")
	}
	requireField(&errs, "manifest_purpose", manifest.ManifestPurpose)
	if manifest.ContractCount != len(manifest.Contracts) {
		errs = append(errs, "contract_count must match contracts")
	}
	if manifest.ContractCount == 0 {
		errs = append(errs, "contracts must not be empty")
	}
	gateCriticalCount := 0
	consumerCount := 0
	for i, contract := range manifest.Contracts {
		prefix := fmt.Sprintf("contracts[%d]", i)
		requireField(&errs, prefix+".id", contract.ID)
		requireField(&errs, prefix+".schema", contract.Schema)
		requireField(&errs, prefix+".owner", contract.Owner)
		if !oneOf(contract.LifecycleClass, "stable", "experimental", "deprecated") {
			errs = append(errs, prefix+".lifecycle_class must be stable, experimental, or deprecated")
		}
		validateRejectedTicketDigest(&errs, prefix+".digest", contract.Digest)
		if !contract.GateCritical {
			errs = append(errs, prefix+".gate_critical must be true for node 8 manifest")
		}
		if len(contract.Consumers) == 0 {
			errs = append(errs, prefix+".consumers must not be empty")
		}
		for j, consumer := range contract.Consumers {
			requireField(&errs, fmt.Sprintf("%s.consumers[%d]", prefix, j), consumer)
		}
		if contract.GateCritical {
			gateCriticalCount++
		}
		consumerCount += len(contract.Consumers)
	}
	if manifest.GateCriticalCount != gateCriticalCount {
		errs = append(errs, "gate_critical_count must match contracts")
	}
	if manifest.ConsumerCount != consumerCount {
		errs = append(errs, "consumer_count must match contracts")
	}
	validateNoAuthorityEffects(&errs, manifest.SchedulesWork, manifest.ExecutesWork, manifest.ApprovesWork, manifest.ClaimsAuthorityAdvance, manifest.RSIRemainsDenied)
	return joinErrors(errs)
}
