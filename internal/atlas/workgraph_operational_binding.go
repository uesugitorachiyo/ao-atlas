package atlas

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
)

const (
	OperationalWorkgraphBindingDocumentContract       = "ao.atlas.workgraph-operational-binding.v0.1"
	OperationalWorkgraphBindingReadbackContract       = "ao.atlas.workgraph-operational-binding-readback.v0.1"
	OperationalWorkgraphBindingDigestReadbackContract = "ao.atlas.workgraph-operational-binding-digest-readback.v0.1"
)

var sourceHeadPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)

func workgraphRequiresOperationalBinding(workgraph Workgraph) bool {
	if workgraph.OperationalBinding != nil {
		return true
	}
	for _, node := range workgraph.Nodes {
		if node.OperationalBinding != nil {
			return true
		}
	}
	return false
}

// LoadWorkgraph keeps legacy v0.1 documents permissive while making a claimed
// operational binding strict. This prevents misspelled operational fields from
// being silently discarded by a typed load/complete/write cycle.
func LoadWorkgraph(path string) (Workgraph, error) {
	var workgraph Workgraph
	data, err := os.ReadFile(path)
	if err != nil {
		return workgraph, err
	}
	var marker struct {
		OperationalBinding json.RawMessage `json:"operational_binding"`
	}
	if err := json.Unmarshal(data, &marker); err != nil {
		return workgraph, err
	}
	if len(marker.OperationalBinding) == 0 {
		if err := json.Unmarshal(data, &workgraph); err != nil {
			return workgraph, err
		}
		if workgraphRequiresOperationalBinding(workgraph) {
			return workgraph, fmt.Errorf("operational_binding is required when operational markers are present")
		}
		return workgraph, nil
	}
	if bytes.Equal(bytes.TrimSpace(marker.OperationalBinding), []byte("null")) {
		return workgraph, fmt.Errorf("operational_binding must not be null")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&workgraph); err != nil {
		return workgraph, fmt.Errorf("strict operational workgraph decode: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return workgraph, fmt.Errorf("strict operational workgraph decode: trailing JSON value")
	}
	if err := captureOperationalNodePresence(data, &workgraph); err != nil {
		return workgraph, err
	}
	return workgraph, nil
}

func LoadOperationalWorkgraphBindingDocument(path string) (OperationalWorkgraphBindingDocument, error) {
	var binding OperationalWorkgraphBindingDocument
	data, err := os.ReadFile(path)
	if err != nil {
		return binding, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&binding); err != nil {
		return binding, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return binding, fmt.Errorf("trailing JSON value")
	}
	return binding, nil
}

type operationalGraphDigestProjection struct {
	ContractVersion      string                            `json:"contract_version"`
	ID                   string                            `json:"id"`
	TargetInstance       string                            `json:"target_instance"`
	MissionID            string                            `json:"mission_id"`
	ObjectiveID          string                            `json:"objective_id"`
	ObjectiveDigest      string                            `json:"objective_digest"`
	CorrelationID        string                            `json:"correlation_id"`
	SoakID               string                            `json:"soak_id"`
	PlanID               string                            `json:"plan_id"`
	PolicyDigest         string                            `json:"policy_digest"`
	ActivationID         string                            `json:"activation_id"`
	EvidenceRootIdentity string                            `json:"evidence_root_identity"`
	MissionSourceHead    string                            `json:"mission_source_head"`
	AtlasSourceHead      string                            `json:"atlas_source_head"`
	MissionBinarySHA256  string                            `json:"mission_binary_sha256"`
	AtlasBinarySHA256    string                            `json:"atlas_binary_sha256"`
	OperationalBinding   OperationalWorkgraphBinding       `json:"operational_binding"`
	Nodes                []operationalNodeDigestProjection `json:"nodes"`
}

type operationalNodeDigestProjection struct {
	ID                 string                           `json:"id"`
	FactoryTask        FactoryTask                      `json:"factory_task"`
	Dependencies       []string                         `json:"dependencies"`
	Blockers           []string                         `json:"blockers"`
	StitchTask         bool                             `json:"stitch_task"`
	OperationalBinding *WorkgraphNodeOperationalBinding `json:"operational_binding"`
}

// ComputeOperationalWorkgraphBindingDigest excludes mutable node status and the
// digest field itself. Completion can therefore preserve the approved execution
// contract without making the digest self-referential.
func ComputeOperationalWorkgraphBindingDigest(workgraph Workgraph) (string, error) {
	if workgraph.OperationalBinding == nil {
		return "", fmt.Errorf("operational_binding is required")
	}
	binding := *workgraph.OperationalBinding
	binding.GraphBindingDigest = ""
	binding.PlannerPartitions = append([]WorkgraphPlannerPartition(nil), binding.PlannerPartitions...)
	sort.Slice(binding.PlannerPartitions, func(i, j int) bool {
		if binding.PlannerPartitions[i].NodeID != binding.PlannerPartitions[j].NodeID {
			return binding.PlannerPartitions[i].NodeID < binding.PlannerPartitions[j].NodeID
		}
		return binding.PlannerPartitions[i].PartitionID < binding.PlannerPartitions[j].PartitionID
	})
	projection := operationalGraphDigestProjection{
		ContractVersion:      workgraph.ContractVersion,
		ID:                   workgraph.ID,
		TargetInstance:       workgraph.TargetInstance,
		MissionID:            workgraph.MissionID,
		ObjectiveID:          workgraph.ObjectiveID,
		ObjectiveDigest:      workgraph.ObjectiveDigest,
		CorrelationID:        workgraph.CorrelationID,
		SoakID:               workgraph.SoakID,
		PlanID:               workgraph.PlanID,
		PolicyDigest:         workgraph.PolicyDigest,
		ActivationID:         workgraph.ActivationID,
		EvidenceRootIdentity: workgraph.EvidenceRootIdentity,
		MissionSourceHead:    workgraph.MissionSourceHead,
		AtlasSourceHead:      workgraph.AtlasSourceHead,
		MissionBinarySHA256:  workgraph.MissionBinarySHA256,
		AtlasBinarySHA256:    workgraph.AtlasBinarySHA256,
		OperationalBinding:   binding,
	}
	for _, node := range workgraph.Nodes {
		projection.Nodes = append(projection.Nodes, operationalNodeDigestProjection{
			ID:                 node.ID,
			FactoryTask:        node.FactoryTask,
			Dependencies:       append([]string(nil), node.Dependencies...),
			Blockers:           append([]string(nil), node.Blockers...),
			StitchTask:         node.StitchTask,
			OperationalBinding: node.OperationalBinding,
		})
	}
	body, err := json.Marshal(projection)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(body)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

func ValidateOperationalWorkgraphBinding(workgraph Workgraph, binding OperationalWorkgraphBindingDocument) OperationalWorkgraphBindingReadback {
	conflicts := []string{}
	if err := ValidateWorkgraph(workgraph); err != nil {
		conflicts = append(conflicts, "WORKGRAPH_INVALID")
	}
	if binding.ContractVersion != OperationalWorkgraphBindingDocumentContract {
		conflicts = append(conflicts, "BINDING_INVALID_CONTRACT_VERSION")
	}
	if workgraph.OperationalBinding == nil {
		conflicts = append(conflicts, "OPERATIONAL_BINDING_MISSING")
	} else {
		conflicts = append(conflicts, operationalWorkgraphConflicts(workgraph)...)
	}

	compareStringConflict(&conflicts, workgraph.ID, binding.WorkgraphID, "IDENTITY_MISMATCH_WORKGRAPH_ID")
	compareStringConflict(&conflicts, workgraph.TargetInstance, binding.TargetInstance, "IDENTITY_MISMATCH_TARGET_INSTANCE")
	compareStringConflict(&conflicts, workgraph.MissionID, binding.MissionID, "IDENTITY_MISMATCH_MISSION_ID")
	compareStringConflict(&conflicts, workgraph.ObjectiveID, binding.ObjectiveID, "IDENTITY_MISMATCH_OBJECTIVE_ID")
	compareStringConflict(&conflicts, workgraph.ObjectiveDigest, binding.ObjectiveDigest, "DIGEST_MISMATCH_OBJECTIVE")
	compareStringConflict(&conflicts, workgraph.CorrelationID, binding.CorrelationID, "IDENTITY_MISMATCH_CORRELATION_ID")
	compareStringConflict(&conflicts, workgraph.SoakID, binding.SoakID, "IDENTITY_MISMATCH_SOAK_ID")
	compareStringConflict(&conflicts, workgraph.PlanID, binding.PlanID, "IDENTITY_MISMATCH_PLAN_ID")
	compareStringConflict(&conflicts, workgraph.PolicyDigest, binding.PolicyDigest, "DIGEST_MISMATCH_POLICY")
	compareStringConflict(&conflicts, workgraph.ActivationID, binding.ActivationID, "IDENTITY_MISMATCH_ACTIVATION_ID")
	compareStringConflict(&conflicts, workgraph.EvidenceRootIdentity, binding.EvidenceRootIdentity, "IDENTITY_MISMATCH_EVIDENCE_ROOT")
	compareStringConflict(&conflicts, workgraph.MissionSourceHead, binding.MissionSourceHead, "SOURCE_HEAD_MISMATCH_MISSION")
	compareStringConflict(&conflicts, workgraph.AtlasSourceHead, binding.AtlasSourceHead, "SOURCE_HEAD_MISMATCH_ATLAS")
	compareStringConflict(&conflicts, workgraph.MissionBinarySHA256, binding.MissionBinarySHA256, "BINARY_DIGEST_MISMATCH_MISSION")
	compareStringConflict(&conflicts, workgraph.AtlasBinarySHA256, binding.AtlasBinarySHA256, "BINARY_DIGEST_MISMATCH_ATLAS")

	if workgraph.OperationalBinding != nil {
		expected := workgraph.OperationalBinding
		compareStringConflict(&conflicts, expected.ExecutionProfileDigest, binding.ExecutionProfileDigest, "DIGEST_MISMATCH_EXECUTION_PROFILE")
		compareStringConflict(&conflicts, expected.CommandCatalogDigest, binding.CommandCatalogDigest, "DIGEST_MISMATCH_COMMAND_CATALOG")
		compareStringConflict(&conflicts, expected.DurationHistoryDigest, binding.DurationHistoryDigest, "DIGEST_MISMATCH_DURATION_HISTORY")
		compareStringConflict(&conflicts, expected.PlannerInputDigest, binding.PlannerInputDigest, "DIGEST_MISMATCH_PLANNER_INPUT")
		compareStringConflict(&conflicts, expected.PlannerReadbackDigest, binding.PlannerReadbackDigest, "DIGEST_MISMATCH_PLANNER_READBACK")
		compareStringConflict(&conflicts, expected.GraphBindingDigest, binding.GraphBindingDigest, "GRAPH_BINDING_DIGEST_MISMATCH")
		if !equalRetryPolicy(expected.RetryPolicy, binding.RetryPolicy) {
			conflicts = append(conflicts, "RETRY_POLICY_MISMATCH")
		}
		conflicts = append(conflicts, retryPolicyConflicts(binding.RetryPolicy, len(binding.PlannerPartitions))...)
		conflicts = append(conflicts, comparePlannerPartitions(expected.PlannerPartitions, binding.PlannerPartitions)...)
		if !equalSafetyBoundaries(expected.SafetyBoundaries, binding.SafetyBoundaries) {
			conflicts = append(conflicts, "SAFETY_BOUNDARY_MISMATCH")
		}
		conflicts = append(conflicts, unsafeBoundaryConflicts(binding.SafetyBoundaries)...)
	}

	conflicts = sortedUniqueStrings(conflicts)
	allowed := len(conflicts) == 0
	next := "correct the operational Workgraph and binding conflicts; do not activate nodes"
	if allowed {
		next = "binding validated; an external controller may evaluate its separate activation gate"
	}
	return OperationalWorkgraphBindingReadback{
		ContractVersion:      OperationalWorkgraphBindingReadbackContract,
		WorkgraphID:          workgraph.ID,
		GraphBindingDigest:   binding.GraphBindingDigest,
		ActivationAllowed:    allowed,
		ConflictCodes:        conflicts,
		ChildProcessLaunches: 0,
		ExecutesWork:         false,
		SafeToExecute:        false,
		ExactNextAction:      next,
	}
}

func operationalWorkgraphConflicts(workgraph Workgraph) []string {
	if workgraph.OperationalBinding == nil {
		return []string{"OPERATIONAL_BINDING_MISSING"}
	}
	conflicts := []string{}
	requiredFields := []struct {
		value string
		code  string
	}{
		{workgraph.MissionID, "OPERATIONAL_FIELD_MISSING_MISSION_ID"},
		{workgraph.ObjectiveID, "OPERATIONAL_FIELD_MISSING_OBJECTIVE_ID"},
		{workgraph.ObjectiveDigest, "OPERATIONAL_FIELD_MISSING_OBJECTIVE_DIGEST"},
		{workgraph.CorrelationID, "OPERATIONAL_FIELD_MISSING_CORRELATION_ID"},
		{workgraph.SoakID, "OPERATIONAL_FIELD_MISSING_SOAK_ID"},
		{workgraph.PlanID, "OPERATIONAL_FIELD_MISSING_PLAN_ID"},
		{workgraph.PolicyDigest, "OPERATIONAL_FIELD_MISSING_POLICY_DIGEST"},
		{workgraph.ActivationID, "OPERATIONAL_FIELD_MISSING_ACTIVATION_ID"},
		{workgraph.EvidenceRootIdentity, "OPERATIONAL_FIELD_MISSING_EVIDENCE_ROOT_IDENTITY"},
		{workgraph.MissionSourceHead, "OPERATIONAL_FIELD_MISSING_MISSION_SOURCE_HEAD"},
		{workgraph.AtlasSourceHead, "OPERATIONAL_FIELD_MISSING_ATLAS_SOURCE_HEAD"},
		{workgraph.MissionBinarySHA256, "OPERATIONAL_FIELD_MISSING_MISSION_BINARY_SHA256"},
		{workgraph.AtlasBinarySHA256, "OPERATIONAL_FIELD_MISSING_ATLAS_BINARY_SHA256"},
		{workgraph.OperationalBinding.ExecutionProfileDigest, "OPERATIONAL_FIELD_MISSING_EXECUTION_PROFILE_DIGEST"},
		{workgraph.OperationalBinding.CommandCatalogDigest, "OPERATIONAL_FIELD_MISSING_COMMAND_CATALOG_DIGEST"},
		{workgraph.OperationalBinding.DurationHistoryDigest, "OPERATIONAL_FIELD_MISSING_DURATION_HISTORY_DIGEST"},
		{workgraph.OperationalBinding.PlannerInputDigest, "OPERATIONAL_FIELD_MISSING_PLANNER_INPUT_DIGEST"},
		{workgraph.OperationalBinding.PlannerReadbackDigest, "OPERATIONAL_FIELD_MISSING_PLANNER_READBACK_DIGEST"},
		{workgraph.OperationalBinding.GraphBindingDigest, "OPERATIONAL_FIELD_MISSING_GRAPH_BINDING_DIGEST"},
	}
	for _, field := range requiredFields {
		if strings.TrimSpace(field.value) == "" {
			conflicts = append(conflicts, field.code)
		}
	}
	digests := []struct {
		value string
		code  string
	}{
		{workgraph.ObjectiveDigest, "OPERATIONAL_FIELD_INVALID_OBJECTIVE_DIGEST"},
		{workgraph.PolicyDigest, "OPERATIONAL_FIELD_INVALID_POLICY_DIGEST"},
		{workgraph.MissionBinarySHA256, "OPERATIONAL_FIELD_INVALID_MISSION_BINARY_SHA256"},
		{workgraph.AtlasBinarySHA256, "OPERATIONAL_FIELD_INVALID_ATLAS_BINARY_SHA256"},
		{workgraph.OperationalBinding.ExecutionProfileDigest, "OPERATIONAL_FIELD_INVALID_EXECUTION_PROFILE_DIGEST"},
		{workgraph.OperationalBinding.CommandCatalogDigest, "OPERATIONAL_FIELD_INVALID_COMMAND_CATALOG_DIGEST"},
		{workgraph.OperationalBinding.DurationHistoryDigest, "OPERATIONAL_FIELD_INVALID_DURATION_HISTORY_DIGEST"},
		{workgraph.OperationalBinding.PlannerInputDigest, "OPERATIONAL_FIELD_INVALID_PLANNER_INPUT_DIGEST"},
		{workgraph.OperationalBinding.PlannerReadbackDigest, "OPERATIONAL_FIELD_INVALID_PLANNER_READBACK_DIGEST"},
		{workgraph.OperationalBinding.GraphBindingDigest, "OPERATIONAL_FIELD_INVALID_GRAPH_BINDING_DIGEST"},
	}
	for _, field := range digests {
		if strings.TrimSpace(field.value) != "" && !digestPattern.MatchString(field.value) {
			conflicts = append(conflicts, field.code)
		}
	}
	if !sourceHeadPattern.MatchString(workgraph.MissionSourceHead) {
		conflicts = append(conflicts, "OPERATIONAL_FIELD_INVALID_MISSION_SOURCE_HEAD")
	}
	if !sourceHeadPattern.MatchString(workgraph.AtlasSourceHead) {
		conflicts = append(conflicts, "OPERATIONAL_FIELD_INVALID_ATLAS_SOURCE_HEAD")
	}
	conflicts = append(conflicts, retryPolicyConflicts(workgraph.OperationalBinding.RetryPolicy, len(workgraph.OperationalBinding.PlannerPartitions))...)
	conflicts = append(conflicts, plannerPartitionConflicts(workgraph)...)
	conflicts = append(conflicts, unsafeBoundaryConflicts(workgraph.OperationalBinding.SafetyBoundaries)...)
	computed, err := ComputeOperationalWorkgraphBindingDigest(workgraph)
	if err != nil || computed != workgraph.OperationalBinding.GraphBindingDigest {
		conflicts = append(conflicts, "GRAPH_BINDING_DIGEST_MISMATCH")
	}
	return sortedUniqueStrings(conflicts)
}

func retryPolicyConflicts(policy WorkgraphRetryPolicy, partitionCount int) []string {
	conflicts := []string{}
	if policy.presentFields != nil {
		for _, field := range []string{"maximum_attempts", "maximum_total_retries"} {
			if !policy.presentFields[field] {
				conflicts = append(conflicts, "RETRY_POLICY_MISSING_"+strings.ToUpper(field))
			}
		}
	}
	if policy.MaximumAttempts < 1 || policy.MaximumAttempts > 4 {
		conflicts = append(conflicts, "RETRY_POLICY_INVALID")
	}
	if policy.MaximumTotalRetries < 0 {
		conflicts = append(conflicts, "RETRY_POLICY_INVALID")
	}
	if policy.MaximumAttempts == 1 && policy.MaximumTotalRetries > 0 {
		conflicts = append(conflicts, "RETRY_POLICY_INVALID")
	}
	attemptRetries := policy.MaximumAttempts - 1
	maxInt := int(^uint(0) >> 1)
	if partitionCount < 0 || attemptRetries < 0 ||
		(attemptRetries > 0 && partitionCount > maxInt/attemptRetries) {
		conflicts = append(conflicts, "RETRY_POLICY_INVALID")
		return conflicts
	}
	retryCapacity := partitionCount * attemptRetries
	if policy.MaximumTotalRetries > retryCapacity {
		conflicts = append(conflicts, "RETRY_POLICY_CAPACITY_EXCEEDED")
	}
	return conflicts
}

func plannerPartitionConflicts(workgraph Workgraph) []string {
	binding := workgraph.OperationalBinding
	if binding == nil {
		return []string{"OPERATIONAL_BINDING_MISSING"}
	}
	conflicts := []string{}
	if len(binding.PlannerPartitions) == 0 {
		conflicts = append(conflicts, "PARTITION_MISSING")
	}
	nodes := map[string]WorkgraphNode{}
	for _, node := range workgraph.Nodes {
		nodes[node.ID] = node
	}
	partitionIDs := map[string]bool{}
	partitionNodes := map[string]bool{}
	partitions := map[string]WorkgraphPlannerPartition{}
	for _, partition := range binding.PlannerPartitions {
		conflicts = append(conflicts, partitionPresenceConflicts(partition)...)
		if partition.PartitionID == "" || partition.NodeID == "" || partition.TestID == "" || partition.Classification == "" {
			conflicts = append(conflicts, "PARTITION_INVALID")
		}
		if partitionIDs[partition.PartitionID] {
			conflicts = append(conflicts, "PARTITION_DUPLICATE_ID")
		}
		if partitionNodes[partition.NodeID] {
			conflicts = append(conflicts, "PARTITION_DUPLICATE_NODE")
		}
		partitionIDs[partition.PartitionID] = true
		partitionNodes[partition.NodeID] = true
		partitions[partition.PartitionID] = partition
		if _, ok := nodes[partition.NodeID]; !ok {
			conflicts = append(conflicts, "PARTITION_UNREFERENCED")
		}
		if partition.RequestedRepeatCount < 1 || partition.EffectiveRepeatCount < 1 ||
			partition.EstimatedDurationMS < 1 || partition.RetryAllowanceMS < 0 ||
			partition.PerAttemptTimeoutMS < 1 || partition.TotalNodeTimeoutMS < 1 ||
			partition.NodeBudgetMS < 1 {
			conflicts = append(conflicts, "PARTITION_INVALID")
		}
		if partition.EffectiveRepeatCount > partition.RequestedRepeatCount ||
			partition.PerAttemptTimeoutMS < partition.EstimatedDurationMS ||
			partition.TotalNodeTimeoutMS < partition.PerAttemptTimeoutMS ||
			partition.NodeBudgetMS < partition.TotalNodeTimeoutMS ||
			partition.RetryAllowanceMS > partition.TotalNodeTimeoutMS ||
			(binding.RetryPolicy.MaximumAttempts > 1 && partition.RetryAllowanceMS < partition.EstimatedDurationMS) {
			conflicts = append(conflicts, "PARTITION_POLICY_CONTRADICTION")
		}
		maxInt64 := int64(^uint64(0) >> 1)
		if binding.RetryPolicy.MaximumAttempts > 0 {
			attempts := int64(binding.RetryPolicy.MaximumAttempts)
			if partition.EstimatedDurationMS > maxInt64/attempts {
				conflicts = append(conflicts, "PARTITION_RETRY_ALLOWANCE_OVERFLOW")
			} else if partition.RetryAllowanceMS != partition.EstimatedDurationMS*attempts {
				conflicts = append(conflicts, "PARTITION_RETRY_ALLOWANCE_MISMATCH")
			}
		}
		if !oneOf(partition.Classification, "regular", "scale") {
			conflicts = append(conflicts, "PARTITION_CLASSIFICATION_INVALID")
		}
		if partition.Classification == "scale" && strings.TrimSpace(partition.ScaleDimension) == "" {
			conflicts = append(conflicts, "PARTITION_SCALE_DIMENSION_MISSING")
		}
		if partition.Classification == "scale" && partition.EffectiveRepeatCount != 1 {
			conflicts = append(conflicts, "PARTITION_SCALE_REPEAT_INVALID")
		}
		if partition.Classification == "regular" && strings.TrimSpace(partition.ScaleDimension) != "" {
			conflicts = append(conflicts, "PARTITION_REGULAR_SCALE_DIMENSION_PRESENT")
		}
	}
	for _, node := range workgraph.Nodes {
		conflicts = append(conflicts, operationalNodePresenceConflicts(node)...)
		if node.OperationalBinding == nil {
			conflicts = append(conflicts, "PARTITION_MISSING_FOR_NODE")
			continue
		}
		partition, ok := partitions[node.OperationalBinding.PlannerPartitionID]
		if !ok {
			conflicts = append(conflicts, "NODE_PARTITION_REFERENCE_UNKNOWN")
			continue
		}
		if partition.NodeID != node.ID {
			conflicts = append(conflicts, "NODE_PARTITION_MISMATCH_NODE_ID")
		}
		if partition.TestID != node.OperationalBinding.TestID {
			conflicts = append(conflicts, "NODE_PARTITION_MISMATCH_TEST_ID")
		}
		if partition.Classification != node.OperationalBinding.Classification {
			conflicts = append(conflicts, "NODE_PARTITION_MISMATCH_CLASSIFICATION")
		}
		if !equalSafetyBoundaries(binding.SafetyBoundaries, node.OperationalBinding.SafetyBoundaries) {
			conflicts = append(conflicts, "NODE_SAFETY_BOUNDARY_MISMATCH")
		}
		conflicts = append(conflicts, unsafeBoundaryConflicts(node.OperationalBinding.SafetyBoundaries)...)
	}
	return conflicts
}

func captureOperationalNodePresence(data []byte, workgraph *Workgraph) error {
	var raw struct {
		Nodes []map[string]json.RawMessage `json:"nodes"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw.Nodes) != len(workgraph.Nodes) {
		return fmt.Errorf("strict operational workgraph decode: node count changed during presence capture")
	}
	for i, fields := range raw.Nodes {
		workgraph.Nodes[i].presentFields = make(map[string]bool, len(fields))
		workgraph.Nodes[i].nullFields = map[string]bool{}
		for field, value := range fields {
			workgraph.Nodes[i].presentFields[field] = true
			if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
				workgraph.Nodes[i].nullFields[field] = true
				return fmt.Errorf("strict operational workgraph decode: nodes[%d].%s must not be null", i, field)
			}
		}
	}
	return nil
}

func operationalNodePresenceConflicts(node WorkgraphNode) []string {
	if node.presentFields == nil {
		return nil
	}
	required := []string{
		"id",
		"status",
		"factory_task",
		"dependencies",
		"blockers",
		"stitch_task",
		"operational_binding",
	}
	conflicts := []string{}
	for _, field := range required {
		if !node.presentFields[field] {
			conflicts = append(conflicts, "NODE_FIELD_MISSING_"+strings.ToUpper(field))
		}
	}
	for _, field := range []string{"dependencies", "blockers"} {
		if node.nullFields[field] {
			conflicts = append(conflicts, "NODE_FIELD_NULL_"+strings.ToUpper(field))
		}
	}
	return conflicts
}

func comparePlannerPartitions(expected, actual []WorkgraphPlannerPartition) []string {
	conflicts := []string{}
	if len(expected) != len(actual) {
		conflicts = append(conflicts, "PARTITION_COUNT_MISMATCH")
	}
	expectedByID := map[string]WorkgraphPlannerPartition{}
	for _, partition := range expected {
		expectedByID[partition.PartitionID] = partition
	}
	actualIDs := map[string]bool{}
	for _, got := range actual {
		conflicts = append(conflicts, partitionPresenceConflicts(got)...)
		if actualIDs[got.PartitionID] {
			conflicts = append(conflicts, "PARTITION_DUPLICATE_ID")
		}
		actualIDs[got.PartitionID] = true
		want, ok := expectedByID[got.PartitionID]
		if !ok {
			conflicts = append(conflicts, "PARTITION_UNREFERENCED")
			continue
		}
		comparePartitionAttribute(&conflicts, want.NodeID == got.NodeID, "NODE_PARTITION_MISMATCH_NODE_ID")
		comparePartitionAttribute(&conflicts, want.TestID == got.TestID, "NODE_PARTITION_MISMATCH_TEST_ID")
		comparePartitionAttribute(&conflicts, want.Classification == got.Classification, "NODE_PARTITION_MISMATCH_CLASSIFICATION")
		comparePartitionAttribute(&conflicts, want.RequestedRepeatCount == got.RequestedRepeatCount, "NODE_PARTITION_MISMATCH_REQUESTED_REPEAT_COUNT")
		comparePartitionAttribute(&conflicts, want.EffectiveRepeatCount == got.EffectiveRepeatCount, "NODE_PARTITION_MISMATCH_EFFECTIVE_REPEAT_COUNT")
		comparePartitionAttribute(&conflicts, want.ScaleDimension == got.ScaleDimension, "NODE_PARTITION_MISMATCH_SCALE_DIMENSION")
		comparePartitionAttribute(&conflicts, want.EstimatedDurationMS == got.EstimatedDurationMS, "NODE_PARTITION_MISMATCH_ESTIMATED_DURATION_MS")
		comparePartitionAttribute(&conflicts, want.RetryAllowanceMS == got.RetryAllowanceMS, "NODE_PARTITION_MISMATCH_RETRY_ALLOWANCE_MS")
		comparePartitionAttribute(&conflicts, want.PerAttemptTimeoutMS == got.PerAttemptTimeoutMS, "NODE_PARTITION_MISMATCH_PER_ATTEMPT_TIMEOUT_MS")
		comparePartitionAttribute(&conflicts, want.TotalNodeTimeoutMS == got.TotalNodeTimeoutMS, "NODE_PARTITION_MISMATCH_TOTAL_NODE_TIMEOUT_MS")
		comparePartitionAttribute(&conflicts, want.NodeBudgetMS == got.NodeBudgetMS, "NODE_PARTITION_MISMATCH_NODE_BUDGET_MS")
	}
	for partitionID := range expectedByID {
		if !actualIDs[partitionID] {
			conflicts = append(conflicts, "PARTITION_MISSING")
		}
	}
	return conflicts
}

func unsafeBoundaryConflicts(safety WorkgraphSafetyBoundaries) []string {
	conflicts := []string{}
	if safety.presentFields != nil {
		required := []string{
			"mutates_repositories",
			"releases_or_publishes",
			"calls_providers",
			"expands_authority",
			"allows_unplanned_commands",
			"allows_unbounded_retries",
			"activates_child_processes",
			"schedules_work",
			"executes_work",
			"approves_work",
			"activation_requires_validated_binding",
			"preserves_phase_clock_on_retry",
		}
		for _, field := range required {
			if !safety.presentFields[field] {
				conflicts = append(conflicts, "SAFETY_BOUNDARY_MISSING_"+strings.ToUpper(field))
			}
		}
	}
	if safety.MutatesRepositories {
		conflicts = append(conflicts, "SAFETY_BOUNDARY_UNSAFE_MUTATES_REPOSITORIES")
	}
	if safety.ReleasesOrPublishes {
		conflicts = append(conflicts, "SAFETY_BOUNDARY_UNSAFE_RELEASES_OR_PUBLISHES")
	}
	if safety.CallsProviders {
		conflicts = append(conflicts, "SAFETY_BOUNDARY_UNSAFE_CALLS_PROVIDERS")
	}
	if safety.ExpandsAuthority {
		conflicts = append(conflicts, "SAFETY_BOUNDARY_UNSAFE_EXPANDS_AUTHORITY")
	}
	if safety.AllowsUnplannedCommands {
		conflicts = append(conflicts, "SAFETY_BOUNDARY_UNSAFE_ALLOWS_UNPLANNED_COMMANDS")
	}
	if safety.AllowsUnboundedRetries {
		conflicts = append(conflicts, "SAFETY_BOUNDARY_UNSAFE_ALLOWS_UNBOUNDED_RETRIES")
	}
	if safety.ActivatesChildProcesses {
		conflicts = append(conflicts, "SAFETY_BOUNDARY_UNSAFE_ACTIVATES_CHILD_PROCESSES")
	}
	if safety.SchedulesWork {
		conflicts = append(conflicts, "SAFETY_BOUNDARY_UNSAFE_SCHEDULES_WORK")
	}
	if safety.ExecutesWork {
		conflicts = append(conflicts, "SAFETY_BOUNDARY_UNSAFE_EXECUTES_WORK")
	}
	if safety.ApprovesWork {
		conflicts = append(conflicts, "SAFETY_BOUNDARY_UNSAFE_APPROVES_WORK")
	}
	if !safety.ActivationRequiresValidatedBinding {
		conflicts = append(conflicts, "SAFETY_BOUNDARY_UNSAFE_ACTIVATION_REQUIRES_VALIDATED_BINDING")
	}
	if !safety.PreservesPhaseClockOnRetry {
		conflicts = append(conflicts, "SAFETY_BOUNDARY_UNSAFE_PRESERVES_PHASE_CLOCK_ON_RETRY")
	}
	return conflicts
}

func (safety *WorkgraphSafetyBoundaries) UnmarshalJSON(data []byte) error {
	type safetyAlias WorkgraphSafetyBoundaries
	var decoded safetyAlias
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&decoded); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("trailing safety boundary JSON value")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	for field, value := range fields {
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return fmt.Errorf("safety boundary %s must not be null", field)
		}
	}
	*safety = WorkgraphSafetyBoundaries(decoded)
	safety.presentFields = make(map[string]bool, len(fields))
	for field := range fields {
		safety.presentFields[field] = true
	}
	return nil
}

func equalSafetyBoundaries(left, right WorkgraphSafetyBoundaries) bool {
	return left.MutatesRepositories == right.MutatesRepositories &&
		left.ReleasesOrPublishes == right.ReleasesOrPublishes &&
		left.CallsProviders == right.CallsProviders &&
		left.ExpandsAuthority == right.ExpandsAuthority &&
		left.AllowsUnplannedCommands == right.AllowsUnplannedCommands &&
		left.AllowsUnboundedRetries == right.AllowsUnboundedRetries &&
		left.ActivatesChildProcesses == right.ActivatesChildProcesses &&
		left.SchedulesWork == right.SchedulesWork &&
		left.ExecutesWork == right.ExecutesWork &&
		left.ApprovesWork == right.ApprovesWork &&
		left.ActivationRequiresValidatedBinding == right.ActivationRequiresValidatedBinding &&
		left.PreservesPhaseClockOnRetry == right.PreservesPhaseClockOnRetry
}

func (policy *WorkgraphRetryPolicy) UnmarshalJSON(data []byte) error {
	type retryPolicyAlias WorkgraphRetryPolicy
	var decoded retryPolicyAlias
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&decoded); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("trailing retry policy JSON value")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	for field, value := range fields {
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return fmt.Errorf("retry policy %s must not be null", field)
		}
	}
	*policy = WorkgraphRetryPolicy(decoded)
	policy.presentFields = make(map[string]bool, len(fields))
	for field := range fields {
		policy.presentFields[field] = true
	}
	return nil
}

func (partition *WorkgraphPlannerPartition) UnmarshalJSON(data []byte) error {
	type partitionAlias WorkgraphPlannerPartition
	var decoded partitionAlias
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&decoded); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("trailing planner partition JSON value")
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	for field, value := range fields {
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return fmt.Errorf("planner partition %s must not be null", field)
		}
	}
	*partition = WorkgraphPlannerPartition(decoded)
	partition.presentFields = make(map[string]bool, len(fields))
	for field := range fields {
		partition.presentFields[field] = true
	}
	return nil
}

func partitionPresenceConflicts(partition WorkgraphPlannerPartition) []string {
	if partition.presentFields == nil {
		return nil
	}
	required := []string{
		"partition_id",
		"node_id",
		"test_id",
		"classification",
		"requested_repeat_count",
		"effective_repeat_count",
		"estimated_duration_ms",
		"retry_allowance_ms",
		"per_attempt_timeout_ms",
		"total_node_timeout_ms",
		"node_budget_ms",
	}
	conflicts := []string{}
	for _, field := range required {
		if !partition.presentFields[field] {
			conflicts = append(conflicts, "PARTITION_FIELD_MISSING_"+strings.ToUpper(field))
		}
	}
	return conflicts
}

func equalRetryPolicy(left, right WorkgraphRetryPolicy) bool {
	return left.MaximumAttempts == right.MaximumAttempts &&
		left.MaximumTotalRetries == right.MaximumTotalRetries
}

func compareStringConflict(conflicts *[]string, expected, actual, code string) {
	if expected != actual {
		*conflicts = append(*conflicts, code)
	}
}

func comparePartitionAttribute(conflicts *[]string, matches bool, code string) {
	if !matches {
		*conflicts = append(*conflicts, code)
	}
}

func sortedUniqueStrings(values []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
