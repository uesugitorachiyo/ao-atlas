package atlas

import (
	"fmt"
	"strings"
)

func BuildAtlasMonth3SchemaOwnerRegistryProposal(nodeID, registryManifestPath string) (AtlasMonth3SchemaOwnerRegistryProposal, error) {
	nodeID = strings.TrimSpace(nodeID)
	if nodeID == "" {
		return AtlasMonth3SchemaOwnerRegistryProposal{}, fmt.Errorf("node id is required")
	}
	manifest, err := LoadJSON[AtlasCanonicalContractRegistryManifest](registryManifestPath)
	if err != nil {
		return AtlasMonth3SchemaOwnerRegistryProposal{}, err
	}
	if err := ValidateAtlasCanonicalContractRegistryManifest(manifest); err != nil {
		return AtlasMonth3SchemaOwnerRegistryProposal{}, err
	}
	manifestDigest, err := digestTextFileWithNormalizedLineEndings(registryManifestPath)
	if err != nil {
		return AtlasMonth3SchemaOwnerRegistryProposal{}, err
	}
	contracts := make([]AtlasMonth3SchemaOwnerRegistryContract, 0, len(manifest.Contracts))
	consumerChecks := 0
	for _, contract := range manifest.Contracts {
		checks := make([]string, 0, len(contract.Consumers))
		for _, consumer := range contract.Consumers {
			checks = append(checks, consumer+" consumes "+contract.Schema)
		}
		consumerChecks += len(checks)
		contracts = append(contracts, AtlasMonth3SchemaOwnerRegistryContract{
			ID:                          contract.ID,
			SchemaName:                  contract.Schema,
			RegistryOwner:               "ao-covenant",
			ProducerOwner:               contract.Owner,
			LifecycleClass:              contract.LifecycleClass,
			GateCritical:                contract.GateCritical,
			ConsumerCompatibilityChecks: checks,
		})
	}
	proposal := AtlasMonth3SchemaOwnerRegistryProposal{
		Schema:                                AtlasMonth3SchemaOwnerRegistryProposalContract,
		NodeID:                                nodeID,
		Status:                                "schema_owner_registry_proposal_ready",
		RegistryManifestPath:                  publicArtifactRef(registryManifestPath),
		RegistryManifestDigest:                manifestDigest,
		RegistryAuthorityOwner:                "ao-covenant",
		ContractCount:                         len(contracts),
		ConsumerCompatibilityCheckCount:       consumerChecks,
		Contracts:                             contracts,
		CovenantOwnsRegistry:                  true,
		ProducersRetainContractImplementation: true,
		ConsumerCompatibilityRequired:         true,
		SchedulesWork:                         false,
		ExecutesWork:                          false,
		ApprovesWork:                          false,
		ClaimsAuthorityAdvance:                manifest.ClaimsAuthorityAdvance,
		RSIRemainsDenied:                      manifest.RSIRemainsDenied,
	}
	if !proposal.CovenantOwnsRegistry || proposal.ClaimsAuthorityAdvance || !proposal.RSIRemainsDenied {
		proposal.Status = "schema_owner_registry_proposal_failed"
	}
	if err := ValidateAtlasMonth3SchemaOwnerRegistryProposal(proposal); err != nil {
		return AtlasMonth3SchemaOwnerRegistryProposal{}, err
	}
	return proposal, nil
}

func ValidateAtlasMonth3SchemaOwnerRegistryProposal(proposal AtlasMonth3SchemaOwnerRegistryProposal) error {
	var errs []string
	requireContract(&errs, "month3_schema_owner_registry_proposal", proposal.Schema, AtlasMonth3SchemaOwnerRegistryProposalContract)
	requireField(&errs, "node_id", proposal.NodeID)
	checkPublicPath(&errs, "node_id", proposal.NodeID, true)
	if !oneOf(proposal.Status, "schema_owner_registry_proposal_ready", "schema_owner_registry_proposal_failed") {
		errs = append(errs, "status must be schema_owner_registry_proposal_ready or schema_owner_registry_proposal_failed")
	}
	requireField(&errs, "registry_manifest_path", proposal.RegistryManifestPath)
	checkPublicPath(&errs, "registry_manifest_path", proposal.RegistryManifestPath, true)
	if !digestPattern.MatchString(proposal.RegistryManifestDigest) {
		errs = append(errs, "registry_manifest_digest must be sha256 digest")
	}
	if proposal.RegistryAuthorityOwner != "ao-covenant" {
		errs = append(errs, "registry_authority_owner must be ao-covenant")
	}
	if proposal.ContractCount != len(proposal.Contracts) || proposal.ContractCount == 0 {
		errs = append(errs, "contract_count must match non-empty contracts")
	}
	if !proposal.CovenantOwnsRegistry {
		errs = append(errs, "covenant_owns_registry must be true")
	}
	if !proposal.ProducersRetainContractImplementation {
		errs = append(errs, "producers_retain_contract_implementation must be true")
	}
	if !proposal.ConsumerCompatibilityRequired {
		errs = append(errs, "consumer_compatibility_required must be true")
	}
	checkCount := 0
	seen := map[string]bool{}
	for i, contract := range proposal.Contracts {
		prefix := fmt.Sprintf("contracts[%d]", i)
		requireField(&errs, prefix+".id", contract.ID)
		requireField(&errs, prefix+".schema_name", contract.SchemaName)
		if seen[contract.ID] {
			errs = append(errs, prefix+".id must be unique")
		}
		seen[contract.ID] = true
		if contract.RegistryOwner != "ao-covenant" {
			errs = append(errs, prefix+".registry_owner must be ao-covenant")
		}
		requireField(&errs, prefix+".producer_owner", contract.ProducerOwner)
		if !oneOf(contract.LifecycleClass, "stable", "experimental", "deprecated") {
			errs = append(errs, prefix+".lifecycle_class must be stable, experimental, or deprecated")
		}
		if !contract.GateCritical {
			errs = append(errs, prefix+".gate_critical must be true")
		}
		requireList(&errs, prefix+".consumer_compatibility_checks", contract.ConsumerCompatibilityChecks)
		checkCount += len(contract.ConsumerCompatibilityChecks)
	}
	if proposal.ConsumerCompatibilityCheckCount != checkCount || checkCount == 0 {
		errs = append(errs, "consumer_compatibility_check_count must match non-empty compatibility checks")
	}
	validateNoAuthorityEffects(&errs, proposal.SchedulesWork, proposal.ExecutesWork, proposal.ApprovesWork, proposal.ClaimsAuthorityAdvance, proposal.RSIRemainsDenied)
	return joinErrors(errs)
}

func WriteAtlasMonth3SchemaOwnerRegistryProposal(path string, proposal AtlasMonth3SchemaOwnerRegistryProposal) error {
	return WriteJSON(path, proposal)
}
