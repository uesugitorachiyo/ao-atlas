package atlas

import (
	"fmt"
	"strings"
)

func BuildAtlasMonth3ControlPlaneObserverBinding(nodeID, adapterFixturePath, sourceReadbackPath string) (AtlasMonth3ControlPlaneObserverBinding, error) {
	nodeID = strings.TrimSpace(nodeID)
	if nodeID == "" {
		return AtlasMonth3ControlPlaneObserverBinding{}, fmt.Errorf("node id is required")
	}
	adapter, err := LoadJSON[AtlasCommandReadbackAdapterBoundaryFixture](adapterFixturePath)
	if err != nil {
		return AtlasMonth3ControlPlaneObserverBinding{}, err
	}
	if err := ValidateAtlasCommandReadbackAdapterBoundaryFixture(adapter); err != nil {
		return AtlasMonth3ControlPlaneObserverBinding{}, err
	}
	readback, err := LoadJSON[AtlasRecommendationReadback](sourceReadbackPath)
	if err != nil {
		return AtlasMonth3ControlPlaneObserverBinding{}, err
	}
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		return AtlasMonth3ControlPlaneObserverBinding{}, err
	}
	adapterDigest, err := digestTextFileWithNormalizedLineEndings(adapterFixturePath)
	if err != nil {
		return AtlasMonth3ControlPlaneObserverBinding{}, err
	}
	readbackDigest, err := digestTextFileWithNormalizedLineEndings(sourceReadbackPath)
	if err != nil {
		return AtlasMonth3ControlPlaneObserverBinding{}, err
	}
	binding := AtlasMonth3ControlPlaneObserverBinding{
		Schema:                         AtlasMonth3ControlPlaneObserverBindingContract,
		NodeID:                         nodeID,
		Status:                         "control_plane_observer_bound",
		AdapterFixturePath:             publicArtifactRef(adapterFixturePath),
		AdapterFixtureDigest:           adapterDigest,
		SourceReadbackPath:             publicArtifactRef(sourceReadbackPath),
		SourceReadbackDigest:           readbackDigest,
		AdapterDelegatesToControlPlane: month3AdapterDelegatesTo(adapter, "ao2-control-plane"),
		MissionTimelineReadbackBound:   readback.CompletedNodes > 0 && readback.ReadyNodes > 0,
		ObserverReadOnly:               adapter.PresentationOnly && adapter.DelegatesDomainDecisions,
		DuplicatesDomainDecisions:      adapter.DuplicatesDomainDecisions,
		CompletedNodes:                 readback.CompletedNodes,
		ReadyNodes:                     readback.ReadyNodes,
		BlockedNodes:                   readback.BlockedNodes,
		FailedNodes:                    readback.FailedNodes,
		NextExecutableNode:             readback.FirstExecutableNode,
		FinalResponseAllowed:           readback.FinalResponseAllowed,
		SchedulesWork:                  false,
		ExecutesWork:                   false,
		ApprovesWork:                   false,
		ClaimsAuthorityAdvance:         adapter.ClaimsAuthorityAdvance,
		RSIRemainsDenied:               adapter.RSIRemainsDenied && readback.SafetyBoundaries["rsi_remains_denied"],
	}
	if !binding.AdapterDelegatesToControlPlane ||
		!binding.MissionTimelineReadbackBound ||
		!binding.ObserverReadOnly ||
		binding.DuplicatesDomainDecisions ||
		binding.ClaimsAuthorityAdvance ||
		!binding.RSIRemainsDenied {
		binding.Status = "control_plane_observer_failed"
	}
	if err := ValidateAtlasMonth3ControlPlaneObserverBinding(binding); err != nil {
		return AtlasMonth3ControlPlaneObserverBinding{}, err
	}
	return binding, nil
}

func month3AdapterDelegatesTo(fixture AtlasCommandReadbackAdapterBoundaryFixture, target string) bool {
	for _, adapter := range fixture.Adapters {
		if adapter.DelegatesDecisionTo == target && adapter.DomainDecisionReadOnly {
			return true
		}
	}
	return false
}

func ValidateAtlasMonth3ControlPlaneObserverBinding(binding AtlasMonth3ControlPlaneObserverBinding) error {
	var errs []string
	requireContract(&errs, "month3_control_plane_observer_binding", binding.Schema, AtlasMonth3ControlPlaneObserverBindingContract)
	requireField(&errs, "node_id", binding.NodeID)
	checkPublicPath(&errs, "node_id", binding.NodeID, true)
	if !oneOf(binding.Status, "control_plane_observer_bound", "control_plane_observer_failed") {
		errs = append(errs, "status must be control_plane_observer_bound or control_plane_observer_failed")
	}
	for field, value := range map[string]string{
		"adapter_fixture_path": binding.AdapterFixturePath,
		"source_readback_path": binding.SourceReadbackPath,
	} {
		requireField(&errs, field, value)
		checkPublicPath(&errs, field, value, true)
	}
	for field, value := range map[string]string{
		"adapter_fixture_digest": binding.AdapterFixtureDigest,
		"source_readback_digest": binding.SourceReadbackDigest,
	} {
		if !digestPattern.MatchString(value) {
			errs = append(errs, field+" must be sha256 digest")
		}
	}
	if !binding.AdapterDelegatesToControlPlane {
		errs = append(errs, "adapter_delegates_to_control_plane must be true")
	}
	if !binding.MissionTimelineReadbackBound {
		errs = append(errs, "mission_timeline_readback_bound must be true")
	}
	if !binding.ObserverReadOnly {
		errs = append(errs, "observer_read_only must be true")
	}
	if binding.DuplicatesDomainDecisions {
		errs = append(errs, "duplicates_domain_decisions must be false")
	}
	if binding.CompletedNodes <= 0 || binding.ReadyNodes <= 0 || binding.BlockedNodes != 0 || binding.FailedNodes != 0 {
		errs = append(errs, "readback counts must show completed and ready work with zero blocked or failed nodes")
	}
	requireField(&errs, "next_executable_node", binding.NextExecutableNode)
	if binding.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must be false while ready work remains")
	}
	validateNoAuthorityEffects(&errs, binding.SchedulesWork, binding.ExecutesWork, binding.ApprovesWork, binding.ClaimsAuthorityAdvance, binding.RSIRemainsDenied)
	return joinErrors(errs)
}

func WriteAtlasMonth3ControlPlaneObserverBinding(path string, binding AtlasMonth3ControlPlaneObserverBinding) error {
	return WriteJSON(path, binding)
}
