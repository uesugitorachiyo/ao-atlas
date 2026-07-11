package atlas

import (
	"fmt"
	"strings"
)

func BuildAtlasMonth3RollbackReplayNegative(nodeID, sourceReadbackPath string) (AtlasMonth3RollbackReplayNegative, error) {
	nodeID = strings.TrimSpace(nodeID)
	if nodeID == "" {
		return AtlasMonth3RollbackReplayNegative{}, fmt.Errorf("node id is required")
	}
	readback, err := LoadJSON[AtlasRecommendationReadback](sourceReadbackPath)
	if err != nil {
		return AtlasMonth3RollbackReplayNegative{}, err
	}
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		return AtlasMonth3RollbackReplayNegative{}, err
	}
	readbackDigest, err := digestTextFileWithNormalizedLineEndings(sourceReadbackPath)
	if err != nil {
		return AtlasMonth3RollbackReplayNegative{}, err
	}
	cases := []AtlasMonth3RollbackReplayCase{
		{Name: "stale_base_commit", ReplayClass: "stale_base_commit", ExpectedState: "rejected_before_rollback_replay", Accepted: false, Replayable: true},
		{Name: "receipt_digest_mismatch", ReplayClass: "receipt_digest_mismatch", ExpectedState: "rejected_before_rollback_replay", Accepted: false, Replayable: true},
	}
	fixture := AtlasMonth3RollbackReplayNegative{
		Schema:                        AtlasMonth3RollbackReplayNegativeContract,
		NodeID:                        nodeID,
		Status:                        "rollback_replay_negative_ready",
		SourceReadbackPath:            publicArtifactRef(sourceReadbackPath),
		SourceReadbackDigest:          readbackDigest,
		Cases:                         cases,
		CaseCount:                     len(cases),
		AcceptedCaseCount:             month3AcceptedRollbackReplayCaseCount(cases),
		StaleBaseCommitRejected:       month3RollbackReplayCaseRejected(cases, "stale_base_commit"),
		ReceiptDigestMismatchRejected: month3RollbackReplayCaseRejected(cases, "receipt_digest_mismatch"),
		FinalResponseAllowed:          readback.FinalResponseAllowed,
		SchedulesWork:                 false,
		ExecutesWork:                  false,
		ApprovesWork:                  false,
		ClaimsAuthorityAdvance:        false,
		RSIRemainsDenied:              readback.SafetyBoundaries["rsi_remains_denied"],
	}
	if fixture.AcceptedCaseCount != 0 || !fixture.StaleBaseCommitRejected || !fixture.ReceiptDigestMismatchRejected {
		fixture.Status = "rollback_replay_negative_failed"
	}
	if err := ValidateAtlasMonth3RollbackReplayNegative(fixture); err != nil {
		return AtlasMonth3RollbackReplayNegative{}, err
	}
	return fixture, nil
}

func month3AcceptedRollbackReplayCaseCount(cases []AtlasMonth3RollbackReplayCase) int {
	count := 0
	for _, item := range cases {
		if item.Accepted {
			count++
		}
	}
	return count
}

func month3RollbackReplayCaseRejected(cases []AtlasMonth3RollbackReplayCase, replayClass string) bool {
	for _, item := range cases {
		if item.ReplayClass == replayClass {
			return !item.Accepted && item.ExpectedState == "rejected_before_rollback_replay"
		}
	}
	return false
}

func ValidateAtlasMonth3RollbackReplayNegative(fixture AtlasMonth3RollbackReplayNegative) error {
	var errs []string
	requireContract(&errs, "month3_rollback_replay_negative", fixture.Schema, AtlasMonth3RollbackReplayNegativeContract)
	requireField(&errs, "node_id", fixture.NodeID)
	checkPublicPath(&errs, "node_id", fixture.NodeID, true)
	if !oneOf(fixture.Status, "rollback_replay_negative_ready", "rollback_replay_negative_failed") {
		errs = append(errs, "status must be rollback_replay_negative_ready or rollback_replay_negative_failed")
	}
	requireField(&errs, "source_readback_path", fixture.SourceReadbackPath)
	checkPublicPath(&errs, "source_readback_path", fixture.SourceReadbackPath, true)
	if !digestPattern.MatchString(fixture.SourceReadbackDigest) {
		errs = append(errs, "source_readback_digest must be sha256 digest")
	}
	if fixture.CaseCount != len(fixture.Cases) || fixture.CaseCount != 2 {
		errs = append(errs, "case_count must match two rollback replay negative cases")
	}
	if fixture.AcceptedCaseCount != 0 {
		errs = append(errs, "accepted_case_count must be zero")
	}
	if !fixture.StaleBaseCommitRejected {
		errs = append(errs, "stale_base_commit_rejected must be true")
	}
	if !fixture.ReceiptDigestMismatchRejected {
		errs = append(errs, "receipt_digest_mismatch_rejected must be true")
	}
	required := map[string]bool{"stale_base_commit": false, "receipt_digest_mismatch": false}
	for i, item := range fixture.Cases {
		prefix := fmt.Sprintf("cases[%d]", i)
		requireField(&errs, prefix+".name", item.Name)
		requireField(&errs, prefix+".replay_class", item.ReplayClass)
		requireField(&errs, prefix+".expected_state", item.ExpectedState)
		if _, ok := required[item.ReplayClass]; ok {
			required[item.ReplayClass] = true
		}
		if item.Accepted {
			errs = append(errs, prefix+".accepted must be false")
		}
		if !item.Replayable {
			errs = append(errs, prefix+".replayable must be true")
		}
	}
	for name, seen := range required {
		if !seen {
			errs = append(errs, name+" case is required")
		}
	}
	if fixture.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must be false while ready work remains")
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}

func WriteAtlasMonth3RollbackReplayNegative(path string, fixture AtlasMonth3RollbackReplayNegative) error {
	return WriteJSON(path, fixture)
}
