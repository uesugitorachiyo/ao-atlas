package atlas

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func BuildFoundryHandoff(workgraph Workgraph) FoundryHandoff {
	tasks := []FoundryTaskEntry{}
	for _, node := range workgraph.Nodes {
		if node.Status != "ready" {
			continue
		}
		task := node.FactoryTask
		tasks = append(tasks, FoundryTaskEntry{
			ID:                task.ID,
			Objective:         task.Objective,
			TargetFactoryRepo: task.TargetFactoryRepo,
			FactoryFolder:     task.FactoryFolder,
			Verification:      task.Verification,
			RequiredEvidence:  task.RequiredEvidence,
		})
	}
	return FoundryHandoff{
		ContractVersion: FoundryHandoffContract,
		ID:              workgraph.ID + "-foundry-handoff",
		TargetInstance:  workgraph.TargetInstance,
		Status:          "ready_for_foundry",
		Tasks:           tasks,
	}
}

// FoundryContinuationHandoffBuilder creates the operator-ready Foundry
// continuation prompt and validates that the import still matches the workgraph.
type FoundryContinuationHandoffBuilder struct {
	Workgraph     Workgraph
	FoundryImport FoundryImport
	Inputs        FoundryContinuationHandoffInputs
}

type FoundryContinuationHandoffInputs struct {
	BlueprintPackPath               string
	AtlasImportPath                 string
	WorkgraphPath                   string
	FoundryImportPath               string
	MissionContinuationEvidencePath string
}

func BuildFoundryContinuationHandoff(workgraph Workgraph, foundryImport FoundryImport, inputs FoundryContinuationHandoffInputs) (FoundryContinuationHandoff, error) {
	return FoundryContinuationHandoffBuilder{
		Workgraph:     workgraph,
		FoundryImport: foundryImport,
		Inputs:        inputs,
	}.Build()
}

func (b FoundryContinuationHandoffBuilder) Build() (FoundryContinuationHandoff, error) {
	if err := ValidateWorkgraph(b.Workgraph); err != nil {
		return FoundryContinuationHandoff{}, err
	}
	if err := ValidateFoundryImport(b.FoundryImport); err != nil {
		return FoundryContinuationHandoff{}, err
	}
	if err := ValidateFoundryImportMatchesWorkgraph(b.Workgraph, b.FoundryImport); err != nil {
		return FoundryContinuationHandoff{}, err
	}
	state, err := BuildWorkgraphState(b.Workgraph)
	if err != nil {
		return FoundryContinuationHandoff{}, err
	}
	firstSafeNode := ""
	if len(b.FoundryImport.Tasks) > 0 {
		firstSafeNode = b.FoundryImport.Tasks[0].NodeID
	}
	classBoundary := foundryContinuationClassBoundary(b.FoundryImport)
	handoff := FoundryContinuationHandoff{
		ContractVersion:                 FoundryContinuationHandoffContract,
		ID:                              b.FoundryImport.ID + "-continuation-handoff",
		TargetFolder:                    foundryContinuationTargetFolder(),
		Command:                         "codex --yolo",
		BlueprintPackPath:               slashOrNotProvided(b.Inputs.BlueprintPackPath),
		AtlasImportPath:                 slashOrNotProvided(b.Inputs.AtlasImportPath),
		WorkgraphPath:                   slashOrNotProvided(b.Inputs.WorkgraphPath),
		FoundryImportPath:               slashOrNotProvided(b.Inputs.FoundryImportPath),
		MissionContinuationEvidencePath: slashOrNotProvided(b.Inputs.MissionContinuationEvidencePath),
		FirstSafeNode:                   firstSafeNode,
		TotalNodeCount:                  len(b.Workgraph.Nodes),
		CompletedNodeCount:              state.NodeCounts["completed"],
		BlockedNodeCount:                state.NodeCounts["blocked"],
		ReadyNodeCount:                  state.NodeCounts["ready"],
		ClassBoundary:                   classBoundary,
		StopConditions: []string{
			"done",
			"final denial",
			"hard blocker",
			"CI failure",
			"unsafe scope drift",
			"kill switch",
		},
		SafetyProhibitions: []string{
			"Atlas must not execute live mutation",
			"no direct main mutation",
			"no release deploy publish upload tag provider call credential use dependency update auth policy widening secret env exposure or config expansion",
			"do not claim fully_unsupervised_complex_mutation or RSI is proven",
			"do not claim complex_repo_mutation is live-proven unless downstream evidence proves it",
		},
		SchedulesWork: false,
		ExecutesWork:  false,
		ApprovesWork:  false,
	}
	handoff.NextRecommendedAction = "Move to " + handoff.TargetFolder + "; Run codex --yolo; Paste this prompt"
	handoff.Prompt = buildFoundryContinuationPrompt(handoff)
	if err := ValidateFoundryContinuationHandoff(handoff); err != nil {
		return FoundryContinuationHandoff{}, err
	}
	return handoff, nil
}

func WriteFoundryContinuationPrompt(path string, handoff FoundryContinuationHandoff) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(foundryContinuationPromptMarkdown(handoff)), 0o644)
}

func foundryContinuationTargetFolder() string {
	return strings.Join([]string{"", "Users", "torachiyouesugi", "Documents", "public", "ao-foundry"}, "/")
}

func slashOrNotProvided(path string) string {
	if strings.TrimSpace(path) == "" {
		return "not_provided"
	}
	return filepath.ToSlash(path)
}

func foundryContinuationClassBoundary(foundryImport FoundryImport) string {
	classes := map[string]bool{}
	for _, task := range foundryImport.Tasks {
		if strings.TrimSpace(task.MutationClass) != "" {
			classes[task.MutationClass] = true
		}
	}
	if len(classes) == 0 {
		return "Atlas import only; Foundry must preserve Atlas no-execution boundary"
	}
	names := make([]string, 0, len(classes))
	for name := range classes {
		names = append(names, name)
	}
	sort.Strings(names)
	return "Atlas import only for " + strings.Join(names, ", ") + "; Foundry must preserve Atlas no-execution boundary"
}

func buildFoundryContinuationPrompt(handoff FoundryContinuationHandoff) string {
	var b strings.Builder
	b.WriteString("You are AO Foundry. Continue from the AO Atlas first-phase handoff.\n\n")
	b.WriteString("Source artifacts:\n")
	b.WriteString("- Blueprint pack: " + handoff.BlueprintPackPath + "\n")
	b.WriteString("- Atlas import: " + handoff.AtlasImportPath + "\n")
	b.WriteString("- Atlas workgraph: " + handoff.WorkgraphPath + "\n")
	b.WriteString("- Foundry import: " + handoff.FoundryImportPath + "\n")
	if handoff.MissionContinuationEvidencePath != "not_provided" {
		b.WriteString("- Mission continuation evidence: " + handoff.MissionContinuationEvidencePath + "\n")
	}
	b.WriteString("\nCurrent Atlas readback:\n")
	b.WriteString(fmt.Sprintf("- first safe node: %s\n", handoff.FirstSafeNode))
	b.WriteString(fmt.Sprintf("- total nodes: %d\n", handoff.TotalNodeCount))
	b.WriteString(fmt.Sprintf("- completed nodes: %d\n", handoff.CompletedNodeCount))
	b.WriteString(fmt.Sprintf("- ready nodes: %d\n", handoff.ReadyNodeCount))
	b.WriteString(fmt.Sprintf("- blocked nodes: %d\n", handoff.BlockedNodeCount))
	b.WriteString("- class boundary: " + handoff.ClassBoundary + "\n\n")
	b.WriteString("Required continuation behavior:\n")
	b.WriteString("- Move to AO Foundry.\n")
	b.WriteString("- Run codex --yolo.\n")
	b.WriteString("- Paste this prompt.\n")
	b.WriteString("- Import and validate the Foundry import.\n")
	b.WriteString("- do not stop after import validation.\n")
	b.WriteString("- do not stop after one gate artifact.\n")
	b.WriteString("- do not stop after one node.\n")
	b.WriteString("- Continue until all generated slices/tasks/nodes are consumed or a true hard blocker remains.\n")
	b.WriteString("- If evidence/schema/readback support is missing and can be safely implemented, implement it with PR/CI/merge.\n\n")
	b.WriteString("Hard safety prohibitions:\n")
	for _, prohibition := range handoff.SafetyProhibitions {
		b.WriteString("- " + prohibition + "\n")
	}
	b.WriteString("- fully_unsupervised_complex_mutation remains denied.\n")
	b.WriteString("- RSI remains denied.\n\n")
	b.WriteString("Stop conditions:\n")
	for _, condition := range handoff.StopConditions {
		b.WriteString("- " + condition + "\n")
	}
	b.WriteString("\nStop only on done, final denial, hard blocker, CI failure, unsafe scope drift, or kill switch.\n")
	return b.String()
}

func foundryContinuationPromptMarkdown(handoff FoundryContinuationHandoff) string {
	var b strings.Builder
	b.WriteString("# AO Foundry Continuation Handoff\n\n")
	b.WriteString("Move to AO Foundry:\n\n")
	b.WriteString("```sh\n")
	b.WriteString("cd " + handoff.TargetFolder + "\n")
	b.WriteString("codex --yolo\n")
	b.WriteString("```\n\n")
	b.WriteString("Paste this prompt:\n\n")
	b.WriteString("```text\n")
	b.WriteString(handoff.Prompt)
	b.WriteString("```\n")
	return b.String()
}
