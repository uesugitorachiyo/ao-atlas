package atlas

import (
	"fmt"
	"strings"
)

func BuildAtlasMonth3ProviderModelProvenance(nodeID, sourceReadbackPath string) (AtlasMonth3ProviderModelProvenance, error) {
	nodeID = strings.TrimSpace(nodeID)
	if nodeID == "" {
		return AtlasMonth3ProviderModelProvenance{}, fmt.Errorf("node id is required")
	}
	readback, err := LoadJSON[AtlasRecommendationReadback](sourceReadbackPath)
	if err != nil {
		return AtlasMonth3ProviderModelProvenance{}, err
	}
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		return AtlasMonth3ProviderModelProvenance{}, err
	}
	readbackDigest, err := digestTextFileWithNormalizedLineEndings(sourceReadbackPath)
	if err != nil {
		return AtlasMonth3ProviderModelProvenance{}, err
	}
	records := []AtlasMonth3ProviderModelRunRecord{
		{ID: "mission_supervisor_prompt", RecordClass: "operator_prompt", Provider: "openai", Model: "gpt-5.6-sol", ModelClass: "frontier_reasoning", ReasoningProfile: "high", LiveProviderCall: false},
		{ID: "atlas_planning_readback", RecordClass: "planning_readback", Provider: "openai", Model: "gpt-5.6-sol", ModelClass: "frontier_reasoning", ReasoningProfile: "high", LiveProviderCall: false},
		{ID: "foundry_bounded_handoff", RecordClass: "bounded_handoff", Provider: "openai", Model: "gpt-5.6-sol", ModelClass: "frontier_reasoning", ReasoningProfile: "high", LiveProviderCall: false},
		{ID: "command_operator_readback", RecordClass: "operator_readback", Provider: "openai", Model: "gpt-5.6-mini", ModelClass: "compact_reasoning", ReasoningProfile: "standard", LiveProviderCall: false},
	}
	fixture := AtlasMonth3ProviderModelProvenance{
		Schema:                 AtlasMonth3ProviderModelProvenanceContract,
		NodeID:                 nodeID,
		Status:                 "provider_model_provenance_ready",
		SourceReadbackPath:     publicArtifactRef(sourceReadbackPath),
		SourceReadbackDigest:   readbackDigest,
		RunRecords:             records,
		RunRecordCount:         len(records),
		EveryRunHasProvider:    month3EveryRunHasProvider(records),
		EveryRunHasModel:       month3EveryRunHasModel(records),
		EveryRunHasModelClass:  month3EveryRunHasModelClass(records),
		LiveProviderCallCount:  month3LiveProviderCallCount(records),
		FinalResponseAllowed:   readback.FinalResponseAllowed,
		SchedulesWork:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
		ClaimsAuthorityAdvance: false,
		RSIRemainsDenied:       readback.SafetyBoundaries["rsi_remains_denied"],
	}
	if !fixture.EveryRunHasProvider || !fixture.EveryRunHasModel || !fixture.EveryRunHasModelClass || fixture.LiveProviderCallCount != 0 {
		fixture.Status = "provider_model_provenance_failed"
	}
	if err := ValidateAtlasMonth3ProviderModelProvenance(fixture); err != nil {
		return AtlasMonth3ProviderModelProvenance{}, err
	}
	return fixture, nil
}

func month3EveryRunHasProvider(records []AtlasMonth3ProviderModelRunRecord) bool {
	for _, record := range records {
		if strings.TrimSpace(record.Provider) == "" {
			return false
		}
	}
	return len(records) > 0
}

func month3EveryRunHasModel(records []AtlasMonth3ProviderModelRunRecord) bool {
	for _, record := range records {
		if strings.TrimSpace(record.Model) == "" {
			return false
		}
	}
	return len(records) > 0
}

func month3EveryRunHasModelClass(records []AtlasMonth3ProviderModelRunRecord) bool {
	for _, record := range records {
		if strings.TrimSpace(record.ModelClass) == "" {
			return false
		}
	}
	return len(records) > 0
}

func month3LiveProviderCallCount(records []AtlasMonth3ProviderModelRunRecord) int {
	count := 0
	for _, record := range records {
		if record.LiveProviderCall {
			count++
		}
	}
	return count
}

func ValidateAtlasMonth3ProviderModelProvenance(fixture AtlasMonth3ProviderModelProvenance) error {
	var errs []string
	requireContract(&errs, "month3_provider_model_provenance", fixture.Schema, AtlasMonth3ProviderModelProvenanceContract)
	requireField(&errs, "node_id", fixture.NodeID)
	checkPublicPath(&errs, "node_id", fixture.NodeID, true)
	if !oneOf(fixture.Status, "provider_model_provenance_ready", "provider_model_provenance_failed") {
		errs = append(errs, "status must be provider_model_provenance_ready or provider_model_provenance_failed")
	}
	requireField(&errs, "source_readback_path", fixture.SourceReadbackPath)
	checkPublicPath(&errs, "source_readback_path", fixture.SourceReadbackPath, true)
	if !digestPattern.MatchString(fixture.SourceReadbackDigest) {
		errs = append(errs, "source_readback_digest must be sha256 digest")
	}
	if fixture.RunRecordCount != len(fixture.RunRecords) || fixture.RunRecordCount != 4 {
		errs = append(errs, "run_record_count must match four model-backed run records")
	}
	if !fixture.EveryRunHasProvider {
		errs = append(errs, "every_run_has_provider must be true")
	}
	if !fixture.EveryRunHasModel {
		errs = append(errs, "every_run_has_model must be true")
	}
	if !fixture.EveryRunHasModelClass {
		errs = append(errs, "every_run_has_model_class must be true")
	}
	if fixture.LiveProviderCallCount != 0 {
		errs = append(errs, "live_provider_call_count must be zero")
	}
	seen := map[string]bool{}
	for i, record := range fixture.RunRecords {
		prefix := fmt.Sprintf("run_records[%d]", i)
		requireField(&errs, prefix+".id", record.ID)
		requireField(&errs, prefix+".record_class", record.RecordClass)
		requireField(&errs, prefix+".provider", record.Provider)
		requireField(&errs, prefix+".model", record.Model)
		requireField(&errs, prefix+".model_class", record.ModelClass)
		requireField(&errs, prefix+".reasoning_profile", record.ReasoningProfile)
		if seen[record.ID] {
			errs = append(errs, prefix+".id must be unique")
		}
		seen[record.ID] = true
		if record.LiveProviderCall {
			errs = append(errs, prefix+".live_provider_call must be false")
		}
	}
	if fixture.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must be false while ready work remains")
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}

func WriteAtlasMonth3ProviderModelProvenance(path string, fixture AtlasMonth3ProviderModelProvenance) error {
	return WriteJSON(path, fixture)
}
