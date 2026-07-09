package atlas

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func BuildAtlasPublicSafetyReadbackBinding(readbackPath, sentinelPath, verificationPath, nodeID string) (AtlasPublicSafetyReadbackBinding, error) {
	readback, err := LoadJSON[AtlasRecommendationReadback](readbackPath)
	if err != nil {
		return AtlasPublicSafetyReadbackBinding{}, err
	}
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		return AtlasPublicSafetyReadbackBinding{}, err
	}
	sentinel, err := LoadJSON[AtlasNodeSentinelPublicSafetyEvidence](sentinelPath)
	if err != nil {
		return AtlasPublicSafetyReadbackBinding{}, err
	}
	if err := ValidateAtlasNodeSentinelPublicSafetyEvidence(sentinel); err != nil {
		return AtlasPublicSafetyReadbackBinding{}, err
	}
	if sentinel.Status != "passed" || sentinel.UnsafePublicClaimDetected || sentinel.PromotionClaimDetected || sentinel.RSIClaimDetected || !sentinel.RSIRemainsDenied {
		return AtlasPublicSafetyReadbackBinding{}, fmt.Errorf("sentinel public-safety evidence must be passed with no unsafe, promotion, or RSI claim")
	}
	verification, verificationDigest, err := loadVerificationSummary(verificationPath)
	if err != nil {
		return AtlasPublicSafetyReadbackBinding{}, err
	}
	if !verificationPassedForPublicSafety(verification) {
		return AtlasPublicSafetyReadbackBinding{}, fmt.Errorf("verification summary must prove passed public-safety verification")
	}
	sourceDigest, err := digestJSONArtifact(readbackPath)
	if err != nil {
		return AtlasPublicSafetyReadbackBinding{}, err
	}
	sentinelDigest, err := digestJSONArtifact(sentinelPath)
	if err != nil {
		return AtlasPublicSafetyReadbackBinding{}, err
	}
	binding := AtlasPublicSafetyReadbackBinding{
		Schema:                         AtlasPublicSafetyReadbackBindingContract,
		NodeID:                         strings.TrimSpace(nodeID),
		Status:                         "bound",
		SourceReadbackPath:             publicArtifactRef(readbackPath),
		SourceReadbackDigest:           sourceDigest,
		SentinelEvidencePath:           publicArtifactRef(sentinelPath),
		SentinelEvidenceDigest:         sentinelDigest,
		VerificationSummaryPath:        publicArtifactRef(verificationPath),
		VerificationSummaryDigest:      verificationDigest,
		BoundPublicSafetyScanStatus:    "passed",
		PreviousPublicSafetyScanStatus: readback.PublicSafetyScanStatus,
		ReadyNodesAfterBinding:         readback.ReadyNodes,
		FinalResponseAllowedAfter:      readback.FinalResponseAllowed,
		RSIRemainsDenied:               true,
	}
	if err := ValidateAtlasPublicSafetyReadbackBinding(binding); err != nil {
		return AtlasPublicSafetyReadbackBinding{}, err
	}
	return binding, nil
}

func ValidateAtlasPublicSafetyReadbackBinding(binding AtlasPublicSafetyReadbackBinding) error {
	var errs []string
	requireContract(&errs, "public_safety_readback_binding", binding.Schema, AtlasPublicSafetyReadbackBindingContract)
	requireField(&errs, "node_id", binding.NodeID)
	checkPublicPath(&errs, "node_id", binding.NodeID, true)
	if binding.Status != "bound" {
		errs = append(errs, "status must be bound")
	}
	for field, value := range map[string]string{
		"source_readback_path":       binding.SourceReadbackPath,
		"sentinel_evidence_path":     binding.SentinelEvidencePath,
		"verification_summary_path":  binding.VerificationSummaryPath,
		"bound_public_safety_status": binding.BoundPublicSafetyScanStatus,
		"previous_public_safety":     binding.PreviousPublicSafetyScanStatus,
	} {
		requireField(&errs, field, value)
		checkPublicPath(&errs, field, value, true)
	}
	checkOptionalDigest(&errs, "source_readback_digest", binding.SourceReadbackDigest)
	checkOptionalDigest(&errs, "sentinel_evidence_digest", binding.SentinelEvidenceDigest)
	checkOptionalDigest(&errs, "verification_summary_digest", binding.VerificationSummaryDigest)
	if binding.BoundPublicSafetyScanStatus != "passed" {
		errs = append(errs, "bound_public_safety_scan_status must be passed")
	}
	if binding.PreviousPublicSafetyScanStatus == "passed" {
		errs = append(errs, "previous_public_safety_scan_status must show an unbound or pending source state")
	}
	if binding.ReadyNodesAfterBinding < 0 {
		errs = append(errs, "ready_nodes_after_binding must be non-negative")
	}
	if binding.FinalResponseAllowedAfter && binding.ReadyNodesAfterBinding != 0 {
		errs = append(errs, "final_response_allowed_after_binding must remain false while feature-depth ready work remains")
	}
	if !binding.RSIRemainsDenied {
		errs = append(errs, "rsi_remains_denied must be true")
	}
	return joinErrors(errs)
}

func WriteAtlasPublicSafetyReadbackBinding(path string, binding AtlasPublicSafetyReadbackBinding) error {
	return WriteJSON(path, binding)
}

func loadVerificationSummary(path string) (map[string]any, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	var value map[string]any
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, "", err
	}
	return value, digestValue(value), nil
}

func verificationPassedForPublicSafety(verification map[string]any) bool {
	status, _ := verification["status"].(string)
	if status != "passed" {
		return false
	}
	if passed, ok := verification["public_safety_scan_passed"].(bool); ok && passed {
		return true
	}
	commands, _ := verification["commands"].([]any)
	for _, raw := range commands {
		item, _ := raw.(map[string]any)
		command, _ := item["command"].(string)
		commandStatus, _ := item["status"].(string)
		if commandStatus == "passed" && strings.Contains(strings.ToLower(command), "public-safety") {
			return true
		}
	}
	return false
}
