package atlas

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"strings"
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
		{"atlas binary", "BINARY_DIGEST_MISMATCH_ATLAS", func(v *OperationalWorkgraphBindingDocument) { v.AtlasBinarySHA256 = testDigest('f') }},
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
		{"unsafe release", "SAFETY_BOUNDARY_UNSAFE_RELEASES_OR_PUBLISHES", func(v *OperationalWorkgraphBindingDocument) { v.SafetyBoundaries.ReleasesOrPublishes = true }},
		{"unsafe provider call", "SAFETY_BOUNDARY_UNSAFE_CALLS_PROVIDERS", func(v *OperationalWorkgraphBindingDocument) { v.SafetyBoundaries.CallsProviders = true }},
		{"unsafe authority", "SAFETY_BOUNDARY_UNSAFE_EXPANDS_AUTHORITY", func(v *OperationalWorkgraphBindingDocument) { v.SafetyBoundaries.ExpandsAuthority = true }},
		{"unsafe unplanned command", "SAFETY_BOUNDARY_UNSAFE_ALLOWS_UNPLANNED_COMMANDS", func(v *OperationalWorkgraphBindingDocument) { v.SafetyBoundaries.AllowsUnplannedCommands = true }},
		{"unsafe unbounded retry", "SAFETY_BOUNDARY_UNSAFE_ALLOWS_UNBOUNDED_RETRIES", func(v *OperationalWorkgraphBindingDocument) { v.SafetyBoundaries.AllowsUnboundedRetries = true }},
		{"unsafe child activation", "SAFETY_BOUNDARY_UNSAFE_ACTIVATES_CHILD_PROCESSES", func(v *OperationalWorkgraphBindingDocument) { v.SafetyBoundaries.ActivatesChildProcesses = true }},
		{"unsafe scheduling", "SAFETY_BOUNDARY_UNSAFE_SCHEDULES_WORK", func(v *OperationalWorkgraphBindingDocument) { v.SafetyBoundaries.SchedulesWork = true }},
		{"unsafe execution", "SAFETY_BOUNDARY_UNSAFE_EXECUTES_WORK", func(v *OperationalWorkgraphBindingDocument) { v.SafetyBoundaries.ExecutesWork = true }},
		{"unsafe approval", "SAFETY_BOUNDARY_UNSAFE_APPROVES_WORK", func(v *OperationalWorkgraphBindingDocument) { v.SafetyBoundaries.ApprovesWork = true }},
		{"unvalidated activation", "SAFETY_BOUNDARY_UNSAFE_ACTIVATION_REQUIRES_VALIDATED_BINDING", func(v *OperationalWorkgraphBindingDocument) {
			v.SafetyBoundaries.ActivationRequiresValidatedBinding = false
		}},
		{"phase clock drift", "SAFETY_BOUNDARY_UNSAFE_PRESERVES_PHASE_CLOCK_ON_RETRY", func(v *OperationalWorkgraphBindingDocument) {
			v.SafetyBoundaries.PreservesPhaseClockOnRetry = false
		}},
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

func TestOperationalGraphDigestBindsCompleteFactoryTask(t *testing.T) {
	mutations := []struct {
		name string
		edit func(*FactoryTask)
	}{
		{"objective", func(v *FactoryTask) { v.Objective += " tampered" }},
		{"target repository", func(v *FactoryTask) { v.TargetFactoryRepo += "-tampered" }},
		{"factory folder", func(v *FactoryTask) { v.FactoryFolder += "-tampered" }},
		{"mutation class", func(v *FactoryTask) { v.MutationClass += "-tampered" }},
		{"acceptance", func(v *FactoryTask) { v.Acceptance = append(v.Acceptance, "tampered") }},
		{"non goals", func(v *FactoryTask) { v.NonGoals = append(v.NonGoals, "tampered") }},
		{"write scope", func(v *FactoryTask) { v.WriteScope = append(v.WriteScope, "tampered") }},
		{"required gates", func(v *FactoryTask) { v.RequiredGates = append(v.RequiredGates, "tampered") }},
		{"rollback scope", func(v *FactoryTask) { v.RollbackScope = append(v.RollbackScope, "tampered") }},
		{"verification", func(v *FactoryTask) { v.Verification = append(v.Verification, "tampered") }},
		{"required evidence", func(v *FactoryTask) { v.RequiredEvidence = append(v.RequiredEvidence, "tampered") }},
		{"safety limits", func(v *FactoryTask) { v.SafetyLimits = append(v.SafetyLimits, "tampered") }},
		{"authority boundary", func(v *FactoryTask) { v.AuthorityBoundary += "-tampered" }},
		{"dependency refs", func(v *FactoryTask) { v.DependencyRefs = append(v.DependencyRefs, "tampered") }},
		{"context pack refs", func(v *FactoryTask) { v.ContextPackRefs = append(v.ContextPackRefs, "tampered") }},
	}
	for _, test := range mutations {
		t.Run(test.name, func(t *testing.T) {
			workgraph, binding := validOperationalWorkgraphAndBinding(t)
			test.edit(&workgraph.Nodes[0].FactoryTask)
			readback := ValidateOperationalWorkgraphBinding(workgraph, binding)
			if readback.ActivationAllowed || !containsString(readback.ConflictCodes, "GRAPH_BINDING_DIGEST_MISMATCH") {
				t.Fatalf("FactoryTask drift was not digest-denied: %#v", readback)
			}
		})
	}
}

func TestOperationalWorkgraphCannotDowngradeToLegacy(t *testing.T) {
	workgraph, binding := validOperationalWorkgraphAndBinding(t)
	workgraph.OperationalBinding = nil
	if err := ValidateWorkgraph(workgraph); err == nil || !strings.Contains(err.Error(), "OPERATIONAL_BINDING_MISSING") {
		t.Fatalf("partial operational Workgraph passed validation: %v", err)
	}

	dir := t.TempDir()
	path := dir + "/partial.json"
	if err := WriteJSON(path, workgraph); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadWorkgraph(path); err == nil || !strings.Contains(err.Error(), "operational_binding is required") {
		t.Fatalf("loader downgraded omitted operational binding: %v", err)
	}
	workgraph.Nodes[0].OperationalBinding = nil
	if err := ValidateWorkgraph(workgraph); err != nil {
		t.Fatalf("pre-existing lineage-only Workgraph lost legacy compatibility: %v", err)
	}
	readback := ValidateOperationalWorkgraphBinding(workgraph, binding)
	if readback.ActivationAllowed || !containsString(readback.ConflictCodes, "OPERATIONAL_BINDING_MISSING") {
		t.Fatalf("activation binding validator accepted a downgraded Workgraph: %#v", readback)
	}
	if err := WriteJSON(path, workgraph); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadWorkgraph(path); err != nil {
		t.Fatalf("legacy lineage-only Workgraph no longer loads: %v", err)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	withNull := bytes.Replace(body, []byte(`"nodes":`), []byte(`"operational_binding": null, "nodes":`), 1)
	if err := os.WriteFile(path, withNull, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadWorkgraph(path); err == nil || !strings.Contains(err.Error(), "must not be null") {
		t.Fatalf("loader downgraded null operational binding: %v", err)
	}
}

func TestOperationalWorkgraphRejectsRetryArithmeticAndScaleContradictions(t *testing.T) {
	tests := []struct {
		name string
		code string
		edit func(*Workgraph)
	}{
		{"retry cap exceeds capacity", "RETRY_POLICY_CAPACITY_EXCEEDED", func(v *Workgraph) { v.OperationalBinding.RetryPolicy.MaximumTotalRetries = 100 }},
		{"attempt cap exceeds Mission bound", "RETRY_POLICY_INVALID", func(v *Workgraph) { v.OperationalBinding.RetryPolicy.MaximumAttempts = 5 }},
		{"allowance differs from Mission formula", "PARTITION_RETRY_ALLOWANCE_MISMATCH", func(v *Workgraph) { v.OperationalBinding.PlannerPartitions[0].RetryAllowanceMS = 1000 }},
		{"allowance multiplication overflows", "PARTITION_RETRY_ALLOWANCE_OVERFLOW", func(v *Workgraph) {
			maxInt64 := int64(^uint64(0) >> 1)
			v.OperationalBinding.RetryPolicy.MaximumAttempts = 4
			v.OperationalBinding.PlannerPartitions[0].EstimatedDurationMS = maxInt64/4 + 1
			v.OperationalBinding.PlannerPartitions[0].RetryAllowanceMS = maxInt64
			v.OperationalBinding.PlannerPartitions[0].PerAttemptTimeoutMS = maxInt64
			v.OperationalBinding.PlannerPartitions[0].TotalNodeTimeoutMS = maxInt64
			v.OperationalBinding.PlannerPartitions[0].NodeBudgetMS = maxInt64
		}},
		{"allowance exceeds total timeout", "PARTITION_POLICY_CONTRADICTION", func(v *Workgraph) { v.OperationalBinding.PlannerPartitions[0].RetryAllowanceMS = 5000 }},
		{"unknown classification", "PARTITION_CLASSIFICATION_INVALID", func(v *Workgraph) { v.OperationalBinding.PlannerPartitions[0].Classification = "other" }},
		{"scale repeats more than once", "PARTITION_SCALE_REPEAT_INVALID", func(v *Workgraph) {
			v.OperationalBinding.PlannerPartitions[0].Classification = "scale"
			v.OperationalBinding.PlannerPartitions[0].ScaleDimension = "nodes=60"
		}},
		{"regular carries scale metadata", "PARTITION_REGULAR_SCALE_DIMENSION_PRESENT", func(v *Workgraph) {
			v.OperationalBinding.PlannerPartitions[0].ScaleDimension = "nodes=60"
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			workgraph, binding := validOperationalWorkgraphAndBinding(t)
			test.edit(&workgraph)
			readback := ValidateOperationalWorkgraphBinding(workgraph, binding)
			if readback.ActivationAllowed || !containsString(readback.ConflictCodes, test.code) {
				t.Fatalf("policy contradiction was not denied with %s: %#v", test.code, readback)
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

func TestOperationalWorkgraphCompleteCLIRoundTripPreservesTypedBinding(t *testing.T) {
	workgraph, _ := validOperationalWorkgraphAndBinding(t)
	dir := t.TempDir()
	evidenceRoot := dir + "/evidence"
	if err := os.MkdirAll(evidenceRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(evidenceRoot+"/result.json", []byte("{\"status\":\"passed\"}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	link, err := BuildEvidenceBoundRunLink(
		workgraph.Nodes[0].FactoryTask.ID,
		"completed",
		map[string]string{"result": "result.json"},
		evidenceRoot,
		"operational-round-trip",
	)
	if err != nil {
		t.Fatal(err)
	}
	workgraphPath := dir + "/workgraph.json"
	runLinkPath := dir + "/run-link.json"
	outputPath := dir + "/workgraph-after.json"
	if err := WriteJSON(workgraphPath, workgraph); err != nil {
		t.Fatal(err)
	}
	if err := WriteJSON(runLinkPath, link); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{
		"workgraph", "complete",
		"--workgraph", workgraphPath,
		"--run-link", runLinkPath,
		"--evidence-root", evidenceRoot,
		"--evidence-root-id", "operational-round-trip",
		"--out", outputPath,
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("operational complete failed: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	completed, err := LoadWorkgraph(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	wantBinding, err := json.Marshal(struct {
		Graph *OperationalWorkgraphBinding     `json:"graph"`
		Node  *WorkgraphNodeOperationalBinding `json:"node"`
	}{workgraph.OperationalBinding, workgraph.Nodes[0].OperationalBinding})
	if err != nil {
		t.Fatal(err)
	}
	gotBinding, err := json.Marshal(struct {
		Graph *OperationalWorkgraphBinding     `json:"graph"`
		Node  *WorkgraphNodeOperationalBinding `json:"node"`
	}{completed.OperationalBinding, completed.Nodes[0].OperationalBinding})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotBinding, wantBinding) {
		t.Fatalf("CLI round trip dropped operational binding")
	}
}

func TestOperationalWorkgraphBindingRejectsPartitionTopologyDefects(t *testing.T) {
	tests := []struct {
		name string
		code string
		edit func(*Workgraph)
	}{
		{"missing partitions", "PARTITION_MISSING", func(v *Workgraph) { v.OperationalBinding.PlannerPartitions = nil }},
		{"duplicate partition id", "PARTITION_DUPLICATE_ID", func(v *Workgraph) {
			v.OperationalBinding.PlannerPartitions = append(v.OperationalBinding.PlannerPartitions, v.OperationalBinding.PlannerPartitions[0])
		}},
		{"duplicate partition node", "PARTITION_DUPLICATE_NODE", func(v *Workgraph) {
			duplicate := v.OperationalBinding.PlannerPartitions[0]
			duplicate.PartitionID += "-duplicate"
			v.OperationalBinding.PlannerPartitions = append(v.OperationalBinding.PlannerPartitions, duplicate)
		}},
		{"unreferenced partition node", "PARTITION_UNREFERENCED", func(v *Workgraph) {
			v.OperationalBinding.PlannerPartitions[0].NodeID = "unknown-node"
		}},
		{"unknown node reference", "NODE_PARTITION_REFERENCE_UNKNOWN", func(v *Workgraph) {
			v.Nodes[0].OperationalBinding.PlannerPartitionID = "unknown-partition"
		}},
		{"missing node binding", "PARTITION_MISSING_FOR_NODE", func(v *Workgraph) {
			v.Nodes[0].OperationalBinding = nil
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			workgraph, binding := validOperationalWorkgraphAndBinding(t)
			test.edit(&workgraph)
			readback := ValidateOperationalWorkgraphBinding(workgraph, binding)
			if readback.ActivationAllowed || !containsString(readback.ConflictCodes, test.code) {
				t.Fatalf("topology defect was not denied with %s: %#v", test.code, readback)
			}
		})
	}
}

func TestOperationalBindingDocumentRejectsDuplicateAndMissingPartitions(t *testing.T) {
	workgraph, binding := validOperationalWorkgraphAndBinding(t)
	binding.PlannerPartitions = append(binding.PlannerPartitions, binding.PlannerPartitions[0])
	readback := ValidateOperationalWorkgraphBinding(workgraph, binding)
	if readback.ActivationAllowed || !containsString(readback.ConflictCodes, "PARTITION_DUPLICATE_ID") {
		t.Fatalf("binding-side duplicate partition was accepted: %#v", readback)
	}

	workgraph, binding = validOperationalWorkgraphAndBinding(t)
	binding.PlannerPartitions = nil
	readback = ValidateOperationalWorkgraphBinding(workgraph, binding)
	if readback.ActivationAllowed || !containsString(readback.ConflictCodes, "PARTITION_MISSING") {
		t.Fatalf("binding-side missing partition was accepted: %#v", readback)
	}
}

func TestOperationalWorkgraphStrictLoadRejectsUnknownAndMissingSafetyFields(t *testing.T) {
	workgraph, binding := validOperationalWorkgraphAndBinding(t)
	dir := t.TempDir()
	workgraphPath := dir + "/workgraph.json"
	bindingPath := dir + "/binding.json"
	if err := WriteJSON(workgraphPath, workgraph); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(workgraphPath)
	if err != nil {
		t.Fatal(err)
	}
	tampered := bytes.Replace(body, []byte(`"execution_profile_digest":`), []byte(`"unknown_operational_field": "denied", "execution_profile_digest":`), 1)
	if err := os.WriteFile(workgraphPath, tampered, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadWorkgraph(workgraphPath); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("strict operational load accepted unknown field: %v", err)
	}
	if err := WriteJSON(workgraphPath, workgraph); err != nil {
		t.Fatal(err)
	}
	body, err = os.ReadFile(workgraphPath)
	if err != nil {
		t.Fatal(err)
	}
	nestedUnknown := bytes.Replace(body, []byte(`"calls_providers": false,`), []byte(`"calls_providers": false, "unknown_safety_boundary": false,`), 1)
	if err := os.WriteFile(workgraphPath, nestedUnknown, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadWorkgraph(workgraphPath); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("strict operational load accepted nested unknown safety field: %v", err)
	}

	if err := WriteJSON(bindingPath, binding); err != nil {
		t.Fatal(err)
	}
	bindingBody, err := os.ReadFile(bindingPath)
	if err != nil {
		t.Fatal(err)
	}
	missingSafety := bytes.Replace(bindingBody, []byte(`    "calls_providers": false,`+"\n"), nil, 1)
	if err := os.WriteFile(bindingPath, missingSafety, 0o644); err != nil {
		t.Fatal(err)
	}
	loadedBinding, err := LoadOperationalWorkgraphBindingDocument(bindingPath)
	if err != nil {
		t.Fatal(err)
	}
	readback := ValidateOperationalWorkgraphBinding(workgraph, loadedBinding)
	if readback.ActivationAllowed || !containsString(readback.ConflictCodes, "SAFETY_BOUNDARY_MISSING_CALLS_PROVIDERS") {
		t.Fatalf("omitted safety boolean was not denied: %#v", readback)
	}

	if err := WriteJSON(bindingPath, binding); err != nil {
		t.Fatal(err)
	}
	bindingBody, err = os.ReadFile(bindingPath)
	if err != nil {
		t.Fatal(err)
	}
	missingRetryCap := bytes.Replace(
		bindingBody,
		[]byte(`    "maximum_attempts": 2,`+"\n"+`    "maximum_total_retries": 1`),
		[]byte(`    "maximum_attempts": 2`),
		1,
	)
	if err := os.WriteFile(bindingPath, missingRetryCap, 0o644); err != nil {
		t.Fatal(err)
	}
	loadedBinding, err = LoadOperationalWorkgraphBindingDocument(bindingPath)
	if err != nil {
		t.Fatal(err)
	}
	readback = ValidateOperationalWorkgraphBinding(workgraph, loadedBinding)
	if readback.ActivationAllowed || !containsString(readback.ConflictCodes, "RETRY_POLICY_MISSING_MAXIMUM_TOTAL_RETRIES") {
		t.Fatalf("omitted total retry cap was not denied: %#v", readback)
	}
}

func TestOperationalBindingRejectsOmittedZeroRetryAllowance(t *testing.T) {
	workgraph, binding := validOperationalWorkgraphAndBinding(t)
	workgraph.OperationalBinding.RetryPolicy = WorkgraphRetryPolicy{MaximumAttempts: 1, MaximumTotalRetries: 0}
	workgraph.OperationalBinding.PlannerPartitions[0].RetryAllowanceMS = workgraph.OperationalBinding.PlannerPartitions[0].EstimatedDurationMS
	digest, err := ComputeOperationalWorkgraphBindingDigest(workgraph)
	if err != nil {
		t.Fatal(err)
	}
	workgraph.OperationalBinding.GraphBindingDigest = digest
	binding.OperationalWorkgraphBinding = *workgraph.OperationalBinding
	binding.PlannerPartitions = append([]WorkgraphPlannerPartition(nil), workgraph.OperationalBinding.PlannerPartitions...)

	dir := t.TempDir()
	bindingPath := dir + "/binding.json"
	if err := WriteJSON(bindingPath, binding); err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(bindingPath)
	if err != nil {
		t.Fatal(err)
	}
	body = bytes.Replace(body, []byte(`      "retry_allowance_ms": 1000,`+"\n"), nil, 1)
	if err := os.WriteFile(bindingPath, body, 0o644); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadOperationalWorkgraphBindingDocument(bindingPath)
	if err != nil {
		t.Fatal(err)
	}
	readback := ValidateOperationalWorkgraphBinding(workgraph, loaded)
	if readback.ActivationAllowed || !containsString(readback.ConflictCodes, "PARTITION_FIELD_MISSING_RETRY_ALLOWANCE_MS") {
		t.Fatalf("omitted zero retry allowance was accepted: %#v", readback)
	}
}

func TestOperationalWorkgraphRejectsOmittedZeroValuedNodeFields(t *testing.T) {
	tests := []struct {
		name string
		line string
		code string
	}{
		{"dependencies", `      "dependencies": [],` + "\n", "NODE_FIELD_MISSING_DEPENDENCIES"},
		{"blockers", `      "blockers": [],` + "\n", "NODE_FIELD_MISSING_BLOCKERS"},
		{"stitch task", `      "stitch_task": false,` + "\n", "NODE_FIELD_MISSING_STITCH_TASK"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			workgraph, binding := validOperationalWorkgraphAndBinding(t)
			dir := t.TempDir()
			path := dir + "/workgraph.json"
			if err := WriteJSON(path, workgraph); err != nil {
				t.Fatal(err)
			}
			body, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			tampered := bytes.Replace(body, []byte(test.line), nil, 1)
			if bytes.Equal(tampered, body) {
				t.Fatalf("test did not remove %s", test.name)
			}
			if err := os.WriteFile(path, tampered, 0o644); err != nil {
				t.Fatal(err)
			}
			loaded, err := LoadWorkgraph(path)
			if err != nil {
				t.Fatal(err)
			}
			readback := ValidateOperationalWorkgraphBinding(loaded, binding)
			if readback.ActivationAllowed || !containsString(readback.ConflictCodes, test.code) {
				t.Fatalf("omitted %s was accepted: %#v", test.name, readback)
			}
		})
	}
}

func TestOperationalDocumentsRejectNullRequiredScalars(t *testing.T) {
	workgraph, binding := validOperationalWorkgraphAndBinding(t)
	workgraph.OperationalBinding.RetryPolicy = WorkgraphRetryPolicy{MaximumAttempts: 1, MaximumTotalRetries: 0}
	workgraph.OperationalBinding.PlannerPartitions[0].RetryAllowanceMS = workgraph.OperationalBinding.PlannerPartitions[0].EstimatedDurationMS
	digest, err := ComputeOperationalWorkgraphBindingDigest(workgraph)
	if err != nil {
		t.Fatal(err)
	}
	workgraph.OperationalBinding.GraphBindingDigest = digest
	binding.OperationalWorkgraphBinding = *workgraph.OperationalBinding
	binding.PlannerPartitions = append([]WorkgraphPlannerPartition(nil), workgraph.OperationalBinding.PlannerPartitions...)

	dir := t.TempDir()
	workgraphPath := dir + "/workgraph.json"
	bindingPath := dir + "/binding.json"
	if err := WriteJSON(workgraphPath, workgraph); err != nil {
		t.Fatal(err)
	}
	if err := WriteJSON(bindingPath, binding); err != nil {
		t.Fatal(err)
	}
	validBinding, err := os.ReadFile(bindingPath)
	if err != nil {
		t.Fatal(err)
	}
	bindingMutations := []struct {
		name string
		from string
		to   string
	}{
		{"safety bool", `"calls_providers": false`, `"calls_providers": null`},
		{"retry cap", `"maximum_total_retries": 0`, `"maximum_total_retries": null`},
		{"retry allowance", `"retry_allowance_ms": 1000`, `"retry_allowance_ms": null`},
	}
	for _, test := range bindingMutations {
		t.Run(test.name, func(t *testing.T) {
			tampered := bytes.Replace(validBinding, []byte(test.from), []byte(test.to), 1)
			if bytes.Equal(tampered, validBinding) {
				t.Fatalf("test did not replace %s", test.name)
			}
			if err := os.WriteFile(bindingPath, tampered, 0o644); err != nil {
				t.Fatal(err)
			}
			if _, err := LoadOperationalWorkgraphBindingDocument(bindingPath); err == nil || !strings.Contains(err.Error(), "must not be null") {
				t.Fatalf("binding accepted null %s: %v", test.name, err)
			}
		})
	}

	tamperedBinding := bytes.Replace(validBinding, []byte(`"calls_providers": false`), []byte(`"calls_providers": null`), 1)
	if err := os.WriteFile(bindingPath, tamperedBinding, 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"workgraph", "validate-binding", "--workgraph", workgraphPath, "--binding", bindingPath, "--json"}, &stdout, &stderr)
	if code == 0 {
		t.Fatalf("CLI accepted null safety bool: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	var readback OperationalWorkgraphBindingReadback
	if err := json.Unmarshal(stdout.Bytes(), &readback); err != nil {
		t.Fatalf("CLI null denial did not emit JSON: %v", err)
	}
	if !containsString(readback.ConflictCodes, "BINDING_PARSE_ERROR") || readback.ChildProcessLaunches != 0 {
		t.Fatalf("bad CLI null denial: %#v", readback)
	}

	validWorkgraph, err := os.ReadFile(workgraphPath)
	if err != nil {
		t.Fatal(err)
	}
	nullStitch := bytes.Replace(validWorkgraph, []byte(`"stitch_task": false`), []byte(`"stitch_task": null`), 1)
	if err := os.WriteFile(workgraphPath, nullStitch, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadWorkgraph(workgraphPath); err == nil || !strings.Contains(err.Error(), "must not be null") {
		t.Fatalf("Workgraph accepted null stitch_task: %v", err)
	}
}

func TestOperationalWorkgraphRejectsNonHexSourceHead(t *testing.T) {
	workgraph, binding := validOperationalWorkgraphAndBinding(t)
	workgraph.MissionSourceHead = strings.Repeat("z", 40)
	binding.MissionSourceHead = workgraph.MissionSourceHead
	readback := ValidateOperationalWorkgraphBinding(workgraph, binding)
	if readback.ActivationAllowed || !containsString(readback.ConflictCodes, "OPERATIONAL_FIELD_INVALID_MISSION_SOURCE_HEAD") {
		t.Fatalf("non-hex source head was not denied: %#v", readback)
	}
}

func TestOperationalWorkgraphMissionImportMustMatchMissionIdentity(t *testing.T) {
	workgraph, _ := validOperationalWorkgraphAndBinding(t)
	dir := t.TempDir()
	workgraphPath := dir + "/workgraph.json"
	importPath := dir + "/ao-mission-import.json"
	if err := WriteJSON(workgraphPath, workgraph); err != nil {
		t.Fatal(err)
	}
	if err := WriteJSON(importPath, AOMissionImport{
		ContractVersion: AOMissionImportContract,
		MissionID:       "mission-other",
		SourceArtifacts: []AOMissionSourceArtifact{{Name: "record", Path: "record.json", SHA256: testDigest('e')}},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := BuildAOMissionWorkgraphMetadata(importPath, workgraphPath); err == nil || !strings.Contains(err.Error(), "mission_id must match") {
		t.Fatalf("mismatched Mission import was accepted: %v", err)
	}
	metadata := AOMissionWorkgraphMetadata{
		ContractVersion:   AOMissionWorkgraphMetadataContract,
		MissionID:         "mission-other",
		WorkgraphID:       workgraph.ID,
		TargetInstance:    workgraph.TargetInstance,
		NodeCounts:        map[string]int{"total": 1, "ready": 1},
		SourceArtifacts:   map[string]string{"workgraph": testDigest('f')},
		MissionProvenance: map[string]int{"record": 1},
	}
	if err := ValidateAOMissionWorkgraphMetadata(metadata, workgraph); err == nil || !strings.Contains(err.Error(), "mission_id must match") {
		t.Fatalf("mismatched Mission metadata was accepted: %v", err)
	}
}

func TestOperationalWorkgraphProvenanceAugmentationFailsClosed(t *testing.T) {
	workgraph, _ := validOperationalWorkgraphAndBinding(t)
	importRecord := AOMissionImport{
		ContractVersion: AOMissionImportContract,
		MissionID:       workgraph.MissionID,
		SourceArtifacts: []AOMissionSourceArtifact{{Name: "record", Path: "record.json", SHA256: testDigest('e')}},
	}
	_, err := BuildAOMissionProvenanceWorkgraph(importRecord, workgraph)
	if err == nil || !strings.Contains(err.Error(), "would expand the approved planner partitions") {
		t.Fatalf("operational provenance augmentation did not fail closed: %v", err)
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

func TestOperationalWorkgraphValidateBindingCLIAllowsExactContract(t *testing.T) {
	workgraph, binding := validOperationalWorkgraphAndBinding(t)
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
	if code != 0 {
		t.Fatalf("valid binding failed: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	var readback OperationalWorkgraphBindingReadback
	if err := json.Unmarshal(stdout.Bytes(), &readback); err != nil {
		t.Fatalf("success stdout is not JSON: %v", err)
	}
	if !readback.ActivationAllowed || len(readback.ConflictCodes) != 0 || readback.ChildProcessLaunches != 0 {
		t.Fatalf("bad success readback: %#v", readback)
	}
}

func TestOperationalWorkgraphBindingDigestCLIUsesSourceOwnedCanonicalProjection(t *testing.T) {
	workgraph, _ := validOperationalWorkgraphAndBinding(t)
	want := workgraph.OperationalBinding.GraphBindingDigest
	workgraph.OperationalBinding.GraphBindingDigest = ""
	dir := t.TempDir()
	workgraphPath := dir + "/workgraph.json"
	if err := WriteJSON(workgraphPath, workgraph); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := Run([]string{"workgraph", "binding-digest", "--workgraph", workgraphPath, "--json"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("binding digest failed: stdout=%s stderr=%s", stdout.String(), stderr.String())
	}
	var readback OperationalWorkgraphBindingDigestReadback
	if err := json.Unmarshal(stdout.Bytes(), &readback); err != nil {
		t.Fatal(err)
	}
	if readback.GraphBindingDigest != want || readback.ChildProcessLaunches != 0 || readback.ExecutesWork || readback.SafeToExecute {
		t.Fatalf("bad source-owned binding digest readback: %#v want=%s", readback, want)
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
	workgraph.Nodes[0].Dependencies = []string{}
	workgraph.Nodes[0].Blockers = []string{}
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
		MutatesRepositories:                false,
		ReleasesOrPublishes:                false,
		CallsProviders:                     false,
		ExpandsAuthority:                   false,
		AllowsUnplannedCommands:            false,
		AllowsUnboundedRetries:             false,
		ActivationRequiresValidatedBinding: true,
		PreservesPhaseClockOnRetry:         true,
	}
	partition := WorkgraphPlannerPartition{
		PartitionID:          "partition-01",
		NodeID:               workgraph.Nodes[0].ID,
		TestID:               "atlas-test-01",
		Classification:       "regular",
		RequestedRepeatCount: 60,
		EffectiveRepeatCount: 60,
		EstimatedDurationMS:  1000,
		RetryAllowanceMS:     2000,
		PerAttemptTimeoutMS:  2000,
		TotalNodeTimeoutMS:   4000,
		NodeBudgetMS:         4000,
	}
	workgraph.Nodes[0].OperationalBinding = &WorkgraphNodeOperationalBinding{
		PlannerPartitionID: partition.PartitionID,
		TestID:             partition.TestID,
		Classification:     partition.Classification,
		SafetyBoundaries:   safety,
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
		ContractVersion:             OperationalWorkgraphBindingDocumentContract,
		WorkgraphID:                 workgraph.ID,
		TargetInstance:              workgraph.TargetInstance,
		MissionID:                   workgraph.MissionID,
		ObjectiveID:                 workgraph.ObjectiveID,
		ObjectiveDigest:             workgraph.ObjectiveDigest,
		CorrelationID:               workgraph.CorrelationID,
		SoakID:                      workgraph.SoakID,
		PlanID:                      workgraph.PlanID,
		PolicyDigest:                workgraph.PolicyDigest,
		ActivationID:                workgraph.ActivationID,
		EvidenceRootIdentity:        workgraph.EvidenceRootIdentity,
		MissionSourceHead:           workgraph.MissionSourceHead,
		AtlasSourceHead:             workgraph.AtlasSourceHead,
		MissionBinarySHA256:         workgraph.MissionBinarySHA256,
		AtlasBinarySHA256:           workgraph.AtlasBinarySHA256,
		OperationalWorkgraphBinding: *workgraph.OperationalBinding,
	}
	binding.PlannerPartitions = append([]WorkgraphPlannerPartition(nil), workgraph.OperationalBinding.PlannerPartitions...)
	return workgraph, binding
}

func testDigest(fill byte) string {
	return "sha256:" + string(bytes.Repeat([]byte{fill}, 64))
}

func testHead(fill byte) string {
	return string(bytes.Repeat([]byte{fill}, 40))
}
