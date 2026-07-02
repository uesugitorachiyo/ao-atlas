package atlas

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestInstanceInitValidateAndRegistry(t *testing.T) {
	dir := t.TempDir()
	instancePath := filepath.Join(dir, "instance.json")
	var out bytes.Buffer
	code := Run([]string{"instance", "init", "--id", "demo", "--state-root", ".atlas-local/state", "--toolchain-root", "../toolchain", "--out", instancePath}, &out, &out)
	if code != 0 {
		t.Fatalf("init failed: %s", out.String())
	}
	code = Run([]string{"instance", "validate", "--instance", instancePath}, &out, &out)
	if code != 0 {
		t.Fatalf("validate failed: %s", out.String())
	}
	registryPath := filepath.Join(dir, "foundry-registry.json")
	code = Run([]string{"instance", "registry", "emit", "--instance", instancePath, "--out", registryPath}, &out, &out)
	if code != 0 {
		t.Fatalf("registry failed: %s", out.String())
	}
	if _, err := os.Stat(registryPath); err != nil {
		t.Fatal(err)
	}
}

func TestProductionReadinessOpsWorkflowRunsAtlasReadinessGate(t *testing.T) {
	root := repoRoot(t)
	workflowPath := filepath.Join(root, ".github", "workflows", "production-readiness-ops.yml")
	content, err := os.ReadFile(workflowPath)
	if err != nil {
		t.Fatalf("read production-readiness-ops workflow: %v", err)
	}
	workflow := string(content)
	for _, want := range []string{
		"name: Production Readiness Ops",
		"workflow_dispatch:",
		"contents: read",
		"actions/checkout@v7",
		"actions/setup-go@v6",
		"scripts/production-readiness.sh",
	} {
		if !strings.Contains(workflow, want) {
			t.Fatalf("production-readiness-ops workflow missing %q", want)
		}
	}
	for _, forbidden := range []string{"gh release", "git push", "upload-artifact", "OPENAI_" + "API_KEY", "ANTHROPIC_" + "API_KEY"} {
		if strings.Contains(workflow, forbidden) {
			t.Fatalf("production-readiness-ops workflow contains forbidden capability %q", forbidden)
		}
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime caller unavailable")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func mustLoadJSON[T any](t *testing.T, path string) T {
	t.Helper()
	value, err := LoadJSON[T](path)
	if err != nil {
		t.Fatalf("load JSON %s: %v", path, err)
	}
	return value
}

func TestInstanceDoctorValidatesRootsAndRegistryParity(t *testing.T) {
	dir := t.TempDir()
	instancePath := filepath.Join(dir, "instance.json")
	registryPath := filepath.Join(dir, "registry.json")
	doctorPath := filepath.Join(dir, "doctor.json")
	var out bytes.Buffer
	if code := Run([]string{"instance", "init", "--id", "demo", "--state-root", ".atlas-local/state", "--toolchain-root", "../toolchain", "--out", instancePath}, &out, &out); code != 0 {
		t.Fatalf("init failed: %s", out.String())
	}
	if code := Run([]string{"instance", "registry", "emit", "--instance", instancePath, "--out", registryPath}, &out, &out); code != 0 {
		t.Fatalf("registry failed: %s", out.String())
	}
	if code := Run([]string{"instance", "doctor", "--instance", instancePath, "--registry", registryPath, "--out", doctorPath}, &out, &out); code != 0 {
		t.Fatalf("doctor failed: %s", out.String())
	}
	report, err := LoadJSON[InstanceDoctorReport](doctorPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateInstanceDoctorReport(report); err != nil {
		t.Fatal(err)
	}
	if report.Status != "ready" || report.InstanceID != "demo" {
		t.Fatalf("unexpected doctor report: %#v", report)
	}
}

func TestInstanceDoctorJSONWithoutRegistry(t *testing.T) {
	dir := t.TempDir()
	instancePath := filepath.Join(dir, "instance.json")
	var out bytes.Buffer
	if code := Run([]string{"instance", "init", "--id", "demo", "--state-root", ".atlas-local/state", "--toolchain-root", "../toolchain", "--out", instancePath}, &out, &out); code != 0 {
		t.Fatalf("init failed: %s", out.String())
	}
	out.Reset()
	if code := Run([]string{"instance", "doctor", "--instance", instancePath, "--json"}, &out, &out); code != 0 {
		t.Fatalf("doctor json failed: %s", out.String())
	}
	var report InstanceDoctorReport
	if err := json.Unmarshal(out.Bytes(), &report); err != nil {
		t.Fatalf("doctor did not emit json: %v\n%s", err, out.String())
	}
	if report.Status != "ready" || report.FirstFailingCheck != "" || report.ApprovesWork {
		t.Fatalf("unexpected doctor json report: %#v", report)
	}
}

func TestInstanceDoctorRejectsRegistryParityMismatch(t *testing.T) {
	dir := t.TempDir()
	instancePath := filepath.Join(dir, "instance.json")
	registryPath := filepath.Join(dir, "registry.json")
	var out bytes.Buffer
	if code := Run([]string{"instance", "init", "--id", "demo", "--state-root", ".atlas-local/state", "--toolchain-root", "../toolchain", "--out", instancePath}, &out, &out); code != 0 {
		t.Fatalf("init failed: %s", out.String())
	}
	if err := WriteJSON(registryPath, AtlasRegistry{
		ContractVersion: "ao.atlas.foundry-registry.v0.1",
		InstanceID:      "other-demo",
		ToolchainRoot:   "../toolchain",
		Roots:           map[string]string{"mission": ".atlas-local/state/demo/mission"},
		SchedulesWork:   false,
		ExecutesWork:    false,
	}); err != nil {
		t.Fatal(err)
	}
	code := Run([]string{"instance", "doctor", "--instance", instancePath, "--registry", registryPath, "--out", filepath.Join(dir, "doctor.json")}, &out, &out)
	if code == 0 {
		t.Fatal("expected registry parity mismatch to fail")
	}
	if !strings.Contains(out.String(), "registry instance_id") {
		t.Fatalf("expected registry parity error, got %s", out.String())
	}
}

func TestInstanceDoctorRejectsAuthorityClaims(t *testing.T) {
	instance := DefaultInstance("demo", ".atlas-local/state", "../toolchain")
	registry := AtlasRegistry{
		ContractVersion: AtlasRegistryContract,
		InstanceID:      "demo",
		ToolchainRoot:   "../toolchain",
		Roots:           instance.Roots,
		SchedulesWork:   true,
		ExecutesWork:    false,
		ApprovesWork:    false,
	}
	report, err := BuildInstanceDoctorReport(instance, registry)
	if err == nil {
		t.Fatal("expected authority claim to fail")
	}
	if report.Status != "failed" || report.FirstFailingCheck != "authority_boundary" {
		t.Fatalf("unexpected failed report: %#v", report)
	}
}

func TestIntakeUnderspecifiedEmitsBlueprintRequest(t *testing.T) {
	intake := Intake{ContractVersion: IntakeContract, ID: "short", BroadPrompt: "fix it"}
	request, err := ValidateIntake(intake)
	if err != nil {
		t.Fatal(err)
	}
	if request.Status != "blueprint_required" {
		t.Fatalf("expected blueprint request, got %#v", request)
	}
	if len(request.Missing) == 0 {
		t.Fatal("expected missing fields")
	}
	if err := ValidateBlueprintRequest(request); err != nil {
		t.Fatal(err)
	}
}

func TestBlueprintImportCompilesLowRiskCodePackIntoAtlasAndFoundryMaterial(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "import")
	var out bytes.Buffer
	code := Run([]string{
		"blueprint", "import",
		"--pack", filepath.Join("..", "..", "examples", "valid", "blueprint-import-low-risk-code", "blueprint-pack"),
		"--authorization", filepath.Join("..", "..", "examples", "valid", "blueprint-import-low-risk-code", "build-authorization.json"),
		"--instance", filepath.Join("..", "..", "examples", "valid", "stack-instance.json"),
		"--mutation-classes", filepath.Join("..", "..", "examples", "valid", "mutation-classes.json"),
		"--out", outDir,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("blueprint import failed: %s", out.String())
	}
	targetFolder := strings.Join([]string{"", "Users", "torachiyouesugi", "Documents", "public", "ao-foundry"}, "/")
	if !strings.Contains(out.String(), "foundry_continuation_prompt="+filepath.ToSlash(filepath.Join(outDir, "foundry-import", "foundry-continuation-prompt.md"))) ||
		!strings.Contains(out.String(), "Move to "+targetFolder) ||
		!strings.Contains(out.String(), "Run codex --yolo") ||
		!strings.Contains(out.String(), "Paste this prompt") ||
		strings.Contains(out.String(), "cat ") {
		t.Fatalf("blueprint import output did not report operator-ready continuation action:\n%s", out.String())
	}
	record := mustLoadJSON[BlueprintImport](t, filepath.Join(outDir, "blueprint-import.json"))
	if record.Status != "ready" || record.MutationClass != "low_risk_code" || record.LiveExecutionProven {
		t.Fatalf("unexpected import record: %#v", record)
	}
	if record.ReadyForFoundry != true || record.SafeToExecute != false || record.MutatesRepositories {
		t.Fatalf("unexpected authority flags: %#v", record)
	}
	for _, key := range []string{
		"blueprint_pack",
		"build_authorization",
		"implementation_spec",
		"quality_profile",
		"candidate_rules",
		"mutation_class_model",
		"candidate_selection",
		"workgraph",
		"downstream_foundry_import",
		"downstream_foundry_continuation_handoff",
	} {
		if record.Digests[key] == "" {
			t.Fatalf("record missing digest binding %q: %#v", key, record.Digests)
		}
	}
	if !containsValue(record.SafetyLimits, "do_not_advance:low_risk_code_live_execution_denied") {
		t.Fatalf("low_risk_code import must deny live execution: %#v", record.SafetyLimits)
	}
	if _, err := os.Stat(filepath.Join(outDir, "workgraph.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "candidate-selection.json")); err != nil {
		t.Fatal(err)
	}
	if record.DownstreamFoundryContinuationHandoff.Ref != "foundry-import/foundry-continuation-handoff.json" {
		t.Fatalf("missing continuation handoff ref: %#v", record.DownstreamFoundryContinuationHandoff)
	}
	foundryImport := mustLoadJSON[FoundryImport](t, filepath.Join(outDir, "foundry-import", "foundry-import.json"))
	if err := ValidateFoundryImport(foundryImport); err != nil {
		t.Fatal(err)
	}
	if foundryImport.Tasks[0].MutationClass != "low_risk_code" {
		t.Fatalf("unexpected Foundry import mutation class: %#v", foundryImport.Tasks[0])
	}
	handoff := mustLoadJSON[FoundryContinuationHandoff](t, filepath.Join(outDir, "foundry-import", "foundry-continuation-handoff.json"))
	if err := ValidateFoundryContinuationHandoff(handoff); err != nil {
		t.Fatal(err)
	}
	promptPath := filepath.Join(outDir, "foundry-import", "foundry-continuation-prompt.md")
	promptBytes, err := os.ReadFile(promptPath)
	if err != nil {
		t.Fatal(err)
	}
	prompt := string(promptBytes)
	for _, want := range []string{"Move to AO Foundry", "Run codex --yolo", "Paste this prompt"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("blueprint import continuation prompt missing %q:\n%s", want, prompt)
		}
	}
	if strings.Contains(prompt, "cat ") {
		t.Fatalf("blueprint import continuation prompt must not use cat as the primary next action:\n%s", prompt)
	}
}

func TestBlueprintImportAcceptsExternalCandidateRules(t *testing.T) {
	dir := t.TempDir()
	sourcePack := filepath.Join("..", "..", "examples", "valid", "blueprint-import-low-risk-code", "blueprint-pack")
	packPath := filepath.Join(dir, "blueprint-pack")
	copyDirExcept(t, sourcePack, packPath, "candidate-rules.json")
	candidateRulesPath := filepath.Join(sourcePack, "candidate-rules.json")
	outDir := filepath.Join(dir, "import")

	var out bytes.Buffer
	code := Run([]string{
		"blueprint", "import",
		"--pack", packPath,
		"--candidate-rules", candidateRulesPath,
		"--authorization", filepath.Join("..", "..", "examples", "valid", "blueprint-import-low-risk-code", "build-authorization.json"),
		"--instance", filepath.Join("..", "..", "examples", "valid", "stack-instance.json"),
		"--mutation-classes", filepath.Join("..", "..", "examples", "valid", "mutation-classes.json"),
		"--out", outDir,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("blueprint import failed: %s", out.String())
	}
	record := mustLoadJSON[BlueprintImport](t, filepath.Join(outDir, "blueprint-import.json"))
	if record.Status != "ready" || record.Digests["candidate_rules"] == "" {
		t.Fatalf("external candidate rules were not compiled: %#v", record)
	}
	if record.BlueprintPack.Digest != record.Digests["blueprint_pack"] {
		t.Fatalf("blueprint pack digest must stay bound to the source pack: %#v", record.Digests)
	}
	pack := mustLoadJSON[ContextPack](t, filepath.Join(outDir, "context-packs", "low-risk-code-rehearsal-candidate-context-pack.json"))
	if !contextPackHasSourceRef(pack, publicArtifactRef(candidateRulesPath)) {
		t.Fatalf("context pack must preserve external candidate rules ref, got %#v", pack.SourceRefs)
	}
}

func TestBlueprintImportBlocksWithoutAuthorization(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "blocked")
	var out bytes.Buffer
	code := Run([]string{
		"blueprint", "import",
		"--pack", filepath.Join("..", "..", "examples", "invalid", "blueprint-import-missing-authorization", "blueprint-pack"),
		"--instance", filepath.Join("..", "..", "examples", "valid", "stack-instance.json"),
		"--mutation-classes", filepath.Join("..", "..", "examples", "valid", "mutation-classes.json"),
		"--out", outDir,
	}, &out, &out)
	if code == 0 {
		t.Fatal("expected missing authorization to fail closed")
	}
	record := mustLoadJSON[BlueprintImport](t, filepath.Join(outDir, "blueprint-import.json"))
	if record.Status != "blocked" || record.ReadyForFoundry || record.SafeToExecute {
		t.Fatalf("unexpected blocked import record: %#v", record)
	}
	request := mustLoadJSON[BlueprintRequest](t, filepath.Join(outDir, "blueprint-request.json"))
	if request.Status != "blueprint_required" || !containsValue(request.Missing, "build_authorization") {
		t.Fatalf("unexpected request: %#v", request)
	}
	if _, err := os.Stat(filepath.Join(outDir, "workgraph.json")); !os.IsNotExist(err) {
		t.Fatalf("blocked import must not write ready workgraph: %v", err)
	}
}

func copyDirExcept(t *testing.T, src, dst, excludedBase string) {
	t.Helper()
	if err := filepath.WalkDir(src, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(dst, 0o755)
		}
		if filepath.Base(path) == excludedBase {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		target := filepath.Join(dst, rel)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	}); err != nil {
		t.Fatalf("copy fixture pack: %v", err)
	}
}

func contextPackHasSourceRef(pack ContextPack, want string) bool {
	for _, ref := range pack.SourceRefs {
		if ref.Ref == want {
			return true
		}
	}
	return false
}

func blueprintCompilerValidPaths(outDir string) BlueprintImportPaths {
	return BlueprintImportPaths{
		PackPath:            filepath.Join("..", "..", "examples", "valid", "blueprint-import-low-risk-code", "blueprint-pack"),
		AuthorizationPath:   filepath.Join("..", "..", "examples", "valid", "blueprint-import-low-risk-code", "build-authorization.json"),
		InstancePath:        filepath.Join("..", "..", "examples", "valid", "stack-instance.json"),
		MutationClassesPath: filepath.Join("..", "..", "examples", "valid", "mutation-classes.json"),
		OutDir:              outDir,
	}
}

func TestBlueprintImportBlocksStaleAuthorization(t *testing.T) {
	dir := t.TempDir()
	authPath := filepath.Join(dir, "expired-authorization.json")
	auth := mustLoadJSON[map[string]any](t, filepath.Join("..", "..", "examples", "valid", "blueprint-import-low-risk-code", "build-authorization.json"))
	auth["expires_at_utc"] = "2000-01-01T00:00:00Z"
	if err := WriteJSON(authPath, auth); err != nil {
		t.Fatal(err)
	}
	outDir := filepath.Join(dir, "blocked")
	var out bytes.Buffer
	code := Run([]string{
		"blueprint", "import",
		"--pack", filepath.Join("..", "..", "examples", "valid", "blueprint-import-low-risk-code", "blueprint-pack"),
		"--authorization", authPath,
		"--instance", filepath.Join("..", "..", "examples", "valid", "stack-instance.json"),
		"--mutation-classes", filepath.Join("..", "..", "examples", "valid", "mutation-classes.json"),
		"--out", outDir,
	}, &out, &out)
	if code == 0 {
		t.Fatal("expected stale authorization to fail closed")
	}
	request := mustLoadJSON[BlueprintRequest](t, filepath.Join(outDir, "blueprint-request.json"))
	if !containsValue(request.Missing, "build_authorization_freshness") {
		t.Fatalf("expected stale authorization blocker, got %#v", request.Missing)
	}
	if _, err := os.Stat(filepath.Join(outDir, "foundry-import", "foundry-import.json")); !os.IsNotExist(err) {
		t.Fatalf("blocked import must not write Foundry import: %v", err)
	}
}

func TestBlueprintReadyArtifactWriterLivesInDedicatedModule(t *testing.T) {
	module, err := os.ReadFile("blueprint_ready_artifacts.go")
	if err != nil {
		t.Fatalf("expected dedicated Blueprint ready artifact module: %v", err)
	}
	content := string(module)
	for _, want := range []string{
		"func writeBlueprintReadyArtifacts(",
		"func writeBlueprintFoundryImportArtifacts(",
		"WriteFoundryContinuationPrompt(",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("dedicated ready artifact module missing %q", want)
		}
	}
}

func TestBlueprintImportPersistenceLivesInDedicatedModule(t *testing.T) {
	module, err := os.ReadFile("blueprint_import_persistence.go")
	if err != nil {
		t.Fatalf("expected dedicated Blueprint import persistence module: %v", err)
	}
	content := string(module)
	for _, want := range []string{
		"func persistBlueprintImportArtifacts(",
		"writeBlueprintBlockedArtifacts(",
		"return compileErr",
		"blueprint compiler must emit exactly one context pack",
		"writeBlueprintReadyArtifacts(",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("dedicated import persistence module missing %q", want)
		}
	}
}

func TestBlueprintImportInputValidationLivesInDedicatedModule(t *testing.T) {
	module, err := os.ReadFile("blueprint_import_input_validation.go")
	if err != nil {
		t.Fatalf("expected dedicated Blueprint import input validation module: %v", err)
	}
	content := string(module)
	for _, want := range []string{
		"func validateBlueprintImportInputs(",
		"strings.TrimSpace(paths.OutDir)",
		"errors.New(\"--out is required\")",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("dedicated import input validation module missing %q", want)
		}
	}
}

func TestBlueprintAuthorizationReadinessLivesInDedicatedModule(t *testing.T) {
	module, err := os.ReadFile("blueprint_authorization.go")
	if err != nil {
		t.Fatalf("expected dedicated Blueprint authorization module: %v", err)
	}
	content := string(module)
	for _, want := range []string{
		"func validateBlueprintAuthorization(",
		"func mutationModelIncludes(",
		"time.Parse(time.RFC3339",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("dedicated authorization module missing %q", want)
		}
	}
}

func TestBlueprintArtifactBuildersLiveInDedicatedModule(t *testing.T) {
	module, err := os.ReadFile("blueprint_artifact_builders.go")
	if err != nil {
		t.Fatalf("expected dedicated Blueprint artifact builders module: %v", err)
	}
	content := string(module)
	for _, want := range []string{
		"func buildBlueprintIntake(",
		"func buildBlueprintContextPack(",
		"func buildBlueprintFactoryTask(",
		"func buildBlueprintCandidateSelection(",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("dedicated artifact builders module missing %q", want)
		}
	}
}

func TestBlueprintSourceUtilitiesLiveInDedicatedModule(t *testing.T) {
	module, err := os.ReadFile("blueprint_sources.go")
	if err != nil {
		t.Fatalf("expected dedicated Blueprint sources module: %v", err)
	}
	content := string(module)
	for _, want := range []string{
		"func readJSONIfPossible(",
		"func digestFile(",
		"func digestDirectory(",
		"func publicArtifactRef(",
		"func copyStringMap(",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("dedicated source utilities module missing %q", want)
		}
	}
}

func TestBlueprintImportValidationLivesInDedicatedModule(t *testing.T) {
	module, err := os.ReadFile("blueprint_import_validation.go")
	if err != nil {
		t.Fatalf("expected dedicated Blueprint import validation module: %v", err)
	}
	content := string(module)
	for _, want := range []string{
		"func ValidateBlueprintImport(",
		"ready_for_foundry must be true when status is ready",
		"safe_to_execute must be false",
		"release_or_publish_allowed must be false",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("dedicated import validation module missing %q", want)
		}
	}
}

func TestBlueprintCandidateRulesValidationLivesInDedicatedModule(t *testing.T) {
	module, err := os.ReadFile("blueprint_candidate_rules_validation.go")
	if err != nil {
		t.Fatalf("expected dedicated Blueprint candidate rules validation module: %v", err)
	}
	content := string(module)
	for _, want := range []string{
		"func ValidateBlueprintCandidateRules(",
		"mutation_class must be one of the required mutation classes",
		"checkPublicStrings(&errs, \"required_evidence\"",
		"checkPublicStrings(&errs, \"context_refs\"",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("dedicated candidate rules validation module missing %q", want)
		}
	}
}

func TestBlueprintBlockedArtifactWriterLivesInDedicatedModule(t *testing.T) {
	module, err := os.ReadFile("blueprint_blocked_artifacts.go")
	if err != nil {
		t.Fatalf("expected dedicated Blueprint blocked artifact module: %v", err)
	}
	content := string(module)
	for _, want := range []string{
		"func writeBlueprintBlockedArtifacts(",
		"ValidateBlueprintRequest(request)",
		"ValidateBlueprintImport(record)",
		"blueprint-request.json",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("dedicated blocked artifact module missing %q", want)
		}
	}
}

func TestBlueprintCompileStateLivesInDedicatedModule(t *testing.T) {
	module, err := os.ReadFile("blueprint_compile_state.go")
	if err != nil {
		t.Fatalf("expected dedicated Blueprint compile state module: %v", err)
	}
	content := string(module)
	for _, want := range []string{
		"func newBlockedBlueprintCompileState(",
		"digestDirectory(paths.PackPath)",
		"missing-blueprint-pack:",
		"BlueprintImport{",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("dedicated compile state module missing %q", want)
		}
	}
}

func TestBlueprintCandidateRulesLoaderLivesInDedicatedModule(t *testing.T) {
	module, err := os.ReadFile("blueprint_candidate_rules_loader.go")
	if err != nil {
		t.Fatalf("expected dedicated Blueprint candidate rules loader module: %v", err)
	}
	content := string(module)
	for _, want := range []string{
		"func loadBlueprintCandidateRules(",
		"candidate-rules.json",
		"readJSONIfPossible(rulesPath, &rules)",
		"ValidateBlueprintCandidateRules(rules)",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("dedicated candidate rules loader module missing %q", want)
		}
	}
}

func TestBlueprintRequiredArtifactsLiveInDedicatedModule(t *testing.T) {
	module, err := os.ReadFile("blueprint_required_artifacts.go")
	if err != nil {
		t.Fatalf("expected dedicated Blueprint required artifacts module: %v", err)
	}
	content := string(module)
	for _, want := range []string{
		"func loadBlueprintRequiredArtifacts(",
		"implementation-spec.md",
		"quality-profile.md",
		"digestFile(path)",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("dedicated required artifacts module missing %q", want)
		}
	}
}

func TestBlueprintInstanceLoaderLivesInDedicatedModule(t *testing.T) {
	module, err := os.ReadFile("blueprint_instance_loader.go")
	if err != nil {
		t.Fatalf("expected dedicated Blueprint instance loader module: %v", err)
	}
	content := string(module)
	for _, want := range []string{
		"func loadBlueprintInstance(",
		"readJSONIfPossible(paths.InstancePath, &instance)",
		"ValidateInstance(instance)",
		"stack instance id must match candidate target_instance",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("dedicated instance loader module missing %q", want)
		}
	}
}

func TestBlueprintMutationModelLoaderLivesInDedicatedModule(t *testing.T) {
	module, err := os.ReadFile("blueprint_mutation_model_loader.go")
	if err != nil {
		t.Fatalf("expected dedicated Blueprint mutation model loader module: %v", err)
	}
	content := string(module)
	for _, want := range []string{
		"func loadBlueprintMutationModel(",
		"readJSONIfPossible(paths.MutationClassesPath, &mutationModel)",
		"ValidateMutationClassModel(mutationModel)",
		"mutation class model must include ",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("dedicated mutation model loader module missing %q", want)
		}
	}
}

func TestBlueprintAuthorizationLoaderLivesInDedicatedModule(t *testing.T) {
	module, err := os.ReadFile("blueprint_authorization_loader.go")
	if err != nil {
		t.Fatalf("expected dedicated Blueprint authorization loader module: %v", err)
	}
	content := string(module)
	for _, want := range []string{
		"func loadBlueprintAuthorization(",
		"readJSONIfPossible(paths.AuthorizationPath, &authorization)",
		"validateBlueprintAuthorization(authorization, rules, packDigest)",
		"record.BuildAuthorization = SourceRef{Ref: publicArtifactRef(paths.AuthorizationPath), Digest: authDigest}",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("dedicated authorization loader module missing %q", want)
		}
	}
}

func TestBlueprintBlockedRequestLivesInDedicatedModule(t *testing.T) {
	module, err := os.ReadFile("blueprint_blocked_request.go")
	if err != nil {
		t.Fatalf("expected dedicated Blueprint blocked request module: %v", err)
	}
	content := string(module)
	for _, want := range []string{
		"func buildBlueprintBlockedRequest(",
		"Status:          \"blueprint_required\"",
		"record.BlockingNextActions = uniqueStrings(blockers)",
		"return to AO Blueprint for build authorization",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("dedicated blocked request module missing %q", want)
		}
	}
}

func TestBlueprintWorkgraphBuilderLivesInDedicatedModule(t *testing.T) {
	module, err := os.ReadFile("blueprint_workgraph_builder.go")
	if err != nil {
		t.Fatalf("expected dedicated Blueprint workgraph builder module: %v", err)
	}
	content := string(module)
	for _, want := range []string{
		"func buildBlueprintWorkgraph(",
		"ContractVersion: WorkgraphContract",
		"FactoryTask:  task",
		"ValidateWorkgraph(workgraph)",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("dedicated workgraph builder module missing %q", want)
		}
	}
}

func TestBlueprintFoundrySourcesLiveInDedicatedModule(t *testing.T) {
	module, err := os.ReadFile("blueprint_foundry_sources.go")
	if err != nil {
		t.Fatalf("expected dedicated Blueprint Foundry sources module: %v", err)
	}
	content := string(module)
	for _, want := range []string{
		"func buildBlueprintFoundrySourceArtifacts(",
		"candidate-selection.json",
		"context-packs/\" + contextPack.ID + \".json",
		"workgraph.json",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("dedicated Foundry sources module missing %q", want)
		}
	}
}

func TestBlueprintDownstreamFoundryLivesInDedicatedModule(t *testing.T) {
	module, err := os.ReadFile("blueprint_downstream_foundry.go")
	if err != nil {
		t.Fatalf("expected dedicated Blueprint downstream Foundry module: %v", err)
	}
	content := string(module)
	for _, want := range []string{
		"func buildBlueprintDownstreamFoundry(",
		"BuildFoundryImportForNodes(workgraph, nil, sourceArtifacts)",
		"BuildFoundryContinuationHandoff(workgraph, foundryImport",
		"FoundryImportPath: \"foundry-import/foundry-import.json\"",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("dedicated downstream Foundry module missing %q", want)
		}
	}
}

func TestBlueprintReadyRecordLivesInDedicatedModule(t *testing.T) {
	module, err := os.ReadFile("blueprint_ready_record.go")
	if err != nil {
		t.Fatalf("expected dedicated Blueprint ready record module: %v", err)
	}
	content := string(module)
	for _, want := range []string{
		"func buildBlueprintReadyRecord(",
		"record.Status = \"ready\"",
		"record.DownstreamFoundryImport = SourceRef{Ref: \"foundry-import/foundry-import.json\"",
		"record.ReadyForFoundry = true",
		"ValidateBlueprintImport(record)",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("dedicated ready record module missing %q", want)
		}
	}
}

func TestBlueprintReadyMaterialLivesInDedicatedModule(t *testing.T) {
	module, err := os.ReadFile("blueprint_ready_material.go")
	if err != nil {
		t.Fatalf("expected dedicated Blueprint ready material module: %v", err)
	}
	content := string(module)
	for _, want := range []string{
		"func buildBlueprintReadyMaterial(",
		"buildBlueprintContextPack(",
		"buildBlueprintWorkgraph(",
		"buildBlueprintCandidateSelection(",
		"buildBlueprintDownstreamFoundry(",
		"buildBlueprintReadyRecord(",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("dedicated ready material module missing %q", want)
		}
	}
}

func TestBlueprintSourceLoadingLivesInDedicatedModule(t *testing.T) {
	module, err := os.ReadFile("blueprint_source_loading.go")
	if err != nil {
		t.Fatalf("expected dedicated Blueprint source loading module: %v", err)
	}
	content := string(module)
	for _, want := range []string{
		"func loadBlueprintCompileSources(",
		"loadBlueprintCandidateRules(",
		"loadBlueprintRequiredArtifacts(",
		"loadBlueprintInstance(",
		"loadBlueprintMutationModel(",
		"loadBlueprintAuthorization(",
		"blueprint_pack",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("dedicated source loading module missing %q", want)
		}
	}
}

func TestBlueprintCompileArtifactsLiveInDedicatedModule(t *testing.T) {
	module, err := os.ReadFile("blueprint_compile_artifacts.go")
	if err != nil {
		t.Fatalf("expected dedicated Blueprint compile artifacts module: %v", err)
	}
	content := string(module)
	for _, want := range []string{
		"type BlueprintCompileArtifacts struct",
		"func blueprintCompileArtifactsToResult(",
		"Record:        artifacts.Record",
		"FoundryImport: artifacts.FoundryImport",
		"Handoff:       artifacts.Handoff",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("dedicated compile artifacts module missing %q", want)
		}
	}
}

func TestBlueprintBlockedCompileLivesInDedicatedModule(t *testing.T) {
	module, err := os.ReadFile("blueprint_blocked_compile.go")
	if err != nil {
		t.Fatalf("expected dedicated Blueprint blocked compile module: %v", err)
	}
	content := string(module)
	for _, want := range []string{
		"func buildBlueprintBlockedCompileArtifacts(",
		"buildBlueprintBlockedRequest(",
		"artifacts.Record = blockedRequest.Record",
		"artifacts.Request = blockedRequest.Request",
		"func blueprintBlockedCompileError(",
		"blueprint import blocked:",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("dedicated blocked compile module missing %q", want)
		}
	}
}

func TestBlueprintCompilerContractLivesInDedicatedModule(t *testing.T) {
	module, err := os.ReadFile("blueprint_compiler_contract.go")
	if err != nil {
		t.Fatalf("expected dedicated Blueprint compiler contract module: %v", err)
	}
	content := string(module)
	for _, want := range []string{
		"type BlueprintCompileInputs struct",
		"Paths BlueprintImportPaths",
		"type BlueprintCompiler struct",
		"Inputs BlueprintCompileInputs",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("dedicated compiler contract module missing %q", want)
		}
	}
}

func TestBlueprintReadyCompileLivesInDedicatedModule(t *testing.T) {
	module, err := os.ReadFile("blueprint_ready_compile.go")
	if err != nil {
		t.Fatalf("expected dedicated Blueprint ready compile module: %v", err)
	}
	content := string(module)
	for _, want := range []string{
		"func buildBlueprintReadyCompileArtifacts(",
		"buildBlueprintReadyMaterial(",
		"Rules:      sourceLoad.Rules",
		"AuthDigest: sourceLoad.AuthDigest",
		"return buildBlueprintReadyMaterial(",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("dedicated ready compile module missing %q", want)
		}
	}
}

func TestBlueprintCompilerBlocksWithoutAuthorizationWithoutReadyArtifacts(t *testing.T) {
	paths := blueprintCompilerValidPaths("")
	paths.AuthorizationPath = ""

	artifacts, err := BlueprintCompiler{Inputs: BlueprintCompileInputs{Paths: paths}}.Compile()
	if err == nil || !strings.Contains(err.Error(), "blueprint import blocked") {
		t.Fatalf("expected blocked compiler result, got err=%v artifacts=%#v", err, artifacts)
	}
	if artifacts.Record.Status != "blocked" || artifacts.Record.ReadyForFoundry || artifacts.Record.SafeToExecute {
		t.Fatalf("unexpected blocked compiler record: %#v", artifacts.Record)
	}
	if artifacts.Request.Status != "blueprint_required" || !containsValue(artifacts.Request.Missing, "build_authorization") {
		t.Fatalf("unexpected blocked compiler request: %#v", artifacts.Request)
	}
	if artifacts.Workgraph.ID != "" || artifacts.FoundryImport.ID != "" || artifacts.Handoff.ID != "" {
		t.Fatalf("blocked compiler must not emit ready artifacts: %#v", artifacts)
	}
}

func TestBlueprintCompilerReadyArtifactsRemainNoExecution(t *testing.T) {
	artifacts, err := BlueprintCompiler{Inputs: BlueprintCompileInputs{Paths: blueprintCompilerValidPaths("")}}.Compile()
	if err != nil {
		t.Fatal(err)
	}
	if artifacts.Record.Status != "ready" || !artifacts.Record.ReadyForFoundry {
		t.Fatalf("expected ready compiler record: %#v", artifacts.Record)
	}
	if artifacts.Record.SafeToExecute || artifacts.Record.SchedulesWork || artifacts.Record.ExecutesWork ||
		artifacts.Record.ApprovesWork || artifacts.Record.MutatesRepositories || artifacts.Record.CallsProviders ||
		artifacts.Record.ReleaseOrPublishAllowed {
		t.Fatalf("compiled Blueprint material must remain Atlas no-execution only: %#v", artifacts.Record)
	}
	if artifacts.FoundryImport.ID == "" || len(artifacts.FoundryImport.Tasks) != 1 {
		t.Fatalf("compiler must emit downstream Foundry import material: %#v", artifacts.FoundryImport)
	}
	if artifacts.Handoff.ID == "" || !strings.Contains(artifacts.Handoff.Prompt, "Paste this prompt") {
		t.Fatalf("compiler must emit operator-ready Foundry continuation handoff: %#v", artifacts.Handoff)
	}
}

func TestBlueprintCompilerPreservesExternalCandidateRulesRef(t *testing.T) {
	dir := t.TempDir()
	sourcePack := filepath.Join("..", "..", "examples", "valid", "blueprint-import-low-risk-code", "blueprint-pack")
	packPath := filepath.Join(dir, "blueprint-pack")
	copyDirExcept(t, sourcePack, packPath, "candidate-rules.json")
	candidateRulesPath := filepath.Join(sourcePack, "candidate-rules.json")
	paths := blueprintCompilerValidPaths("")
	paths.PackPath = packPath
	paths.CandidateRulesPath = candidateRulesPath

	artifacts, err := BlueprintCompiler{Inputs: BlueprintCompileInputs{Paths: paths}}.Compile()
	if err != nil {
		t.Fatal(err)
	}
	if len(artifacts.ContextPacks) != 1 || !contextPackHasSourceRef(artifacts.ContextPacks[0], publicArtifactRef(candidateRulesPath)) {
		t.Fatalf("compiler must preserve external candidate-rules ref, got %#v", artifacts.ContextPacks)
	}
}

func TestMissionStatusSummarizesIntakeWorkgraphAndRunLinks(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "mission-status.json")
	var out bytes.Buffer
	code := Run([]string{
		"mission", "status",
		"--intake", filepath.Join("..", "..", "examples", "valid", "intake.json"),
		"--workgraph", filepath.Join("..", "..", "examples", "valid", "workgraph-completed.json"),
		"--run-link", filepath.Join("..", "..", "examples", "valid", "run-link.json"),
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("mission status failed: %s", out.String())
	}
	status, err := LoadJSON[MissionStatus](outPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateMissionStatus(status); err != nil {
		t.Fatal(err)
	}
	if status.IntakeID != "atlas-intake-demo" || status.WorkgraphID != "atlas-readiness-workgraph" {
		t.Fatalf("unexpected mission status: %#v", status)
	}
	if status.CompletionStatus != "completed" || status.NodeCounts["completed"] != 2 {
		t.Fatalf("expected completed mission status, got %#v", status)
	}
}

func TestMissionStatusReportsBlockedWhenRunLinkBlocked(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "mission-status.json")
	var out bytes.Buffer
	code := Run([]string{
		"mission", "status",
		"--intake", filepath.Join("..", "..", "examples", "valid", "intake.json"),
		"--workgraph", filepath.Join("..", "..", "examples", "valid", "workgraph.json"),
		"--run-link", filepath.Join("..", "..", "examples", "invalid", "run-link-blocked.json"),
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("mission status failed: %s", out.String())
	}
	status, err := LoadJSON[MissionStatus](outPath)
	if err != nil {
		t.Fatal(err)
	}
	if status.CompletionStatus != "blocked" {
		t.Fatalf("expected blocked mission status, got %#v", status)
	}
}

func TestMissionStatusJSONReportsMissingContextAndHandoffs(t *testing.T) {
	var out bytes.Buffer
	code := Run([]string{
		"mission", "status",
		"--intake", filepath.Join("..", "..", "examples", "valid", "intake.json"),
		"--workgraph", filepath.Join("..", "..", "examples", "valid", "workgraph.json"),
		"--run-link", filepath.Join("..", "..", "examples", "valid", "run-link-needs-context.json"),
		"--json",
	}, &out, &out)
	if code != 0 {
		t.Fatalf("mission status json failed: %s", out.String())
	}
	var status MissionStatus
	if err := json.Unmarshal(out.Bytes(), &status); err != nil {
		t.Fatalf("mission status did not emit json: %v\n%s", err, out.String())
	}
	if status.CompletionStatus != "blocked" {
		t.Fatalf("expected blocked mission status, got %#v", status)
	}
	if len(status.MissingContextPacks) != 1 || status.MissingContextPacks[0] != "atlas-readiness-task" {
		t.Fatalf("expected missing context pack for run link, got %#v", status.MissingContextPacks)
	}
	if len(status.MissingHandoffs) != 0 {
		t.Fatalf("did not expect missing handoff when run-link exists, got %#v", status.MissingHandoffs)
	}
	if status.NextRecommendedAction != "repack missing context before Foundry handoff" {
		t.Fatalf("unexpected next recommended action: %#v", status.NextRecommendedAction)
	}
}

func TestMissionStatusJSONReportsMissingHandoffs(t *testing.T) {
	var out bytes.Buffer
	code := Run([]string{
		"mission", "status",
		"--intake", filepath.Join("..", "..", "examples", "valid", "intake.json"),
		"--workgraph", filepath.Join("..", "..", "examples", "valid", "workgraph.json"),
		"--json",
	}, &out, &out)
	if code != 0 {
		t.Fatalf("mission status json failed: %s", out.String())
	}
	var status MissionStatus
	if err := json.Unmarshal(out.Bytes(), &status); err != nil {
		t.Fatalf("mission status did not emit json: %v\n%s", err, out.String())
	}
	if len(status.MissingHandoffs) != 1 || status.MissingHandoffs[0] != "atlas-readiness-task" {
		t.Fatalf("expected missing handoff for ready node, got %#v", status.MissingHandoffs)
	}
	if status.NextRecommendedAction != "emit Foundry handoff for ready nodes" {
		t.Fatalf("unexpected next recommended action: %#v", status.NextRecommendedAction)
	}
}

func TestWorkgraphStateCountsAndExecutableReadiness(t *testing.T) {
	workgraph := fixtureWorkgraph()

	state, err := BuildWorkgraphState(workgraph)
	if err != nil {
		t.Fatal(err)
	}

	if state.NodeCounts["completed"] != 1 || state.NodeCounts["ready"] != 2 || state.NodeCounts["blocked"] != 1 {
		t.Fatalf("unexpected node counts: %#v", state.NodeCounts)
	}
	blocked, ok := state.NodeState("task-blocked")
	if !ok {
		t.Fatal("expected task-blocked node state")
	}
	if blocked.ExecutableReady || blocked.DependenciesComplete {
		t.Fatalf("blocked node must not be executable-ready: %#v", blocked)
	}
	if got := state.ExecutableReadyNodeIDs; len(got) != 2 || got[0] != "task-ready" || got[1] != "task-ready-2" {
		t.Fatalf("expected dependency-ready nodes in workgraph order, got %#v", got)
	}
}

func TestWorkgraphStateSkipsReadyNodeWithIncompleteDependency(t *testing.T) {
	workgraph := fixtureWorkgraph()
	workgraph.Nodes[1].Dependencies = []string{"task-blocked"}

	state, err := BuildWorkgraphState(workgraph)
	if err != nil {
		t.Fatal(err)
	}
	taskReady, ok := state.NodeState("task-ready")
	if !ok {
		t.Fatal("expected task-ready node state")
	}
	if taskReady.DependenciesComplete || taskReady.ExecutableReady {
		t.Fatalf("ready node with incomplete dependency must not be executable-ready: %#v", taskReady)
	}
	next, ok := state.NextReadyNode()
	if !ok || next.ID != "task-ready-2" {
		t.Fatalf("expected next dependency-ready node task-ready-2, got ok=%t node=%#v", ok, next)
	}
}

func TestWorkgraphStateCompletesExactlyOneDependencyReadyNode(t *testing.T) {
	workgraph := fixtureWorkgraph()
	link := RunLink{
		ContractVersion: RunLinkContract,
		TaskID:          "factory-task",
		Status:          "completed",
		Evidence:        map[string]string{"pr": "https://github.com/uesugitorachiyo/ao-atlas/pull/121"},
		Digest:          "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}

	state, err := BuildWorkgraphState(workgraph)
	if err != nil {
		t.Fatal(err)
	}
	completed, nodeID, err := state.CompleteWithRunLink(link)
	if err != nil {
		t.Fatal(err)
	}

	if nodeID != "task-ready" {
		t.Fatalf("expected first dependency-ready matching node to complete, got %s", nodeID)
	}
	nodes := workgraphNodesByID(completed)
	if nodes["task-ready"].Status != "completed" {
		t.Fatalf("matching ready node must be completed: %#v", nodes["task-ready"])
	}
	if nodes["done"].Status != "completed" || nodes["task-ready-2"].Status != "ready" {
		t.Fatalf("completion must affect exactly one node, got %#v", nodes)
	}
}

func TestBlueprintRequestFixtureIsValidAndPublicSafe(t *testing.T) {
	request, err := LoadJSON[BlueprintRequest](filepath.Join("..", "..", "examples", "valid", "blueprint-request.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateBlueprintRequest(request); err != nil {
		t.Fatal(err)
	}
	if request.Status != "blueprint_required" {
		t.Fatalf("expected blueprint_required, got %s", request.Status)
	}
}

func TestContextPackRejectsAbsoluteLocalPath(t *testing.T) {
	pack := ContextPack{
		ContractVersion: ContextPackContract,
		ID:              "bad-pack",
		TaskID:          "task",
		BudgetBytes:     4096,
		SourceRefs:      []SourceRef{{Ref: "/" + "absolute/context.md", Digest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}},
		Summaries:       []string{"summary"},
		MissingProtocol: "ask Blueprint for missing context",
	}
	if err := ValidateContextPack(pack, 0); err == nil || !strings.Contains(err.Error(), "absolute local path") {
		t.Fatalf("expected absolute path rejection, got %v", err)
	}
}

func TestContextPackRepackWritesBoundedPackForNeedsContext(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "context-pack.json")
	var out bytes.Buffer
	code := Run([]string{
		"context-pack", "repack",
		"--task", filepath.Join("..", "..", "examples", "valid", "factory-task.json"),
		"--run-link", filepath.Join("..", "..", "examples", "valid", "run-link-needs-context.json"),
		"--source-ref", "docs/sdd/AO-ATLAS-CONTEXT-PACKS.md",
		"--source-digest", "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"--budget", "4096",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("repack failed: %s", out.String())
	}
	pack, err := LoadJSON[ContextPack](outPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateContextPack(pack, 0); err != nil {
		t.Fatal(err)
	}
	if pack.TaskID != "atlas-readiness-task" || len(pack.SourceRefs) != 1 {
		t.Fatalf("unexpected context pack: %#v", pack)
	}
	if pack.MissingContextReason == "" || !strings.Contains(pack.MissingContextReason, "needs_context") {
		t.Fatalf("repacked context pack must include missing context reason: %#v", pack)
	}
	if len(pack.Assumptions) == 0 || len(pack.Exclusions) == 0 {
		t.Fatalf("repacked context pack must include assumptions and exclusions: %#v", pack)
	}
}

func TestContextPackRepackDemoFixtureValidates(t *testing.T) {
	pack, err := LoadJSON[ContextPack](filepath.Join("..", "..", "examples", "valid", "context-pack-needs-context-repack-demo.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateContextPack(pack, 0); err != nil {
		t.Fatal(err)
	}
	if pack.MissingContextReason == "" || !strings.Contains(pack.MissingContextReason, "needs_context") {
		t.Fatalf("demo fixture must include missing context reason: %#v", pack)
	}
}

func TestContextPackRepackRejectsCompletedRunLink(t *testing.T) {
	assertContextPackRepackFails(t, filepath.Join("..", "..", "examples", "valid", "run-link.json"), "blocked or failed")
}

func TestContextPackRepackRejectsBlockedRunLinkWithoutNeedsContext(t *testing.T) {
	assertContextPackRepackFails(t, filepath.Join("..", "..", "examples", "invalid", "run-link-blocked.json"), "needs_context")
}

func assertContextPackRepackFails(t *testing.T, runLinkPath, want string) {
	t.Helper()
	dir := t.TempDir()
	var out bytes.Buffer
	code := Run([]string{
		"context-pack", "repack",
		"--task", filepath.Join("..", "..", "examples", "valid", "factory-task.json"),
		"--run-link", runLinkPath,
		"--source-ref", "docs/sdd/AO-ATLAS-CONTEXT-PACKS.md",
		"--source-digest", "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"--budget", "4096",
		"--out", filepath.Join(dir, "context-pack.json"),
	}, &out, &out)
	if code == 0 {
		t.Fatalf("expected context-pack repack to fail for %s", runLinkPath)
	}
	if !strings.Contains(out.String(), want) {
		t.Fatalf("expected error containing %q, got %s", want, out.String())
	}
}

func TestWorkgraphNextSkipsBlockedAndDependencyIncomplete(t *testing.T) {
	wg := fixtureWorkgraph()
	node, ok := NextReadyNode(wg)
	if !ok {
		t.Fatal("expected ready node")
	}
	if node.ID != "task-ready" {
		t.Fatalf("expected task-ready, got %s", node.ID)
	}
}

func TestFoundryHandoffUsesReadyTasksOnly(t *testing.T) {
	wg := fixtureWorkgraph()
	handoff := BuildFoundryHandoff(wg)
	if err := ValidateFoundryHandoff(handoff); err != nil {
		t.Fatal(err)
	}
	if len(handoff.Tasks) != 2 {
		t.Fatalf("expected two ready tasks, got %d", len(handoff.Tasks))
	}
}

func TestFoundryImportWritesTaskFixturesForReadyNodes(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "foundry-import")
	var out bytes.Buffer
	code := Run([]string{
		"foundry", "import",
		"--workgraph", filepath.Join("..", "..", "examples", "valid", "workgraph.json"),
		"--instance", filepath.Join("..", "..", "examples", "valid", "stack-instance.json"),
		"--out", outDir,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("foundry import failed: %s", out.String())
	}
	manifest, err := LoadJSON[FoundryImport](filepath.Join(outDir, "foundry-import.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateFoundryImport(manifest); err != nil {
		t.Fatal(err)
	}
	if len(manifest.Tasks) != 1 {
		t.Fatalf("expected one dependency-ready task fixture, got %#v", manifest.Tasks)
	}
	if manifest.Tasks[0].TaskID != "atlas-readiness-task" {
		t.Fatalf("unexpected task fixture: %#v", manifest.Tasks[0])
	}
	if len(manifest.SourceArtifacts) != 2 {
		t.Fatalf("expected workgraph and instance source artifacts, got %#v", manifest.SourceArtifacts)
	}
	if len(manifest.Tasks[0].Task.ContextPackRefs) == 0 {
		t.Fatalf("expected context pack refs to be preserved: %#v", manifest.Tasks[0].Task)
	}
	if manifest.Tasks[0].MutationClass != "docs_only_single_file" {
		t.Fatalf("expected mutation class metadata, got %#v", manifest.Tasks[0])
	}
	if !containsString(manifest.Tasks[0].RequiredGates, "atlas_classification") {
		t.Fatalf("expected required gates metadata, got %#v", manifest.Tasks[0].RequiredGates)
	}
	if len(manifest.Tasks[0].RollbackScope) == 0 || manifest.Tasks[0].AuthorityBoundary == "" {
		t.Fatalf("expected rollback scope and authority boundary metadata, got %#v", manifest.Tasks[0])
	}
	if _, err := os.Stat(filepath.Join(outDir, manifest.Tasks[0].Path)); err != nil {
		t.Fatal(err)
	}
	if manifest.SchedulesWork || manifest.ExecutesWork || manifest.ApprovesWork {
		t.Fatalf("foundry import must be fixture-only readback: %#v", manifest)
	}
}

func TestFoundryImportWritesContinuationHandoffPrompt(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "foundry-import")
	workgraphPath := filepath.Join("..", "..", "examples", "valid", "workgraph.json")
	instancePath := filepath.Join("..", "..", "examples", "valid", "stack-instance.json")
	blueprintPackPath := filepath.Join("..", "ao-blueprint", "excluded", "fully_unsupervised_complex_mutation-readiness-blueprint")
	atlasImportPath := filepath.Join(".atlas-local", "fully-unsupervised-readiness", "blueprint-import", "blueprint-import.json")
	missionEvidencePath := filepath.Join(".atlas-local", "fully-unsupervised-readiness", "atlas-first-phase", "mission-continuation-evidence.json")
	targetFolder := strings.Join([]string{"", "Users", "torachiyouesugi", "Documents", "public", "ao-foundry"}, "/")
	var out bytes.Buffer
	code := Run([]string{
		"foundry", "import",
		"--workgraph", workgraphPath,
		"--instance", instancePath,
		"--out", outDir,
		"--blueprint-pack", blueprintPackPath,
		"--atlas-import", atlasImportPath,
		"--mission-continuation", missionEvidencePath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("foundry import failed: %s", out.String())
	}
	handoffPath := filepath.Join(outDir, "foundry-continuation-handoff.json")
	promptPath := filepath.Join(outDir, "foundry-continuation-prompt.md")
	handoff, err := LoadJSON[FoundryContinuationHandoff](handoffPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateFoundryContinuationHandoff(handoff); err != nil {
		t.Fatal(err)
	}
	if handoff.TargetFolder != targetFolder {
		t.Fatalf("unexpected target folder: %q", handoff.TargetFolder)
	}
	if handoff.Command != "codex --yolo" {
		t.Fatalf("unexpected command: %q", handoff.Command)
	}
	if handoff.FoundryImportPath != filepath.ToSlash(filepath.Join(outDir, "foundry-import.json")) {
		t.Fatalf("unexpected foundry import path: %q", handoff.FoundryImportPath)
	}
	if handoff.BlueprintPackPath != filepath.ToSlash(blueprintPackPath) ||
		handoff.AtlasImportPath != filepath.ToSlash(atlasImportPath) ||
		handoff.WorkgraphPath != filepath.ToSlash(workgraphPath) ||
		handoff.MissionContinuationEvidencePath != filepath.ToSlash(missionEvidencePath) {
		t.Fatalf("handoff did not preserve source artifact paths: %#v", handoff)
	}
	if handoff.FirstSafeNode != "readiness-ready" || handoff.TotalNodeCount != 2 {
		t.Fatalf("unexpected node readback: %#v", handoff)
	}
	if handoff.ReadyNodeCount != 1 || handoff.CompletedNodeCount != 1 || handoff.BlockedNodeCount != 0 {
		t.Fatalf("unexpected node counts: %#v", handoff)
	}
	for _, want := range []string{
		"Move to AO Foundry",
		"Run codex --yolo",
		"Paste this prompt",
		"do not stop after import validation",
		"do not stop after one gate artifact",
		"do not stop after one node",
		"Continue until all generated slices/tasks/nodes are consumed or a true hard blocker remains",
		"Atlas must not execute live mutation",
		"fully_unsupervised_complex_mutation remains denied",
		"RSI remains denied",
		filepath.ToSlash(blueprintPackPath),
		filepath.ToSlash(workgraphPath),
		filepath.ToSlash(filepath.Join(outDir, "foundry-import.json")),
		"Stop only on done, final denial, hard blocker, CI failure, unsafe scope drift, or kill switch.",
	} {
		if !strings.Contains(handoff.Prompt, want) {
			t.Fatalf("handoff prompt missing %q:\n%s", want, handoff.Prompt)
		}
	}
	promptBytes, err := os.ReadFile(promptPath)
	if err != nil {
		t.Fatal(err)
	}
	prompt := string(promptBytes)
	for _, want := range []string{"Move to AO Foundry", "Run codex --yolo", "Paste this prompt"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt artifact missing %q:\n%s", want, prompt)
		}
	}
	if strings.Contains(prompt, "cat "+filepath.Join(outDir, "foundry-import.json")) ||
		strings.Contains(handoff.NextRecommendedAction, "cat ") {
		t.Fatalf("continuation handoff must not use cat as the primary next action: %#v\n%s", handoff, prompt)
	}
	if !strings.Contains(out.String(), "foundry_continuation_prompt="+promptPath) ||
		!strings.Contains(out.String(), "Run codex --yolo") ||
		!strings.Contains(out.String(), "Paste this prompt") ||
		strings.Contains(out.String(), "cat ") ||
		!strings.Contains(out.String(), "Move to "+targetFolder) {
		t.Fatalf("CLI output did not report operator-ready continuation action:\n%s", out.String())
	}
}

func TestFoundryContinuationHandoffRejectsBlockedImportNode(t *testing.T) {
	workgraph := fixtureWorkgraph()
	foundryImport, err := BuildFoundryImportForNodes(workgraph, []string{"task-ready"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	foundryImport.Tasks[0].NodeID = "task-blocked"

	_, err = BuildFoundryContinuationHandoff(workgraph, foundryImport, FoundryContinuationHandoffInputs{})
	if err == nil || !strings.Contains(err.Error(), "foundry import node_id must reference a ready workgraph node") {
		t.Fatalf("expected blocked import node rejection, got %v", err)
	}
}

func TestFoundryContinuationPromptImplementationLivesInDedicatedModule(t *testing.T) {
	module, err := os.ReadFile("foundry_handoff.go")
	if err != nil {
		t.Fatalf("expected dedicated Foundry handoff module: %v", err)
	}
	content := string(module)
	for _, want := range []string{
		"func BuildFoundryHandoff(",
		"func BuildFoundryContinuationHandoff(",
		"func WriteFoundryContinuationPrompt(",
		"func buildFoundryContinuationPrompt(",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("dedicated handoff module missing %q", want)
		}
	}
}

func TestFoundryContinuationHandoffRejectsImportTaskMismatch(t *testing.T) {
	workgraph := fixtureWorkgraph()
	foundryImport, err := BuildFoundryImportForNodes(workgraph, []string{"task-ready"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	workgraph.Nodes[1].FactoryTask.ID = "renamed-ready-task"

	_, err = BuildFoundryContinuationHandoff(workgraph, foundryImport, FoundryContinuationHandoffInputs{})
	if err == nil || !strings.Contains(err.Error(), "foundry import task_id must match ready workgraph node task") {
		t.Fatalf("expected import task mismatch rejection, got %v", err)
	}
}

func TestLargeWorkgraphStressFixtureValidatesAndImportsReadyNodes(t *testing.T) {
	workgraph, err := LoadJSON[Workgraph](filepath.Join("..", "..", "examples", "valid", "workgraph-large-stress.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateWorkgraph(workgraph); err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{"ready": 0, "blocked": 0, "completed": 0}
	contextRefs := 0
	for _, node := range workgraph.Nodes {
		counts[node.Status]++
		contextRefs += len(node.FactoryTask.ContextPackRefs)
	}
	if len(workgraph.Nodes) != 12 || counts["completed"] != 4 || counts["ready"] != 5 || counts["blocked"] != 3 {
		t.Fatalf("unexpected stress fixture counts: nodes=%d counts=%#v", len(workgraph.Nodes), counts)
	}
	if contextRefs < 8 {
		t.Fatalf("expected stress fixture to exercise context-pack refs, got %d", contextRefs)
	}
	importManifest, err := BuildFoundryImport(workgraph)
	if err != nil {
		t.Fatal(err)
	}
	if len(importManifest.Tasks) != 5 {
		t.Fatalf("expected five dependency-ready import tasks, got %#v", importManifest.Tasks)
	}
	if importManifest.SchedulesWork || importManifest.ExecutesWork || importManifest.ApprovesWork {
		t.Fatalf("stress import must remain authority-free: %#v", importManifest)
	}
}

func TestComplexRepoMutationRehearsalFixtureModelsSafeDryRunLadder(t *testing.T) {
	workgraph, err := LoadJSON[Workgraph](filepath.Join("..", "..", "examples", "valid", "workgraph-complex-repo-mutation-rehearsal.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateWorkgraph(workgraph); err != nil {
		t.Fatal(err)
	}
	if len(workgraph.Nodes) != 14 {
		t.Fatalf("complex rehearsal should model 14 governed nodes, got %d", len(workgraph.Nodes))
	}
	nodes := map[string]WorkgraphNode{}
	for _, node := range workgraph.Nodes {
		nodes[node.ID] = node
		if node.FactoryTask.MutationClass != "complex_repo_mutation" {
			t.Fatalf("complex rehearsal node %s must stay complex_repo_mutation, got %q", node.ID, node.FactoryTask.MutationClass)
		}
		if node.FactoryTask.AuthorityBoundary != "atlas_classification_only" {
			t.Fatalf("complex rehearsal node %s must remain classification-only: %#v", node.ID, node.FactoryTask)
		}
		if !containsString(node.FactoryTask.SafetyLimits, "do_not_advance:complex_repo_mutation_live_execution_denied") &&
			node.ID != "complex-dependency-gate-blocked" &&
			node.ID != "complex-sentinel-hold-gate-blocked" &&
			node.ID != "complex-promoter-promotion-gate-blocked" &&
			node.ID != "complex-command-readback-blocked" &&
			node.ID != "complex-ci-gate-blocked" {
			t.Fatalf("complex rehearsal node %s must preserve live-execution denial safety: %#v", node.ID, node.FactoryTask.SafetyLimits)
		}
	}
	decomposition := nodes["complex-low-risk-decomposition-ready"]
	if decomposition.Status != "ready" ||
		!containsString(decomposition.FactoryTask.RequiredGates, "low_risk_decomposition") ||
		!containsString(decomposition.FactoryTask.RequiredEvidence, "low_risk_decomposition:complex_repo_mutation") {
		t.Fatalf("complex rehearsal must include low-risk decomposition node: %#v", decomposition)
	}
	rollbackGraph := nodes["complex-rollback-graph-blocked"]
	if rollbackGraph.Status != "blocked" ||
		!containsString(rollbackGraph.FactoryTask.RequiredGates, "rollback_graph") ||
		!containsString(rollbackGraph.FactoryTask.RequiredEvidence, "rollback_graph:complex_repo_mutation") {
		t.Fatalf("complex rehearsal must include rollback graph node: %#v", rollbackGraph)
	}
	if !containsString(nodes["complex-dependency-gate-blocked"].Dependencies, "complex-low-risk-decomposition-ready") {
		t.Fatalf("dependency gate must wait on low-risk decomposition: %#v", nodes["complex-dependency-gate-blocked"].Dependencies)
	}
	if !containsString(nodes["complex-repair-plan-blocked"].Dependencies, "complex-rollback-graph-blocked") {
		t.Fatalf("repair plan must wait on rollback graph: %#v", nodes["complex-repair-plan-blocked"].Dependencies)
	}
	next, ok := NextReadyNode(workgraph)
	if !ok || next.ID != "complex-intake-ready" {
		t.Fatalf("complex rehearsal should expose only the first safe ready node, got ok=%t node=%#v", ok, next)
	}
	importManifest, err := BuildFoundryImportForNodes(workgraph, []string{"complex-intake-ready"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(importManifest.Tasks) != 1 || importManifest.Tasks[0].NodeID != "complex-intake-ready" {
		t.Fatalf("complex rehearsal import must select one dependency-safe node, got %#v", importManifest.Tasks)
	}
	if importManifest.SchedulesWork || importManifest.ExecutesWork || importManifest.ApprovesWork {
		t.Fatalf("complex rehearsal import must remain readback-only: %#v", importManifest)
	}
}

func TestFoundryImportJSONSelectsSingleReadyNode(t *testing.T) {
	var out bytes.Buffer
	code := Run([]string{
		"foundry", "import",
		"--workgraph", filepath.Join("..", "..", "examples", "valid", "workgraph.json"),
		"--instance", filepath.Join("..", "..", "examples", "valid", "stack-instance.json"),
		"--node", "readiness-ready",
		"--json",
	}, &out, &out)
	if code != 0 {
		t.Fatalf("foundry import json failed: %s", out.String())
	}
	var manifest FoundryImport
	if err := json.Unmarshal(out.Bytes(), &manifest); err != nil {
		t.Fatalf("foundry import did not emit json: %v\n%s", err, out.String())
	}
	if len(manifest.Tasks) != 1 || manifest.Tasks[0].NodeID != "readiness-ready" {
		t.Fatalf("expected selected ready node only, got %#v", manifest.Tasks)
	}
	if manifest.SchedulesWork || manifest.ExecutesWork || manifest.ApprovesWork {
		t.Fatalf("foundry import must not claim authority: %#v", manifest)
	}
}

func TestFoundryImportRejectsNoReadyNodes(t *testing.T) {
	wg := fixtureWorkgraph()
	for i := range wg.Nodes {
		if wg.Nodes[i].Status == "ready" {
			wg.Nodes[i].Status = "blocked"
			wg.Nodes[i].Blockers = []string{"not ready"}
		}
	}
	_, err := BuildFoundryImport(wg)
	if err == nil || !strings.Contains(err.Error(), "tasks must not be empty") {
		t.Fatalf("expected no ready tasks rejection, got %v", err)
	}
}

func TestFoundryImportRejectsBlockedSelectedNode(t *testing.T) {
	dir := t.TempDir()
	var out bytes.Buffer
	code := Run([]string{
		"foundry", "import",
		"--workgraph", filepath.Join("..", "..", "examples", "valid", "workgraph.json"),
		"--instance", filepath.Join("..", "..", "examples", "valid", "stack-instance.json"),
		"--node", "contracts-ready",
		"--out", filepath.Join(dir, "foundry-import"),
	}, &out, &out)
	if code == 0 {
		t.Fatal("expected completed selected node to fail")
	}
	if !strings.Contains(out.String(), "selected node") {
		t.Fatalf("expected selected node error, got %s", out.String())
	}
}

func TestFoundryImportRejectsMissingContextPack(t *testing.T) {
	workgraph := fixtureWorkgraph()
	workgraph.TargetInstance = "demo-stack"
	workgraph.Nodes = []WorkgraphNode{{
		ID:           "task-ready",
		Status:       "ready",
		Dependencies: []string{},
		FactoryTask: FactoryTask{
			ContractVersion:   FactoryTaskContract,
			ID:                "missing-context-task",
			Objective:         "Import should fail without referenced context pack.",
			TargetFactoryRepo: "ao-foundry",
			FactoryFolder:     "factory/missing-context",
			MutationClass:     "docs_only_single_file",
			Acceptance:        []string{"context pack exists"},
			NonGoals:          []string{"do not execute"},
			WriteScope:        []string{"factory/missing-context"},
			RequiredGates:     []string{"atlas_classification"},
			RollbackScope:     []string{"factory/missing-context"},
			Verification:      []string{"go test ./..."},
			RequiredEvidence:  []string{"summary.json"},
			SafetyLimits:      []string{"public-safe only"},
			AuthorityBoundary: "atlas_classification_only",
			ContextPackRefs:   []string{"examples/valid/not-present.context-pack.json"},
		},
	}}
	workgraphPath := filepath.Join("..", "..", "target", "test-workgraph-missing-context.json")
	if err := WriteJSON(workgraphPath, workgraph); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	code := Run([]string{
		"foundry", "import",
		"--workgraph", workgraphPath,
		"--instance", filepath.Join("..", "..", "examples", "valid", "stack-instance.json"),
		"--out", filepath.Join("..", "..", "target", "test-foundry-import-missing-context"),
	}, &out, &out)
	if code == 0 {
		t.Fatal("expected missing context pack to fail")
	}
	if !strings.Contains(out.String(), "context pack") {
		t.Fatalf("expected context pack error, got %s", out.String())
	}
}

func TestFoundryImportRejectsReadyNodeMissingMutationClass(t *testing.T) {
	workgraph, err := LoadJSON[Workgraph](filepath.Join("..", "..", "examples", "invalid", "workgraph-foundry-import-missing-mutation-class.json"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = BuildFoundryImport(workgraph)
	if err == nil || !strings.Contains(err.Error(), "mutation_class") {
		t.Fatalf("expected missing mutation_class rejection, got %v", err)
	}
}

func TestFoundryImportRejectsReadyNodeMissingRequiredGates(t *testing.T) {
	workgraph, err := LoadJSON[Workgraph](filepath.Join("..", "..", "examples", "invalid", "workgraph-foundry-import-missing-required-gates.json"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = BuildFoundryImport(workgraph)
	if err == nil || !strings.Contains(err.Error(), "required_gates") {
		t.Fatalf("expected missing required_gates rejection, got %v", err)
	}
}

func TestComplexExecutableNodeEvidenceRejectsCandidateSafeToExecuteFalse(t *testing.T) {
	workgraph := complexNodeEvidenceWorkgraph("complex-docs-intake")
	foundryImport := complexNodeEvidenceImport("complex-docs-intake", "safe_to_execute:true")
	candidate := complexNodeEvidenceCandidate("complex-docs-intake", false)
	rollback := complexNodeEvidenceRollback("complex-docs-intake", true)
	summary := map[string]any{"first_safe_node": "complex-docs-intake"}

	err := ValidateComplexExecutableNodeEvidence(workgraph, foundryImport, candidate, rollback, summary)
	if err == nil || !strings.Contains(err.Error(), "candidate record safe_to_execute must be true") {
		t.Fatalf("expected candidate safe_to_execute rejection, got %v", err)
	}
}

func TestComplexExecutableNodeEvidenceRejectsRollbackSafeToExecuteFalse(t *testing.T) {
	workgraph := complexNodeEvidenceWorkgraph("complex-docs-intake")
	foundryImport := complexNodeEvidenceImport("complex-docs-intake", "safe_to_execute:true")
	candidate := complexNodeEvidenceCandidate("complex-docs-intake", true)
	rollback := complexNodeEvidenceRollback("complex-docs-intake", false)
	summary := map[string]any{"first_safe_node": "complex-docs-intake"}

	err := ValidateComplexExecutableNodeEvidence(workgraph, foundryImport, candidate, rollback, summary)
	if err == nil || !strings.Contains(err.Error(), "rollback record safe_to_execute must be true") {
		t.Fatalf("expected rollback safe_to_execute rejection, got %v", err)
	}
}

func TestComplexExecutableNodeEvidenceRejectsReadyNodeImportMismatch(t *testing.T) {
	workgraph := complexNodeEvidenceWorkgraph("complex-docs-intake")
	foundryImport := complexNodeEvidenceImport("complex-next-node", "safe_to_execute:true")
	candidate := complexNodeEvidenceCandidate("complex-docs-intake", true)
	rollback := complexNodeEvidenceRollback("complex-docs-intake", true)
	summary := map[string]any{"first_safe_node": "complex-docs-intake"}

	err := ValidateComplexExecutableNodeEvidence(workgraph, foundryImport, candidate, rollback, summary)
	if err == nil || !strings.Contains(err.Error(), "foundry import node_id must match ready node") {
		t.Fatalf("expected import mismatch rejection, got %v", err)
	}
}

func TestComplexExecutableNodeEvidenceRejectsConcurrentExecutableNodes(t *testing.T) {
	workgraph := complexNodeEvidenceWorkgraph("complex-docs-intake", "complex-next-node")
	foundryImport := complexNodeEvidenceImport("complex-docs-intake", "safe_to_execute:true")
	candidate := complexNodeEvidenceCandidate("complex-docs-intake", true)
	rollback := complexNodeEvidenceRollback("complex-docs-intake", true)
	summary := map[string]any{"first_safe_node": "complex-docs-intake"}

	err := ValidateComplexExecutableNodeEvidence(workgraph, foundryImport, candidate, rollback, summary)
	if err == nil || !strings.Contains(err.Error(), "exactly one ready executable node is allowed") {
		t.Fatalf("expected concurrent executable node rejection, got %v", err)
	}
}

func TestComplexExecutableNodeEvidenceAcceptsOneBoundSafeNode(t *testing.T) {
	workgraph := complexNodeEvidenceWorkgraph("complex-docs-intake")
	foundryImport := complexNodeEvidenceImport("complex-docs-intake", "safe_to_execute:true")
	candidate := complexNodeEvidenceCandidate("complex-docs-intake", true)
	rollback := complexNodeEvidenceRollback("complex-docs-intake", true)
	summary := map[string]any{"first_safe_node": "complex-docs-intake"}

	if err := ValidateComplexExecutableNodeEvidence(workgraph, foundryImport, candidate, rollback, summary); err != nil {
		t.Fatalf("expected safe node evidence to validate: %v", err)
	}
}

func TestComplexMissionContinuationRepairsNextBlockedNodeAfterCompletedDependency(t *testing.T) {
	workgraph := complexMissionContinuationWorkgraph()
	runLink := RunLink{
		ContractVersion: RunLinkContract,
		TaskID:          "complex-docs-intake-task",
		Status:          "completed",
		Evidence:        map[string]string{"pr": "https://github.com/uesugitorachiyo/ao-atlas/pull/41", "ci": "passed"},
		Digest:          "sha256:f6fe6e812f8793c5cb352b43bd24c15fa4bdead147e48d6b3ff3191b16941b22",
	}

	continued, active, ok, err := RepairNextComplexMissionNode(workgraph, runLink)
	if err != nil {
		t.Fatal(err)
	}
	if !ok || active.ID != "complex-test-scope" {
		t.Fatalf("expected complex-test-scope to become active, got ok=%t node=%#v", ok, active)
	}
	nodes := workgraphNodesByID(continued)
	if nodes["complex-docs-intake"].Status != "completed" {
		t.Fatalf("completed run-link must complete docs node: %#v", nodes["complex-docs-intake"])
	}
	testScope := nodes["complex-test-scope"]
	if testScope.Status != "ready" || len(testScope.Blockers) != 0 {
		t.Fatalf("repairable test-scope node must be ready with no blockers: %#v", testScope)
	}
	if !containsString(testScope.FactoryTask.RequiredEvidence, "safe_to_execute:true") ||
		containsString(testScope.FactoryTask.RequiredEvidence, "safe_to_execute:false") {
		t.Fatalf("repairable node must bind safe_to_execute:true only: %#v", testScope.FactoryTask.RequiredEvidence)
	}
	if ready := readyNodeIDs(continued); len(ready) != 1 || ready[0] != "complex-test-scope" {
		t.Fatalf("continuation must expose exactly one ready node, got %#v", ready)
	}
}

func TestFoundryImportRejectsSameInputAndOutputPath(t *testing.T) {
	var out bytes.Buffer
	path := filepath.Join("..", "..", "examples", "valid", "workgraph.json")
	code := Run([]string{"foundry", "import", "--workgraph", path, "--instance", filepath.Join("..", "..", "examples", "valid", "stack-instance.json"), "--out", path}, &out, &out)
	if code == 0 {
		t.Fatal("expected same input/output path to fail")
	}
	if !strings.Contains(out.String(), "overwrite input") {
		t.Fatalf("expected overwrite error, got %s", out.String())
	}
}

func TestFoundryImportFixtureIsValid(t *testing.T) {
	foundryImport, err := LoadJSON[FoundryImport](filepath.Join("..", "..", "examples", "valid", "foundry-import.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateFoundryImport(foundryImport); err != nil {
		t.Fatal(err)
	}
}

func TestFoundryImportRejectsExecutionAuthority(t *testing.T) {
	foundryImport, err := LoadJSON[FoundryImport](filepath.Join("..", "..", "examples", "invalid", "foundry-import-executes-work.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateFoundryImport(foundryImport); err == nil || !strings.Contains(err.Error(), "executes_work") {
		t.Fatalf("expected executes_work rejection, got %v", err)
	}
}

func TestMutationClassModelFixtureDefinesAllAuthorityFreeClasses(t *testing.T) {
	model, err := LoadJSON[MutationClassModel](filepath.Join("..", "..", "examples", "valid", "mutation-classes.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateMutationClassModel(model); err != nil {
		t.Fatal(err)
	}
	if model.SchedulesWork || model.ExecutesWork || model.ApprovesWork {
		t.Fatalf("mutation class model must classify only, got authority flags: %#v", model)
	}
	want := []string{
		"docs_only_single_file",
		"docs_only_multi_file",
		"docs_config_only",
		"test_only",
		"low_risk_code",
		"multi_repo_low_risk",
		"complex_repo_mutation",
	}
	if len(model.Classes) != len(want) {
		t.Fatalf("expected %d mutation classes, got %d", len(want), len(model.Classes))
	}
	for _, name := range want {
		class, ok := mutationClassByName(model, name)
		if !ok {
			t.Fatalf("missing mutation class %s", name)
		}
		if len(class.AllowedPaths) == 0 || len(class.ForbiddenPaths) == 0 {
			t.Fatalf("%s must define allowed and forbidden paths: %#v", name, class)
		}
		if class.MaxFiles <= 0 {
			t.Fatalf("%s must define positive max_files: %#v", name, class)
		}
		if len(class.RequiredGates) == 0 || len(class.RollbackRequirements) == 0 || len(class.CIRequirements) == 0 || len(class.PromotionRequirements) == 0 {
			t.Fatalf("%s must define gates, rollback, CI, and promotion requirements: %#v", name, class)
		}
	}
	docsMulti, _ := mutationClassByName(model, "docs_only_multi_file")
	if docsMulti.MaxFiles != 2 {
		t.Fatalf("docs_only_multi_file should remain bounded to two files until proven live, got %d", docsMulti.MaxFiles)
	}
	testOnly, _ := mutationClassByName(model, "test_only")
	if testOnly.MaxFiles != 1 {
		t.Fatalf("test_only should remain bounded to one file until proven live, got %d", testOnly.MaxFiles)
	}
	for _, want := range []string{"**/*_test.go", "tests/**", "testdata/**"} {
		if !containsString(testOnly.AllowedPaths, want) {
			t.Fatalf("test_only allowed paths missing %s: %#v", want, testOnly.AllowedPaths)
		}
	}
	if !containsString(testOnly.RequiredGates, "sentinel_coverage_no_hold") {
		t.Fatalf("test_only must require Sentinel coverage no-hold: %#v", testOnly.RequiredGates)
	}
	complex, _ := mutationClassByName(model, "complex_repo_mutation")
	if !containsString(complex.RequiredGates, "all_lower_classes_live_rehearsed") {
		t.Fatalf("complex mutation must be denied until all lower live rehearsals are proven: %#v", complex.RequiredGates)
	}
}

func TestMutationClassModelRejectsAuthorityClaims(t *testing.T) {
	model, err := LoadJSON[MutationClassModel](filepath.Join("..", "..", "examples", "invalid", "mutation-classes-claims-authority.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateMutationClassModel(model); err == nil || !strings.Contains(err.Error(), "executes_work") {
		t.Fatalf("expected executes_work authority rejection, got %v", err)
	}
}

func TestMutationClassModelRejectsMissingRollback(t *testing.T) {
	model, err := LoadJSON[MutationClassModel](filepath.Join("..", "..", "examples", "invalid", "mutation-classes-missing-rollback.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateMutationClassModel(model); err == nil || !strings.Contains(err.Error(), "rollback_requirements") {
		t.Fatalf("expected rollback requirement rejection, got %v", err)
	}
}

func TestMutationClassModelRejectsMissingRequiredClass(t *testing.T) {
	model, err := LoadJSON[MutationClassModel](filepath.Join("..", "..", "examples", "invalid", "mutation-classes-missing-class.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateMutationClassModel(model); err == nil || !strings.Contains(err.Error(), "complex_repo_mutation") {
		t.Fatalf("expected missing class rejection, got %v", err)
	}
}

func TestMutationClassModelValidateCommand(t *testing.T) {
	var out bytes.Buffer
	code := Run([]string{"mutation-classes", "validate", "--model", filepath.Join("..", "..", "examples", "valid", "mutation-classes.json")}, &out, &out)
	if code != 0 {
		t.Fatalf("mutation-classes validate failed: %s", out.String())
	}
	if !strings.Contains(out.String(), "status=valid") {
		t.Fatalf("expected valid status, got %s", out.String())
	}
}

func TestAuthorityLadderWorkgraphFixtureModelsGovernedEscalation(t *testing.T) {
	workgraph, err := LoadJSON[Workgraph](filepath.Join("..", "..", "examples", "valid", "workgraph-authority-ladder.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateWorkgraph(workgraph); err != nil {
		t.Fatal(err)
	}
	if len(workgraph.Nodes) < 16 {
		t.Fatalf("authority ladder fixture must be large enough to model escalation, got %d nodes", len(workgraph.Nodes))
	}
	counts := map[string]int{"ready": 0, "blocked": 0, "completed": 0}
	for _, node := range workgraph.Nodes {
		counts[node.Status]++
	}
	if counts["completed"] < 1 || counts["ready"] < 2 || counts["blocked"] < 8 {
		t.Fatalf("authority ladder must include completed, ready, and blocked states, got %#v", counts)
	}
	for _, want := range []string{"repair", "repack", "sentinel", "promoter", "command"} {
		if !workgraphHasNodeContaining(workgraph, want) {
			t.Fatalf("authority ladder fixture missing %s node", want)
		}
	}
	if !workgraphHasSafetyLimitContaining(workgraph, "do_not_advance") {
		t.Fatalf("authority ladder fixture must encode do-not-advance gates")
	}
}

func TestMissionStatusReportsAuthorityLadderReadback(t *testing.T) {
	var out bytes.Buffer
	code := Run([]string{
		"mission", "status",
		"--intake", filepath.Join("..", "..", "examples", "valid", "intake-authority-ladder.json"),
		"--workgraph", filepath.Join("..", "..", "examples", "valid", "workgraph-authority-ladder.json"),
		"--run-link", filepath.Join("..", "..", "examples", "valid", "run-link-authority-ladder-docs-single.json"),
		"--json",
	}, &out, &out)
	if code != 0 {
		t.Fatalf("mission status authority ladder failed: %s", out.String())
	}
	var status MissionStatus
	if err := json.Unmarshal(out.Bytes(), &status); err != nil {
		t.Fatalf("mission status did not emit json: %v\n%s", err, out.String())
	}
	if err := ValidateMissionStatus(status); err != nil {
		t.Fatal(err)
	}
	if status.AuthorityLadder == nil {
		t.Fatalf("expected authority ladder readback, got %#v", status)
	}
	if status.AuthorityLadder.CurrentClass != "docs_only_single_file" {
		t.Fatalf("expected current class docs_only_single_file, got %#v", status.AuthorityLadder)
	}
	if status.AuthorityLadder.NextClass != "docs_only_multi_file" {
		t.Fatalf("expected next class docs_only_multi_file, got %#v", status.AuthorityLadder)
	}
	if !containsString(status.AuthorityLadder.ProvenLiveClasses, "docs_only_single_file") {
		t.Fatalf("expected docs-only single-file live evidence, got %#v", status.AuthorityLadder.ProvenLiveClasses)
	}
	if !containsString(status.AuthorityLadder.RequiredEvidence, "sentinel_no_hold:docs_only_multi_file") {
		t.Fatalf("expected Sentinel evidence requirement for docs_only_multi_file, got %#v", status.AuthorityLadder.RequiredEvidence)
	}
	if len(status.AuthorityLadder.Blockers) == 0 || !stringSliceContains(status.AuthorityLadder.Blockers, "docs_only_multi_file") {
		t.Fatalf("expected docs_only_multi_file blockers, got %#v", status.AuthorityLadder.Blockers)
	}
	if status.AuthorityLadder.DeniedHigherClasses["complex_repo_mutation"] == "" {
		t.Fatalf("expected complex mutation denial reason, got %#v", status.AuthorityLadder.DeniedHigherClasses)
	}
}

func TestMissionStatusReportsTestOnlyDryRunReadinessWithoutLiveExecution(t *testing.T) {
	var out bytes.Buffer
	code := Run([]string{
		"mission", "status",
		"--intake", filepath.Join("..", "..", "examples", "valid", "intake-authority-ladder.json"),
		"--workgraph", filepath.Join("..", "..", "examples", "valid", "workgraph-authority-ladder-test-only-dry-run.json"),
		"--run-link", filepath.Join("..", "..", "examples", "valid", "run-link-authority-ladder-docs-single.json"),
		"--run-link", filepath.Join("..", "..", "examples", "valid", "run-link-authority-ladder-docs-multi.json"),
		"--json",
	}, &out, &out)
	if code != 0 {
		t.Fatalf("mission status test-only dry run failed: %s", out.String())
	}
	var status MissionStatus
	if err := json.Unmarshal(out.Bytes(), &status); err != nil {
		t.Fatalf("mission status did not emit json: %v\n%s", err, out.String())
	}
	if err := ValidateMissionStatus(status); err != nil {
		t.Fatal(err)
	}
	if status.AuthorityLadder == nil {
		t.Fatalf("expected authority ladder readback, got %#v", status)
	}
	if status.AuthorityLadder.CurrentClass != "docs_only_multi_file" {
		t.Fatalf("expected current class docs_only_multi_file, got %#v", status.AuthorityLadder)
	}
	if status.AuthorityLadder.NextClass != "test_only" {
		t.Fatalf("expected next class test_only, got %#v", status.AuthorityLadder)
	}
	if !containsString(status.AuthorityLadder.DryRunReadyClasses, "test_only") {
		t.Fatalf("expected test_only dry-run readiness, got %#v", status.AuthorityLadder.DryRunReadyClasses)
	}
	for _, want := range []string{"dry_run:test_only", "rollback_plan:test_only", "ci_passed:test_only"} {
		if !containsString(status.AuthorityLadder.RequiredEvidence, want) {
			t.Fatalf("expected required evidence %s, got %#v", want, status.AuthorityLadder.RequiredEvidence)
		}
	}
	if status.SchedulesWork || status.ExecutesWork {
		t.Fatalf("test-only dry run readback must remain non-mutating: %#v", status)
	}
	if status.AuthorityLadder.DeniedHigherClasses["low_risk_code"] == "" ||
		status.AuthorityLadder.DeniedHigherClasses["complex_repo_mutation"] == "" {
		t.Fatalf("expected higher classes to remain denied, got %#v", status.AuthorityLadder.DeniedHigherClasses)
	}
}

func TestMissionStatusReportsLowRiskDryRunReadinessWithoutLiveExecution(t *testing.T) {
	var out bytes.Buffer
	code := Run([]string{
		"mission", "status",
		"--intake", filepath.Join("..", "..", "examples", "valid", "intake-authority-ladder.json"),
		"--workgraph", filepath.Join("..", "..", "examples", "valid", "workgraph-authority-ladder-low-risk-dry-run.json"),
		"--run-link", filepath.Join("..", "..", "examples", "valid", "run-link-authority-ladder-docs-single.json"),
		"--run-link", filepath.Join("..", "..", "examples", "valid", "run-link-authority-ladder-docs-multi.json"),
		"--json",
	}, &out, &out)
	if code != 0 {
		t.Fatalf("mission status low-risk dry run failed: %s", out.String())
	}
	var status MissionStatus
	if err := json.Unmarshal(out.Bytes(), &status); err != nil {
		t.Fatalf("mission status did not emit json: %v\n%s", err, out.String())
	}
	if err := ValidateMissionStatus(status); err != nil {
		t.Fatal(err)
	}
	if status.AuthorityLadder == nil {
		t.Fatalf("expected authority ladder readback, got %#v", status)
	}
	if status.AuthorityLadder.CurrentClass != "test_only" {
		t.Fatalf("expected current class test_only, got %#v", status.AuthorityLadder)
	}
	if status.AuthorityLadder.NextClass != "low_risk_code" {
		t.Fatalf("expected next class low_risk_code, got %#v", status.AuthorityLadder)
	}
	if !containsString(status.AuthorityLadder.ProvenLiveClasses, "test_only") {
		t.Fatalf("expected test_only live proof, got %#v", status.AuthorityLadder.ProvenLiveClasses)
	}
	if !containsString(status.AuthorityLadder.DryRunReadyClasses, "low_risk_code") {
		t.Fatalf("expected low_risk_code dry-run readiness, got %#v", status.AuthorityLadder.DryRunReadyClasses)
	}
	for _, want := range []string{"dry_run:low_risk_code", "rollback_plan:low_risk_code", "sentinel_no_hold:low_risk_code", "promoter_ready:low_risk_code", "command_readback:low_risk_code"} {
		if !containsString(status.AuthorityLadder.RequiredEvidence, want) {
			t.Fatalf("expected required evidence %s, got %#v", want, status.AuthorityLadder.RequiredEvidence)
		}
	}
	if status.SchedulesWork || status.ExecutesWork {
		t.Fatalf("low-risk dry-run readback must remain non-mutating: %#v", status)
	}
	if status.AuthorityLadder.DeniedHigherClasses["multi_repo_low_risk"] == "" ||
		status.AuthorityLadder.DeniedHigherClasses["complex_repo_mutation"] == "" {
		t.Fatalf("expected higher classes to remain denied, got %#v", status.AuthorityLadder.DeniedHigherClasses)
	}
	if status.AuthorityLadder.DeniedHigherClasses["low_risk_code"] != "" {
		t.Fatalf("low_risk_code should be next request class, not a higher denied class: %#v", status.AuthorityLadder.DeniedHigherClasses)
	}
}

func TestLowRiskCodeDenialAuditExplainsBlockedExecution(t *testing.T) {
	audit, err := LoadJSON[LowRiskCodeDenialAudit](filepath.Join("..", "..", "examples", "valid", "low-risk-code-denial-audit.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateLowRiskCodeDenialAudit(audit); err != nil {
		t.Fatal(err)
	}
	if audit.MutationClass != "low_risk_code" ||
		audit.CurrentProvenLiveClass != "test_only" ||
		audit.NextDeniedClass != "low_risk_code" ||
		audit.SafeToExecute {
		t.Fatalf("unexpected low-risk denial audit boundary: %#v", audit)
	}
	for _, want := range []string{
		"policy:low_risk_code_live_promotion",
		"rollback_proof:low_risk_code_live",
		"sentinel_clear:low_risk_code_live",
		"promoter_promotion:low_risk_code_live",
		"command_readback:low_risk_code_live",
		"ci_passed:low_risk_code_live",
	} {
		if !containsString(audit.MissingPolicyEvidence, want) &&
			!containsString(audit.MissingRollbackEvidence, want) &&
			!containsString(audit.MissingSentinelPromoterEvidence, want) &&
			!containsString(audit.CIRequirements, want) {
			t.Fatalf("low-risk denial audit missing %s: %#v", want, audit)
		}
	}
	if audit.SentinelState != "missing_live_no_hold" ||
		audit.PromoterState != "missing_live_promotion" ||
		audit.ExactNextAction != "build_low_risk_code_promotion_prerequisites" {
		t.Fatalf("low-risk denial audit states/next action drifted: %#v", audit)
	}
	if audit.SchedulesWork || audit.ExecutesWork || audit.ApprovesWork {
		t.Fatalf("Atlas denial audit must remain read-only: %#v", audit)
	}
}

func TestMissionStatusReportsMultiRepoLowRiskDryRunReadinessWithoutLiveExecution(t *testing.T) {
	var out bytes.Buffer
	code := Run([]string{
		"mission", "status",
		"--intake", filepath.Join("..", "..", "examples", "valid", "intake-authority-ladder.json"),
		"--workgraph", filepath.Join("..", "..", "examples", "valid", "workgraph-authority-ladder-multi-repo-dry-run.json"),
		"--run-link", filepath.Join("..", "..", "examples", "valid", "run-link-authority-ladder-docs-single.json"),
		"--run-link", filepath.Join("..", "..", "examples", "valid", "run-link-authority-ladder-docs-multi.json"),
		"--json",
	}, &out, &out)
	if code != 0 {
		t.Fatalf("mission status multi-repo dry run failed: %s", out.String())
	}
	var status MissionStatus
	if err := json.Unmarshal(out.Bytes(), &status); err != nil {
		t.Fatalf("mission status did not emit json: %v\n%s", err, out.String())
	}
	if err := ValidateMissionStatus(status); err != nil {
		t.Fatal(err)
	}
	if status.AuthorityLadder == nil {
		t.Fatalf("expected authority ladder readback, got %#v", status)
	}
	if status.AuthorityLadder.CurrentClass != "test_only" {
		t.Fatalf("expected current class test_only until low-risk live evidence exists, got %#v", status.AuthorityLadder)
	}
	if status.AuthorityLadder.NextClass != "low_risk_code" {
		t.Fatalf("expected next live class low_risk_code, got %#v", status.AuthorityLadder)
	}
	if !containsString(status.AuthorityLadder.DryRunReadyClasses, "multi_repo_low_risk") {
		t.Fatalf("expected multi_repo_low_risk dry-run readiness, got %#v", status.AuthorityLadder.DryRunReadyClasses)
	}
	for _, want := range []string{
		"dry_run:multi_repo_low_risk",
		"ordered_pr_dependency:ao-atlas:first",
		"ordered_pr_dependency:ao-foundry:after:ao-atlas",
		"ordered_pr_dependency:ao-command:after:ao-foundry",
		"per_repo_rollback:ao-atlas",
		"per_repo_rollback:ao-foundry",
		"per_repo_rollback:ao-command",
		"command_readback:multi_repo_low_risk",
		"prevent_concurrent_unsafe_execution",
	} {
		if !containsString(status.AuthorityLadder.RequiredEvidence, want) {
			t.Fatalf("expected required evidence %s, got %#v", want, status.AuthorityLadder.RequiredEvidence)
		}
	}
	if !containsString(status.AuthorityLadder.DoNotAdvanceGates, "do_not_advance:low_risk_code_live_execution_denied") ||
		!containsString(status.AuthorityLadder.DoNotAdvanceGates, "do_not_advance:multi_repo_low_risk_live_execution_denied") {
		t.Fatalf("expected do-not-advance gates for live code and multi-repo execution, got %#v", status.AuthorityLadder.DoNotAdvanceGates)
	}
	if status.SchedulesWork || status.ExecutesWork {
		t.Fatalf("multi-repo dry-run readback must remain non-mutating: %#v", status)
	}
	if status.AuthorityLadder.DeniedHigherClasses["multi_repo_low_risk"] == "" ||
		status.AuthorityLadder.DeniedHigherClasses["complex_repo_mutation"] == "" {
		t.Fatalf("expected multi-repo and complex classes to remain denied for live execution, got %#v", status.AuthorityLadder.DeniedHigherClasses)
	}
}

func TestComplexRepoMutationRehearsalWorkgraphStaysDryRunUntilLowerLiveEvidence(t *testing.T) {
	workgraph, err := LoadJSON[Workgraph](filepath.Join("..", "..", "examples", "valid", "workgraph-complex-repo-mutation-rehearsal.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateWorkgraph(workgraph); err != nil {
		t.Fatal(err)
	}
	if len(workgraph.Nodes) < 12 {
		t.Fatalf("complex rehearsal must have at least 12 nodes, got %d", len(workgraph.Nodes))
	}
	for _, want := range []string{
		"context-repack",
		"repair-plan",
		"blocked",
		"dependency-gate",
		"promotion-gate",
		"sentinel",
		"promoter",
		"command",
	} {
		if !workgraphHasNodeContaining(workgraph, want) {
			t.Fatalf("complex rehearsal fixture missing %s node", want)
		}
	}
	for _, want := range []string{
		"all_lower_classes_live_rehearsed",
		"rollback_plan:complex_repo_mutation",
		"sentinel_no_hold:complex_repo_mutation",
		"promoter_ready:complex_repo_mutation",
		"command_readback:complex_repo_mutation",
	} {
		if !workgraphHasRequiredEvidence(workgraph, want) {
			t.Fatalf("complex rehearsal fixture missing required evidence %s", want)
		}
	}
	if !workgraphHasSafetyLimitContaining(workgraph, "do_not_advance:complex_repo_mutation_live_execution_denied") ||
		!workgraphHasSafetyLimitContaining(workgraph, "do_not_advance:fully_unsupervised_complex_repo_mutation_denied") {
		t.Fatalf("complex rehearsal must keep live and fully unsupervised execution denied")
	}
	for _, node := range workgraph.Nodes {
		if node.FactoryTask.MutationClass != "complex_repo_mutation" {
			t.Fatalf("complex rehearsal node %s has wrong class %s", node.ID, node.FactoryTask.MutationClass)
		}
		if node.Status == "completed" && containsString(node.FactoryTask.RequiredEvidence, "live_rehearsal:complex_repo_mutation") {
			t.Fatalf("complex rehearsal must not claim live complex evidence: %#v", node)
		}
		if node.FactoryTask.AuthorityBoundary != "atlas_classification_only" {
			t.Fatalf("Atlas complex rehearsal must remain classification-only: %#v", node.FactoryTask)
		}
	}
}

func TestFoundryRoundtripSmokeValidatesFoundryImport(t *testing.T) {
	script, err := os.ReadFile(filepath.Join("..", "..", "scripts", "atlas-foundry-roundtrip-smoke.sh"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(script)
	for _, want := range []string{
		"foundry import",
		"--workgraph examples/valid/workgraph.json",
		"--instance examples/valid/stack-instance.json",
		"foundry atlas import validate",
		"foundry_import_validation",
		"mutation_class",
		"authority_boundary",
		"foundry atlas readback",
		"FOUNDRY_READBACK",
		"foundry_readback",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("roundtrip smoke missing %q", want)
		}
	}
}

func TestFactoryMaterializeDryRunWritesBoundedSkeleton(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "factory-materialization")
	var out bytes.Buffer
	code := Run([]string{"factory", "materialize", "--task", filepath.Join("..", "..", "examples", "valid", "factory-task.json"), "--out", outDir, "--dry-run"}, &out, &out)
	if code != 0 {
		t.Fatalf("materialize failed: %s", out.String())
	}
	for _, rel := range []string{"README.md", "task.json", "verification.txt", "materialization.json", filepath.Join("evidence", "README.md"), filepath.Join("context", "README.md")} {
		if _, err := os.Stat(filepath.Join(outDir, rel)); err != nil {
			t.Fatalf("expected %s: %v", rel, err)
		}
	}
	manifest, err := LoadJSON[FactoryMaterialization](filepath.Join(outDir, "materialization.json"))
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Mode != "dry_run" || manifest.ExecutesWork {
		t.Fatalf("unexpected materialization manifest: %#v", manifest)
	}
	if strings.Contains(manifest.OutputRoot, string(os.PathSeparator)) {
		t.Fatalf("manifest must not record local output path: %#v", manifest)
	}
}

func TestFactoryMaterializeRequiresDryRun(t *testing.T) {
	dir := t.TempDir()
	var out bytes.Buffer
	code := Run([]string{"factory", "materialize", "--task", filepath.Join("..", "..", "examples", "valid", "factory-task.json"), "--out", filepath.Join(dir, "factory-materialization")}, &out, &out)
	if code == 0 {
		t.Fatal("expected materialize without --dry-run to fail")
	}
	if !strings.Contains(out.String(), "--dry-run") {
		t.Fatalf("expected dry-run error, got %s", out.String())
	}
}

func TestWorkgraphMaterializeNextDryRunUsesNextReadyTask(t *testing.T) {
	dir := t.TempDir()
	outDir := filepath.Join(dir, "next-materialization")
	var out bytes.Buffer
	code := Run([]string{"workgraph", "materialize-next", "--workgraph", filepath.Join("..", "..", "examples", "valid", "workgraph.json"), "--out", outDir, "--dry-run"}, &out, &out)
	if code != 0 {
		t.Fatalf("materialize-next failed: %s", out.String())
	}
	manifest, err := LoadJSON[FactoryMaterialization](filepath.Join(outDir, "materialization.json"))
	if err != nil {
		t.Fatal(err)
	}
	if manifest.TaskID != "atlas-readiness-task" {
		t.Fatalf("expected atlas-readiness-task, got %s", manifest.TaskID)
	}
	if !strings.Contains(out.String(), "node=readiness-ready") {
		t.Fatalf("expected node readback, got %s", out.String())
	}
}

func TestWorkgraphMaterializeNextRequiresDryRun(t *testing.T) {
	dir := t.TempDir()
	var out bytes.Buffer
	code := Run([]string{"workgraph", "materialize-next", "--workgraph", filepath.Join("..", "..", "examples", "valid", "workgraph.json"), "--out", filepath.Join(dir, "next-materialization")}, &out, &out)
	if code == 0 {
		t.Fatal("expected materialize-next without --dry-run to fail")
	}
	if !strings.Contains(out.String(), "--dry-run") {
		t.Fatalf("expected dry-run error, got %s", out.String())
	}
}

func TestWorkgraphCompleteWritesNewCompletedWorkgraph(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "workgraph-completed.json")
	var out bytes.Buffer
	inputPath := filepath.Join("..", "..", "examples", "valid", "workgraph.json")
	code := Run([]string{"workgraph", "complete", "--workgraph", inputPath, "--run-link", filepath.Join("..", "..", "examples", "valid", "run-link.json"), "--out", outPath}, &out, &out)
	if code != 0 {
		t.Fatalf("complete failed: %s", out.String())
	}
	completed, err := LoadJSON[Workgraph](outPath)
	if err != nil {
		t.Fatal(err)
	}
	if completed.Nodes[1].Status != "completed" {
		t.Fatalf("expected completed node, got %#v", completed.Nodes[1])
	}
	original, err := LoadJSON[Workgraph](inputPath)
	if err != nil {
		t.Fatal(err)
	}
	if original.Nodes[1].Status != "ready" {
		t.Fatalf("input workgraph was modified: %#v", original.Nodes[1])
	}
	if !strings.Contains(out.String(), "node=readiness-ready") {
		t.Fatalf("expected node readback, got %s", out.String())
	}
}

func TestWorkgraphCompleteRejectsBlockedRunLink(t *testing.T) {
	assertWorkgraphCompleteFails(t, filepath.Join("..", "..", "examples", "valid", "workgraph.json"), filepath.Join("..", "..", "examples", "invalid", "run-link-blocked.json"), "completed")
}

func TestWorkgraphCompleteRejectsMissingNode(t *testing.T) {
	assertWorkgraphCompleteFails(t, filepath.Join("..", "..", "examples", "valid", "workgraph.json"), filepath.Join("..", "..", "examples", "invalid", "run-link-missing-node.json"), "matching")
}

func TestWorkgraphCompleteRejectsIncompleteDependency(t *testing.T) {
	assertWorkgraphCompleteFails(t, filepath.Join("..", "..", "examples", "invalid", "workgraph-complete-incomplete-dependency.json"), filepath.Join("..", "..", "examples", "valid", "run-link.json"), "dependencies")
}

func TestWorkgraphCompleteRejectsSameInputAndOutputPath(t *testing.T) {
	var out bytes.Buffer
	path := filepath.Join("..", "..", "examples", "valid", "workgraph.json")
	code := Run([]string{"workgraph", "complete", "--workgraph", path, "--run-link", filepath.Join("..", "..", "examples", "valid", "run-link.json"), "--out", path}, &out, &out)
	if code == 0 {
		t.Fatal("expected same input/output path to fail")
	}
	if !strings.Contains(out.String(), "overwrite input") {
		t.Fatalf("expected overwrite error, got %s", out.String())
	}
}

func TestWorkgraphRepairPlanEmitsRepairTaskForBlockedRunLink(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "repair-plan.json")
	var out bytes.Buffer
	code := Run([]string{"workgraph", "repair-plan", "--workgraph", filepath.Join("..", "..", "examples", "valid", "workgraph.json"), "--run-link", filepath.Join("..", "..", "examples", "invalid", "run-link-blocked.json"), "--out", outPath}, &out, &out)
	if code != 0 {
		t.Fatalf("repair-plan failed: %s", out.String())
	}
	plan, err := LoadJSON[WorkgraphRepairPlan](outPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateWorkgraphRepairPlan(plan); err != nil {
		t.Fatal(err)
	}
	if plan.TaskID != "atlas-readiness-task" || len(plan.RepairTasks) != 1 {
		t.Fatalf("unexpected repair plan: %#v", plan)
	}
	repair := plan.RepairTasks[0]
	if repair.ID != "repair-atlas-readiness-task" {
		t.Fatalf("unexpected repair task: %#v", plan.RepairTasks[0])
	}
	if plan.SchedulesWork || plan.ExecutesWork || plan.ApprovesWork {
		t.Fatalf("repair plan must not claim authority: %#v", plan)
	}
	if len(repair.ContextPackRefs) != 1 || repair.ContextPackRefs[0] != "examples/valid/context-pack.json" {
		t.Fatalf("repair task must preserve source context refs: %#v", repair)
	}
	for _, want := range []string{"do not schedule work from Atlas", "do not execute work from Atlas", "do not approve work from Atlas"} {
		if !containsString(repair.NonGoals, want) {
			t.Fatalf("repair task missing non-goal %q: %#v", want, repair.NonGoals)
		}
	}
	if !containsString(repair.SafetyLimits, "repair plan is readback only") {
		t.Fatalf("repair task missing readback-only safety limit: %#v", repair.SafetyLimits)
	}
}

func TestWorkgraphRepairPlanEmitsRepairTaskForFailedRunLink(t *testing.T) {
	plan, err := BuildWorkgraphRepairPlan(loadFixtureWorkgraph(t), loadFixtureRunLink(t, filepath.Join("..", "..", "examples", "valid", "run-link-failed.json")))
	if err != nil {
		t.Fatalf("repair plan failed: %v", err)
	}
	if err := ValidateWorkgraphRepairPlan(plan); err != nil {
		t.Fatal(err)
	}
	if plan.SourceRunLinkStatus != "failed" || len(plan.RepairTasks) != 1 {
		t.Fatalf("unexpected failed-run repair plan: %#v", plan)
	}
	if !strings.Contains(plan.Reason, "failed") {
		t.Fatalf("repair plan reason should name source failure status: %s", plan.Reason)
	}
}

func TestWorkgraphRepairPlanDemoFixtureValidates(t *testing.T) {
	plan, err := LoadJSON[WorkgraphRepairPlan](filepath.Join("..", "..", "examples", "valid", "workgraph-repair-plan-blocked-node-demo.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateWorkgraphRepairPlan(plan); err != nil {
		t.Fatal(err)
	}
	if plan.SourceRunLinkStatus != "blocked" || plan.SchedulesWork || plan.ExecutesWork || plan.ApprovesWork {
		t.Fatalf("demo repair plan must be blocked and authority-free: %#v", plan)
	}
}

func TestWorkgraphRepairPlanRejectsCompletedRunLink(t *testing.T) {
	assertWorkgraphRepairPlanFails(t, filepath.Join("..", "..", "examples", "valid", "run-link.json"), "blocked or failed")
}

func TestWorkgraphRepairPlanRejectsMissingNode(t *testing.T) {
	assertWorkgraphRepairPlanFails(t, filepath.Join("..", "..", "examples", "invalid", "run-link-missing-node-blocked.json"), "matching")
}

func assertWorkgraphRepairPlanFails(t *testing.T, runLinkPath, want string) {
	t.Helper()
	dir := t.TempDir()
	var out bytes.Buffer
	code := Run([]string{"workgraph", "repair-plan", "--workgraph", filepath.Join("..", "..", "examples", "valid", "workgraph.json"), "--run-link", runLinkPath, "--out", filepath.Join(dir, "repair-plan.json")}, &out, &out)
	if code == 0 {
		t.Fatalf("expected repair-plan to fail for %s", runLinkPath)
	}
	if !strings.Contains(out.String(), want) {
		t.Fatalf("expected error containing %q, got %s", want, out.String())
	}
}

func assertWorkgraphCompleteFails(t *testing.T, workgraphPath, runLinkPath, want string) {
	t.Helper()
	dir := t.TempDir()
	var out bytes.Buffer
	code := Run([]string{"workgraph", "complete", "--workgraph", workgraphPath, "--run-link", runLinkPath, "--out", filepath.Join(dir, "completed.json")}, &out, &out)
	if code == 0 {
		t.Fatalf("expected complete to fail for %s", runLinkPath)
	}
	if !strings.Contains(out.String(), want) {
		t.Fatalf("expected error containing %q, got %s", want, out.String())
	}
}

func TestRunLinkAttachWritesDigestBoundPublicSafeLink(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "run-link.json")
	var out bytes.Buffer
	code := Run([]string{
		"run-link", "attach",
		"--task-id", "atlas-readiness-task",
		"--status", "completed",
		"--evidence", "foundry=evidence/foundry/atlas-readiness.json",
		"--evidence", "forge=evidence/forge/atlas-readiness.json",
		"--evidence", "ao2=evidence/ao2/atlas-readiness.json",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("run-link attach failed: %s", out.String())
	}
	link, err := LoadJSON[RunLink](outPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateRunLink(link); err != nil {
		t.Fatal(err)
	}
	if link.TaskID != "atlas-readiness-task" || link.Status != "completed" {
		t.Fatalf("unexpected run link: %#v", link)
	}
	if link.Evidence["ao2"] != "evidence/ao2/atlas-readiness.json" {
		t.Fatalf("expected ao2 evidence path, got %#v", link.Evidence)
	}
	if !strings.HasPrefix(link.Digest, "sha256:") || link.Digest == "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" {
		t.Fatalf("expected computed digest, got %s", link.Digest)
	}
}

func TestRunLinkAttachRejectsPrivateEvidencePath(t *testing.T) {
	dir := t.TempDir()
	var out bytes.Buffer
	code := Run([]string{
		"run-link", "attach",
		"--task-id", "atlas-readiness-task",
		"--status", "completed",
		"--evidence", "ao2=/" + "Users/example/private.json",
		"--out", filepath.Join(dir, "run-link.json"),
	}, &out, &out)
	if code == 0 {
		t.Fatal("expected private evidence path to fail")
	}
	if !strings.Contains(out.String(), "private or machine-local path") {
		t.Fatalf("expected public-safety error, got %s", out.String())
	}
}

func loadFixtureWorkgraph(t *testing.T) Workgraph {
	t.Helper()
	workgraph, err := LoadJSON[Workgraph](filepath.Join("..", "..", "examples", "valid", "workgraph.json"))
	if err != nil {
		t.Fatal(err)
	}
	return workgraph
}

func loadFixtureRunLink(t *testing.T, path string) RunLink {
	t.Helper()
	link, err := LoadJSON[RunLink](path)
	if err != nil {
		t.Fatal(err)
	}
	return link
}

func complexNodeEvidenceWorkgraph(readyNodes ...string) Workgraph {
	ready := map[string]bool{}
	for _, nodeID := range readyNodes {
		ready[nodeID] = true
	}
	nodes := make([]WorkgraphNode, 0, 2)
	for _, nodeID := range []string{"complex-docs-intake", "complex-next-node"} {
		status := "blocked"
		blockers := []string{"dependency evidence not complete"}
		if ready[nodeID] {
			status = "ready"
			blockers = nil
		}
		nodes = append(nodes, WorkgraphNode{
			ID:           nodeID,
			Status:       status,
			FactoryTask:  complexNodeEvidenceTask(nodeID, "safe_to_execute:true"),
			Dependencies: nil,
			Blockers:     blockers,
		})
	}
	return Workgraph{
		ContractVersion: WorkgraphContract,
		ID:              "complex-rehearsal-test",
		TargetInstance:  "ao-stack",
		Nodes:           nodes,
	}
}

func complexNodeEvidenceImport(nodeID, safeEvidence string) FoundryImport {
	task := complexNodeEvidenceTask(nodeID, safeEvidence)
	return FoundryImport{
		ContractVersion: FoundryImportContract,
		ID:              "complex-rehearsal-test-foundry-import",
		WorkgraphID:     "complex-rehearsal-test",
		TargetInstance:  "ao-stack",
		Status:          "ready_for_foundry_fixture_import",
		SourceArtifacts: []SourceRef{{Ref: "examples/valid/workgraph.json", Digest: "sha256:0000000000000000000000000000000000000000000000000000000000000000"}},
		Tasks: []FoundryImportTaskFixture{{
			NodeID:            nodeID,
			TaskID:            task.ID,
			Path:              "tasks/" + task.ID + ".json",
			MutationClass:     task.MutationClass,
			WriteScope:        task.WriteScope,
			RollbackScope:     task.RollbackScope,
			RequiredGates:     task.RequiredGates,
			RequiredEvidence:  task.RequiredEvidence,
			AuthorityBoundary: task.AuthorityBoundary,
			Task:              task,
			TaskHash:          digestFactoryTask(task),
		}},
	}
}

func complexNodeEvidenceTask(nodeID, safeEvidence string) FactoryTask {
	taskID := nodeID + "-task"
	return FactoryTask{
		ContractVersion:   FactoryTaskContract,
		ID:                taskID,
		Objective:         "Prepare a bounded complex repo mutation rehearsal node for Foundry gate evaluation.",
		TargetFactoryRepo: "ao-atlas",
		FactoryFolder:     "factory/complex-repo-mutation-rehearsal/" + nodeID,
		MutationClass:     "complex_repo_mutation",
		Acceptance:        []string{"node evidence is internally consistent"},
		NonGoals:          []string{"do not execute live mutation from Atlas"},
		WriteScope:        []string{"factory/complex-repo-mutation-rehearsal/" + nodeID},
		RequiredGates:     []string{"atlas_complex_node_gate", "rollback_record_ready"},
		RollbackScope:     []string{"factory/complex-repo-mutation-rehearsal/" + nodeID},
		Verification:      []string{"go test ./..."},
		RequiredEvidence:  []string{"candidate_record", "rollback_record", safeEvidence},
		SafetyLimits:      []string{"Atlas emits evidence only; Foundry controls execution"},
		AuthorityBoundary: "atlas_evidence_only_foundry_executes",
		ContextPackRefs:   []string{"examples/valid/context-pack.json"},
	}
}

func complexNodeEvidenceCandidate(nodeID string, safe bool) map[string]any {
	return map[string]any{
		"node_id":          nodeID,
		"task_id":          nodeID + "-task",
		"status":           "ready",
		"executable_ready": true,
		"safe_to_execute":  safe,
		"required_gates":   []any{"atlas_complex_node_gate", "rollback_record_ready"},
	}
}

func complexNodeEvidenceRollback(nodeID string, safe bool) map[string]any {
	return map[string]any{
		"node_id":         nodeID,
		"task_id":         nodeID + "-task",
		"status":          "ready",
		"safe_to_execute": safe,
	}
}

func complexMissionContinuationWorkgraph() Workgraph {
	return Workgraph{
		ContractVersion: WorkgraphContract,
		ID:              "complex-rehearsal-continuation-test",
		TargetInstance:  "ao-stack",
		Nodes: []WorkgraphNode{
			{
				ID:           "complex-docs-intake",
				Status:       "ready",
				FactoryTask:  complexNodeEvidenceTask("complex-docs-intake", "safe_to_execute:true"),
				Dependencies: nil,
			},
			{
				ID:           "complex-test-scope",
				Status:       "blocked",
				FactoryTask:  complexNodeEvidenceTask("complex-test-scope", "safe_to_execute:false"),
				Dependencies: []string{"complex-docs-intake"},
				Blockers:     []string{"test-only node waits for dependency stop gate"},
			},
			{
				ID:           "complex-low-risk-code-scope",
				Status:       "blocked",
				FactoryTask:  complexNodeEvidenceTask("complex-low-risk-code-scope", "safe_to_execute:false"),
				Dependencies: []string{"complex-test-scope"},
				Blockers:     []string{"low-risk code node waits for dependency stop gate"},
			},
		},
	}
}

func workgraphNodesByID(workgraph Workgraph) map[string]WorkgraphNode {
	nodes := map[string]WorkgraphNode{}
	for _, node := range workgraph.Nodes {
		nodes[node.ID] = node
	}
	return nodes
}

func readyNodeIDs(workgraph Workgraph) []string {
	ids := []string{}
	for _, node := range workgraph.Nodes {
		if node.Status == "ready" {
			ids = append(ids, node.ID)
		}
	}
	return ids
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func mutationClassByName(model MutationClassModel, name string) (MutationClassDefinition, bool) {
	for _, class := range model.Classes {
		if class.Name == name {
			return class, true
		}
	}
	return MutationClassDefinition{}, false
}

func workgraphHasNodeContaining(workgraph Workgraph, marker string) bool {
	for _, node := range workgraph.Nodes {
		if strings.Contains(node.ID, marker) {
			return true
		}
	}
	return false
}

func workgraphHasRequiredEvidence(workgraph Workgraph, evidence string) bool {
	for _, node := range workgraph.Nodes {
		if containsString(node.FactoryTask.RequiredEvidence, evidence) {
			return true
		}
	}
	return false
}

func workgraphHasSafetyLimitContaining(workgraph Workgraph, marker string) bool {
	for _, node := range workgraph.Nodes {
		if stringSliceContains(node.FactoryTask.SafetyLimits, marker) {
			return true
		}
	}
	return false
}

func stringSliceContains(values []string, marker string) bool {
	for _, value := range values {
		if strings.Contains(value, marker) {
			return true
		}
	}
	return false
}

func fixtureWorkgraph() Workgraph {
	baseTask := FactoryTask{
		ContractVersion:   FactoryTaskContract,
		ID:                "factory-task",
		Objective:         "Create bounded AO Atlas task material.",
		TargetFactoryRepo: "ao-foundry",
		FactoryFolder:     "factory/atlas-demo",
		MutationClass:     "docs_only_single_file",
		Acceptance:        []string{"evidence exists"},
		NonGoals:          []string{"do not execute"},
		WriteScope:        []string{"factory/atlas-demo"},
		RequiredGates:     []string{"atlas_classification"},
		RollbackScope:     []string{"factory/atlas-demo"},
		Verification:      []string{"go test ./..."},
		RequiredEvidence:  []string{"summary.json"},
		SafetyLimits:      []string{"no provider calls"},
		AuthorityBoundary: "atlas_classification_only",
	}
	return Workgraph{
		ContractVersion: WorkgraphContract,
		ID:              "wg",
		TargetInstance:  "demo",
		Nodes: []WorkgraphNode{
			{ID: "done", Status: "completed", FactoryTask: baseTask},
			{ID: "task-ready", Status: "ready", Dependencies: []string{"done"}, FactoryTask: baseTask},
			{ID: "task-blocked", Status: "blocked", Blockers: []string{"needs Blueprint"}, FactoryTask: baseTask},
			{ID: "task-ready-2", Status: "ready", FactoryTask: baseTask},
		},
	}
}
