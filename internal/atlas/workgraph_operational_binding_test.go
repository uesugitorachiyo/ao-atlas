package atlas

import (
	"bytes"
	"encoding/json"
	"reflect"
	"sort"
	"testing"
)

func TestOperationalWorkgraphBindingAcceptsExactTypedContract(t *testing.T) {
	workgraph, binding := validOperationalWorkgraphAndBinding(t)

	readback := ValidateOperationalWorkgraphBinding(workgraph, binding)
	if !readback.ActivationAllowed || len(readback.ConflictCodes) != 0 {
		t.Fatalf("exact operational binding denied: %#v", readback)
	}
	if readback.ChildProcessLaunches != 0 || readback.ExecutesWork || readback.SafeToExecute {
		t.Fatalf("validator crossed no-execution boundary: %#v", readback)
	}
}

func TestOperationalWorkgraphBindingRejectsEveryCriticalScalarMismatch(t *testing.T) {
	mutations := []struct {
		name string
		code string
		edit func(*OperationalWorkgraphBindingDocument)
	}{
		{"workgraph id", "IDENTITY_MISMATCH_WORKGRAPH_ID", func(v *OperationalWorkgraphBindingDocument) { v.WorkgraphID += "-other" }},
		{"target instance", "IDENTITY_MISMATCH_TARGET_INSTANCE", func(v *OperationalWorkgraphBindingDocument) { v.TargetInstance += "-other" }},
		{"mission id", "IDENTITY_MISMATCH_MISSION_ID", func(v *OperationalWorkgraphBindingDocument) { v.MissionID += "-other" }},
		{"objective id", "IDENTITY_MISMATCH_OBJECTIVE_ID", func(v *OperationalWorkgraphBindingDocument) { v.ObjectiveID += "-other" }},
		{"objective digest", "DIGEST_MISMATCH_OBJECTIVE", func(v *OperationalWorkgraphBindingDocument) { v.ObjectiveDigest = testDigest('9') }},
		{"correlation id", "IDENTITY_MISMATCH_CORRELATION_ID", func(v *OperationalWorkgraphBindingDocument) { v.CorrelationID += "-other" }},
		{"soak id", "IDENTITY_MISMATCH_SOAK_ID", func(v *OperationalWorkgraphBindingDocument) { v.SoakID += "-other" }},
		{"plan id", "IDENTITY_MISMATCH_PLAN_ID", func(v *OperationalWorkgraphBindingDocument) { v.PlanID += "-other" }},
		{"policy digest", "DIGEST_MISMATCH_POLICY", func(v *OperationalWorkgraphBindingDocument) { v.PolicyDigest = testDigest('8') }},
		{"activation id", "IDENTITY_MISMATCH_ACTIVATION_ID", func(v *OperationalWorkgraphBindingDocument) { v.ActivationID += "-other" }},
		{"evidence root", "IDENTITY_MISMATCH_EVIDENCE_ROOT", func(v *OperationalWorkgraphBindingDocument) { v.EvidenceRootIdentity += "-other" }},
		{"mission source", "SOURCE_HEAD_MISMATCH_MISSION", func(v *OperationalWorkgraphBindingDocument) { v.MissionSourceHead = testHead('7') }},
		{"atlas source", "SOURCE_HEAD_MISMATCH_ATLAS", func(v *OperationalWorkgraphBindingDocument) { v.AtlasSourceHead = testHead('6') }},
		{"mission binary", "BINARY_DIGEST_MISMATCH_MISSION", func(v *OperationalWorkgraphBindingDocument) { v.MissionBinarySHA256 = testDigest('5') }},
		{"atlas binary", "BINARY_DIGEST_MISMATCH_ATLAS", func(v *OperationalWorkgraphBindingDocument) { v.AtlasBinarySHA256 = testDigest('4') }},
		{"execution profile", "DIGEST_MISMATCH_EXECUTION_PROFILE", func(v *OperationalWorkgraphBindingDocument) { v.ExecutionProfileDigest = testDigest('3') }},
		{"command catalog", "DIGEST_MISMATCH_COMMAND_CATALOG", func(v *OperationalWorkgraphBindingDocument) { v.CommandCatalogDigest = testDigest('2') }},
		{"duration history", "DIGEST_MISMATCH_DURATION_HISTORY", func(v *OperationalWorkgraphBindingDocument) { v.DurationHistoryDigest = testDigest('1') }},
		{"planner input", "DIGEST_MISMATCH_PLANNER_INPUT", func(v *OperationalWorkgraphBindingDocument) { v.PlannerInputDigest = testDigest('a') }},
		{"planner readback", "DIGEST_MISMATCH_PLANNER_READBACK", func(v *OperationalWorkgraphBindingDocument) { v.PlannerReadbackDigest = testDigest('b') }},
	}
	for _, test := range mutations {
		t.Run(test.name, func(t *testing.T) {
			workgraph, binding := validOperationalWorkgraphAndBinding(t)
			test.edit(&binding)
			readback := ValidateOperationalWorkgraphBinding(workgraph, binding)
			if readback.ActivationAllowed || !containsString(readback.ConflictCodes, test.code) {
				t.Fatalf("mismatch was not denied with %s: %#v", test.code, readback)
			}
			if !sort.StringsAreSorted(readback.ConflictCodes) {
				t.Fatalf("conflicts are not deterministic: %#v", readback.ConflictCodes)
			}
		})
	}
}

func TestOperationalWorkgraphBindingRejectsRetryPartitionAndSafetyDrift(t *testing.T) {
	tests := []struct {
		name string
		code string
		edit func(*OperationalWorkgraphBindingDocument)
	}{
		{"attempt cap", "RETRY_POLICY_MISMATCH", func(v *OperationalWorkgraphBindingDocument) { v.RetryPolicy.MaximumAttempts++ }},
		{"total retry cap", "RETRY_POLICY_MISMATCH", func(v *OperationalWorkgraphBindingDocument) { v.RetryPolicy.MaximumTotalRetries++ }},
		{"partition test", "NODE_PARTITION_MISMATCH_TEST_ID", func(v *OperationalWorkgraphBindingDocument) { v.PlannerPartitions[0].TestID += "-other" }},
		{"partition classification", "NODE_PARTITION_MISMATCH_CLASSIFICATION", func(v *OperationalWorkgraphBindingDocument) { v.PlannerPartitions[0].Classification = "scale" }},
		{"partition requested repeats", "NODE_PARTITION_MISMATCH_REQUESTED_REPEAT_COUNT", func(v *OperationalWorkgraphBindingDocument) { v.PlannerPartitions[0].RequestedRepeatCount++ }},
		{"partition effective repeats", "NODE_PARTITION_MISMATCH_EFFECTIVE_REPEAT_COUNT", func(v *OperationalWorkgraphBindingDocument) { v.PlannerPartitions[0].EffectiveRepeatCount++ }},
		{"partition scale dimension", "NODE_PARTITION_MISMATCH_SCALE_DIMENSION", func(v *OperationalWorkgraphBindingDocument) { v.PlannerPartitions[0].ScaleDimension = "nodes=60" }},
		{"partition estimate", "NODE_PARTITION_MISMATCH_ESTIMATED_DURATION_MS", func(v *OperationalWorkgraphBindingDocument) { v.PlannerPartitions[0].EstimatedDurationMS++ }},
		{"partition allowance", "NODE_PARTITION_MISMATCH_RETRY_ALLOWANCE_MS", func(v *OperationalWorkgraphBindingDocument) { v.PlannerPartitions[0].RetryAllowanceMS++ }},
		{"partition attempt timeout", "NODE_PARTITION_MISMATCH_PER_ATTEMPT_TIMEOUT_MS", func(v *OperationalWorkgraphBindingDocument) { v.PlannerPartitions[0].PerAttemptTimeoutMS++ }},
		{"partition total timeout", "NODE_PARTITION_MISMATCH_TOTAL_NODE_TIMEOUT_MS", func(v *OperationalWorkgraphBindingDocument) { v.PlannerPartitions[0].TotalNodeTimeoutMS++ }},
		{"partition budget", "NODE_PARTITION_MISMATCH_NODE_BUDGET_MS", func(v *OperationalWorkgraphBindingDocument) { v.PlannerPartitions[0].NodeBudgetMS++ }},
		{"unsafe repository mutation", "SAFETY_BOUNDARY_UNSAFE_MUTATES_REPOSITORIES", func(v *OperationalWorkgraphBindingDocument) { v.SafetyBoundaries.MutatesRepositories = true }},
		{"unsafe provider call", "SAFETY_BOUNDARY_UNSAFE_CALLS_PROVIDERS", func(v *OperationalWorkgraphBindingDocument) { v.SafetyBoundaries.CallsProviders = true }},
		{"unsafe authority", "SAFETY_BOUNDARY_UNSAFE_EXPANDS_AUTHORITY", func(v *OperationalWorkgraphBindingDocument) { v.SafetyBoundaries.ExpandsAuthority = true }},
		{"unvalidated activation", "SAFETY_BOUNDARY_UNSAFE_ACTIVATION_REQUIRES_VALIDATED_BINDING", func(v *OperationalWorkgraphBindingDocument) { v.SafetyBoundaries.ActivationRequiresValidatedBinding = false }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			workgraph, binding := validOperationalWorkgraphAndBinding(t)
			test.edit(&binding)
			readback := ValidateOperationalWorkgraphBinding(workgraph, binding)
			if readback.ActivationAllowed || !containsString(readback.ConflictCodes, test.code) {
				t.Fatalf("drift was not denied with %s: %#v", test.code, readback)
			}
		})
	}
}

func TestOperationalWorkgraphCompletePreservesTypedBinding(t *testing.T) {
	workgraph, _ := validOperationalWorkgraphAndBinding(t)
	link := RunLink{
		ContractVersion: RunLinkContract,
		TaskID:          workgraph.Nodes[0].FactoryTask.ID,
		Status:          "completed",
		Evidence:        map[string]string{"result": "evidence/result.json"},
		Digest:          testDigest('d'),
	}
	completed, _, err := CompleteWorkgraph(workgraph, link)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(completed.OperationalBinding, workgraph.OperationalBinding) {
		t.Fatalf("complete dropped operational binding:\nwant=%#v\ngot=%#v", workgraph.OperationalBinding, completed.OperationalBinding)
	}
	if !reflect.DeepEqual(completed.Nodes[0].OperationalBinding, workgraph.Nodes[0].OperationalBinding) {
		t.Fatalf("complete dropped node operational binding")
	}
}

func TestOperationalWorkgraphValidateBindingCLIEmitsJSONOnDenial(t *testing.T) {
	workgraph, binding := validOperationalWorkgraphAndBinding(t)
	binding.PlanID += "-tampered"
	dir := t.TempDir()
	workgraphPath := dir + "/workgraph.json"
	bindingPath := dir + "/binding.json"
	if err := WriteJSON(workgraphPath, workgraph); err != nil {
		t.Fatal(err)
	}
	if err := WriteJSON(bindingPath, binding); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"workgraph", "validate-binding", "--workgraph", workgraphPath, "--binding", bindingPath, "--json"}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("denied binding exited zero: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	var readback OperationalWorkgraphBindingReadback
	if err := json.Unmarshal(stdout.Bytes(), &readback); err != nil {
		t.Fatalf("denial stdout is not JSON: %v\nstdout=%s\nstderr=%s", err, stdout.String(), stderr.String())
	}
	if readback.ActivationAllowed || !containsString(readback.ConflictCodes, "IDENTITY_MISMATCH_PLAN_ID") || readback.ChildProcessLaunches != 0 {
		t.Fatalf("bad denial readback: %#v", readback)
	}
}

func TestLegacyWorkgraphRemainsValidWithoutOperationalBinding(t *testing.T) {
	if err := ValidateWorkgraph(fixtureWorkgraph()); err != nil {
		t.Fatalf("legacy Workgraph compatibility regressed: %v", err)
	}
}

func validOperationalWorkgraphAndBinding(t *testing.T) (Workgraph, OperationalWorkgraphBindingDocument) {
	t.Helper()
	workgraph := fixtureWorkgraph()
	workgraph.Nodes = workgraph.Nodes[:1]
	workgraph.Nodes[0].Status = "ready"
	workgraph.Nodes[0].Dependencies = nil
	workgraph.MissionID = "mission-binding"
	workgraph.ObjectiveID = "objective-binding"
	workgraph.ObjectiveDigest = testDigest('0')
	workgraph.CorrelationID = "correlation-binding"
	workgraph.SoakID = "soak-binding"
	workgraph.PlanID = "plan-binding"
	workgraph.PolicyDigest = testDigest('c')
	workgraph.ActivationID = "activation-binding"
	workgraph.EvidenceRootIdentity = "evidence-binding"
	workgraph.MissionSourceHead = testHead('1')
	workgraph.AtlasSourceHead = testHead('2')
	workgraph.MissionBinarySHA256 = testDigest('3')
	workgraph.AtlasBinarySHA256 = testDigest('4')

	safety := WorkgraphSafetyBoundaries{
		MutatesRepositories:               false,
		ReleasesOrPublishes:               false,
		CallsProviders:                    false,
		ExpandsAuthority:                  false,
		AllowsUnplannedCommands:           false,
		AllowsUnboundedRetries:            false,
		ActivationRequiresValidatedBinding: true,
		PreservesPhaseClockOnRetry:        true,
	}
	partition := WorkgraphPlannerPartition{
		PartitionID:          "partition-01",
		NodeID:               workgraph.Nodes[0].ID,
		TestID:               "atlas-test-01",
		Classification:       "regular",
		RequestedRepeatCount: 60,
		EffectiveRepeatCount: 60,
		EstimatedDurationMS:  1000,
		RetryAllowanceMS:     1000,
		PerAttemptTimeoutMS:  2000,
		TotalNodeTimeoutMS:   4000,
		NodeBudgetMS:         4000,
	}
	workgraph.Nodes[0].OperationalBinding = &WorkgraphNodeOperationalBinding{
		PlannerPartitionID: partition.PartitionID,
		TestID:            partition.TestID,
		Classification:    partition.Classification,
		SafetyBoundaries:  safety,
	}
	workgraph.OperationalBinding = &OperationalWorkgraphBinding{
		ExecutionProfileDigest: testDigest('5'),
		CommandCatalogDigest:   testDigest('6'),
		DurationHistoryDigest:  testDigest('7'),
		PlannerInputDigest:     testDigest('8'),
		PlannerReadbackDigest:  testDigest('9'),
		RetryPolicy:            WorkgraphRetryPolicy{MaximumAttempts: 2, MaximumTotalRetries: 1},
		PlannerPartitions:      []WorkgraphPlannerPartition{partition},
		SafetyBoundaries:       safety,
	}
	graphDigest, err := ComputeOperationalWorkgraphBindingDigest(workgraph)
	if err != nil {
		t.Fatal(err)
	}
	workgraph.OperationalBinding.GraphBindingDigest = graphDigest
	binding := OperationalWorkgraphBindingDocument{
		ContractVersion:       OperationalWorkgraphBindingDocumentContract,
		WorkgraphID:           workgraph.ID,
		TargetInstance:        workgraph.TargetInstance,
		MissionID:             workgraph.MissionID,
		ObjectiveID:           workgraph.ObjectiveID,
		ObjectiveDigest:       workgraph.ObjectiveDigest,
		CorrelationID:         workgraph.CorrelationID,
		SoakID:                workgraph.SoakID,
		PlanID:                workgraph.PlanID,
		PolicyDigest:          workgraph.PolicyDigest,
		ActivationID:          workgraph.ActivationID,
		EvidenceRootIdentity:  workgraph.EvidenceRootIdentity,
		MissionSourceHead:     workgraph.MissionSourceHead,
		AtlasSourceHead:       workgraph.AtlasSourceHead,
		MissionBinarySHA256:   workgraph.MissionBinarySHA256,
		AtlasBinarySHA256:     workgraph.AtlasBinarySHA256,
		OperationalWorkgraphBinding: *workgraph.OperationalBinding,
	}
	return workgraph, binding
}

func testDigest(fill byte) string {
	return "sha256:" + string(bytes.Repeat([]byte{fill}, 64))
}

func testHead(fill byte) string {
	return string(bytes.Repeat([]byte{fill}, 40))
}

