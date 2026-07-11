package atlas

import (
	"fmt"
	"strings"
)

func BuildAtlasMonth3EvidenceExternalizationPlan(nodeID, contentManifestPath string) (AtlasMonth3EvidenceExternalizationPlan, error) {
	nodeID = strings.TrimSpace(nodeID)
	if nodeID == "" {
		return AtlasMonth3EvidenceExternalizationPlan{}, fmt.Errorf("node id is required")
	}
	manifest, err := LoadJSON[AtlasContentAddressedEvidenceManifestFixture](contentManifestPath)
	if err != nil {
		return AtlasMonth3EvidenceExternalizationPlan{}, err
	}
	if err := ValidateAtlasContentAddressedEvidenceManifestFixture(manifest); err != nil {
		return AtlasMonth3EvidenceExternalizationPlan{}, err
	}
	manifestDigest, err := digestTextFileWithNormalizedLineEndings(contentManifestPath)
	if err != nil {
		return AtlasMonth3EvidenceExternalizationPlan{}, err
	}
	externalized := []string{"bulk_campaign_evidence", "long_run_node_records", "ci_run_ledgers"}
	retained := []string{"small_replayable_fixture", "typed_contract_fixture", "operator_readback_fixture"}
	plan := AtlasMonth3EvidenceExternalizationPlan{
		Schema:                          AtlasMonth3EvidenceExternalizationPlanContract,
		NodeID:                          nodeID,
		Status:                          "evidence_externalization_plan_ready",
		ContentManifestPath:             publicArtifactRef(contentManifestPath),
		ContentManifestDigest:           manifestDigest,
		ExternalizedEvidenceClasses:     externalized,
		RetainedFixtureClasses:          retained,
		ExternalizedClassCount:          len(externalized),
		RetainedFixtureClassCount:       len(retained),
		ContentManifestBound:            manifest.Status == "content_addressed_evidence_manifest_ready",
		BulkEvidenceExternalized:        manifest.BulkEvidenceExternalized,
		SmallReplayableFixturesRetained: manifest.SmallReplayableFixturesRetained,
		ContentAddressingRequired:       manifest.ContentAddressingRequired,
		SchedulesWork:                   false,
		ExecutesWork:                    false,
		ApprovesWork:                    false,
		ClaimsAuthorityAdvance:          manifest.ClaimsAuthorityAdvance,
		RSIRemainsDenied:                manifest.RSIRemainsDenied,
	}
	if !plan.ContentManifestBound || !plan.BulkEvidenceExternalized || !plan.SmallReplayableFixturesRetained || plan.ClaimsAuthorityAdvance || !plan.RSIRemainsDenied {
		plan.Status = "evidence_externalization_plan_failed"
	}
	if err := ValidateAtlasMonth3EvidenceExternalizationPlan(plan); err != nil {
		return AtlasMonth3EvidenceExternalizationPlan{}, err
	}
	return plan, nil
}

func ValidateAtlasMonth3EvidenceExternalizationPlan(plan AtlasMonth3EvidenceExternalizationPlan) error {
	var errs []string
	requireContract(&errs, "month3_evidence_externalization_plan", plan.Schema, AtlasMonth3EvidenceExternalizationPlanContract)
	requireField(&errs, "node_id", plan.NodeID)
	checkPublicPath(&errs, "node_id", plan.NodeID, true)
	if !oneOf(plan.Status, "evidence_externalization_plan_ready", "evidence_externalization_plan_failed") {
		errs = append(errs, "status must be evidence_externalization_plan_ready or evidence_externalization_plan_failed")
	}
	requireField(&errs, "content_manifest_path", plan.ContentManifestPath)
	checkPublicPath(&errs, "content_manifest_path", plan.ContentManifestPath, true)
	if !digestPattern.MatchString(plan.ContentManifestDigest) {
		errs = append(errs, "content_manifest_digest must be sha256 digest")
	}
	requireList(&errs, "externalized_evidence_classes", plan.ExternalizedEvidenceClasses)
	requireList(&errs, "retained_fixture_classes", plan.RetainedFixtureClasses)
	if plan.ExternalizedClassCount != len(plan.ExternalizedEvidenceClasses) || plan.ExternalizedClassCount < 3 {
		errs = append(errs, "externalized_class_count must match at least three externalized classes")
	}
	if plan.RetainedFixtureClassCount != len(plan.RetainedFixtureClasses) || plan.RetainedFixtureClassCount < 3 {
		errs = append(errs, "retained_fixture_class_count must match at least three retained fixture classes")
	}
	if !plan.ContentManifestBound {
		errs = append(errs, "content_manifest_bound must be true")
	}
	if !plan.BulkEvidenceExternalized {
		errs = append(errs, "bulk_evidence_externalized must be true")
	}
	if !plan.SmallReplayableFixturesRetained {
		errs = append(errs, "small_replayable_fixtures_retained must be true")
	}
	if !plan.ContentAddressingRequired {
		errs = append(errs, "content_addressing_required must be true")
	}
	validateNoAuthorityEffects(&errs, plan.SchedulesWork, plan.ExecutesWork, plan.ApprovesWork, plan.ClaimsAuthorityAdvance, plan.RSIRemainsDenied)
	return joinErrors(errs)
}

func WriteAtlasMonth3EvidenceExternalizationPlan(path string, plan AtlasMonth3EvidenceExternalizationPlan) error {
	return WriteJSON(path, plan)
}
