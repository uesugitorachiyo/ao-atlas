package atlas

import (
	"fmt"
	"strings"
)

type ComplexExecutableNodeEvidence struct {
	Workgraph     Workgraph
	FoundryImport FoundryImport
	Candidate     map[string]any
	Rollback      map[string]any
	Summary       map[string]any
}

// ValidateComplexExecutableNodeEvidence checks the Atlas-owned handoff envelope
// for the first executable complex_repo_mutation node. It is intentionally
// stricter than generic workgraph/import validation because a selected node must
// have matching candidate, rollback, and import execution evidence.
func ValidateComplexExecutableNodeEvidence(workgraph Workgraph, foundryImport FoundryImport, candidate map[string]any, rollback map[string]any, summary map[string]any) error {
	return ValidateComplexExecutableNodeConsistency(ComplexExecutableNodeEvidence{
		Workgraph:     workgraph,
		FoundryImport: foundryImport,
		Candidate:     candidate,
		Rollback:      rollback,
		Summary:       summary,
	})
}

func ValidateComplexExecutableNodeConsistency(evidence ComplexExecutableNodeEvidence) error {
	if err := ValidateWorkgraph(evidence.Workgraph); err != nil {
		return err
	}
	if err := ValidateFoundryImport(evidence.FoundryImport); err != nil {
		return err
	}

	readyNodes := make([]WorkgraphNode, 0, 1)
	for _, node := range evidence.Workgraph.Nodes {
		if node.Status == "ready" {
			readyNodes = append(readyNodes, node)
		}
	}
	if len(readyNodes) != 1 {
		return fmt.Errorf("exactly one ready executable node is allowed")
	}
	readyNode := readyNodes[0]
	if readyNode.FactoryTask.MutationClass != "complex_repo_mutation" {
		return fmt.Errorf("ready node mutation_class must be complex_repo_mutation")
	}
	if containsValue(readyNode.FactoryTask.RequiredEvidence, "safe_to_execute:false") {
		return fmt.Errorf("ready node must not require safe_to_execute:false")
	}
	if !containsValue(readyNode.FactoryTask.RequiredEvidence, "safe_to_execute:true") {
		return fmt.Errorf("ready node must bind safe_to_execute:true")
	}

	if len(evidence.FoundryImport.Tasks) != 1 {
		return fmt.Errorf("foundry import must contain exactly one selected node")
	}
	importTask := evidence.FoundryImport.Tasks[0]
	if importTask.NodeID != readyNode.ID {
		return fmt.Errorf("foundry import node_id must match ready node")
	}
	if importTask.TaskID != readyNode.FactoryTask.ID {
		return fmt.Errorf("foundry import task_id must match ready node task")
	}
	if err := ValidateFoundryImportMatchesWorkgraph(evidence.Workgraph, evidence.FoundryImport); err != nil {
		return err
	}
	if containsValue(importTask.RequiredEvidence, "safe_to_execute:false") ||
		containsValue(importTask.Task.RequiredEvidence, "safe_to_execute:false") {
		return fmt.Errorf("foundry import must not require safe_to_execute:false")
	}
	if !containsValue(importTask.RequiredEvidence, "safe_to_execute:true") ||
		!containsValue(importTask.Task.RequiredEvidence, "safe_to_execute:true") {
		return fmt.Errorf("foundry import must bind safe_to_execute:true")
	}

	if err := validateComplexCandidateEvidence(evidence.Candidate, readyNode); err != nil {
		return err
	}
	if err := validateComplexRollbackEvidence(evidence.Rollback, readyNode); err != nil {
		return err
	}
	if firstSafeNode := complexEvidenceString(evidence.Summary, "first_safe_node"); firstSafeNode != "" && firstSafeNode != readyNode.ID {
		return fmt.Errorf("first_safe_node must match ready node")
	}
	return nil
}

func validateComplexCandidateEvidence(candidate map[string]any, readyNode WorkgraphNode) error {
	if complexEvidenceString(candidate, "node_id") != readyNode.ID {
		return fmt.Errorf("candidate record node_id must match ready node")
	}
	if complexEvidenceString(candidate, "task_id") != readyNode.FactoryTask.ID {
		return fmt.Errorf("candidate record task_id must match ready node task")
	}
	if complexEvidenceString(candidate, "status") != "ready" {
		return fmt.Errorf("candidate record status must be ready")
	}
	if !complexEvidenceBool(candidate, "executable_ready") {
		return fmt.Errorf("candidate record executable_ready must be true")
	}
	if !complexEvidenceBool(candidate, "safe_to_execute") {
		return fmt.Errorf("candidate record safe_to_execute must be true")
	}
	if containsValue(complexEvidenceStringSlice(candidate, "required_gates"), "safe_to_execute:false") {
		return fmt.Errorf("candidate record must not require safe_to_execute:false")
	}
	return nil
}

func validateComplexRollbackEvidence(rollback map[string]any, readyNode WorkgraphNode) error {
	if complexEvidenceString(rollback, "node_id") != readyNode.ID {
		return fmt.Errorf("rollback record node_id must match ready node")
	}
	if taskID := complexEvidenceString(rollback, "task_id"); taskID != "" && taskID != readyNode.FactoryTask.ID {
		return fmt.Errorf("rollback record task_id must match ready node task")
	}
	status := complexEvidenceString(rollback, "status")
	if status != "" && status != "ready" {
		return fmt.Errorf("rollback record status must be ready")
	}
	if !complexEvidenceBool(rollback, "safe_to_execute") {
		return fmt.Errorf("rollback record safe_to_execute must be true")
	}
	return nil
}

func complexEvidenceString(record map[string]any, key string) string {
	if record == nil {
		return ""
	}
	value, ok := record[key]
	if !ok {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}

func complexEvidenceBool(record map[string]any, key string) bool {
	if record == nil {
		return false
	}
	value, ok := record[key]
	if !ok {
		return false
	}
	flag, ok := value.(bool)
	return ok && flag
}

func complexEvidenceStringSlice(record map[string]any, key string) []string {
	if record == nil {
		return nil
	}
	value, ok := record[key]
	if !ok {
		return nil
	}
	switch typed := value.(type) {
	case []string:
		return typed
	case []any:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			if text, ok := item.(string); ok {
				result = append(result, strings.TrimSpace(text))
			}
		}
		return result
	default:
		return nil
	}
}
