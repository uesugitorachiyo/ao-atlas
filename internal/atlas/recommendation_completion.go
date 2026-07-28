package atlas

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func CompleteAtlasRecommendationNodeWithRunLink(wave AtlasRecommendationWave, workgraph Workgraph, link RunLink, options AtlasRecommendationCompleteNodeOptions) (Workgraph, string, error) {
	if err := ValidateAtlasRecommendationWave(wave); err != nil {
		return Workgraph{}, "", err
	}
	if err := ValidateWorkgraph(workgraph); err != nil {
		return Workgraph{}, "", err
	}
	if err := ValidateRunLink(link); err != nil {
		return Workgraph{}, "", err
	}
	if wave.TargetInstance != workgraph.TargetInstance {
		return Workgraph{}, "", fmt.Errorf("target_instance mismatch between recommendation wave and workgraph")
	}
	if len(wave.Tasks) != len(workgraph.Nodes) {
		return Workgraph{}, "", fmt.Errorf("workgraph node count must match recommendation tasks")
	}
	state, err := BuildWorkgraphState(workgraph)
	if err != nil {
		return Workgraph{}, "", err
	}
	executable, ok := state.NextReadyNode()
	if !ok {
		return Workgraph{}, "", fmt.Errorf("no executable recommendation node remains")
	}
	expectedNodeID := strings.TrimSpace(options.ExpectedNodeID)
	if expectedNodeID != "" && executable.ID != expectedNodeID {
		return Workgraph{}, "", fmt.Errorf("expected executable node %s, got %s", expectedNodeID, executable.ID)
	}
	if link.Status != "completed" {
		return Workgraph{}, "", fmt.Errorf("run-link status must be completed")
	}
	if link.TaskID != executable.FactoryTask.ID {
		return Workgraph{}, "", fmt.Errorf("run-link task_id must match executable node %s task %s", executable.ID, executable.FactoryTask.ID)
	}
	if err := validateRecommendationRunLinkEvidence(executable.FactoryTask, link, options.EvidenceRoot); err != nil {
		return Workgraph{}, "", err
	}
	return CompleteWorkgraph(workgraph, link)
}

func validateRecommendationRunLinkEvidence(task FactoryTask, link RunLink, evidenceRoot string) error {
	for _, key := range requiredRecommendationRunLinkEvidence(task) {
		path := strings.TrimSpace(link.Evidence[key])
		if path == "" {
			return fmt.Errorf("missing evidence %s", key)
		}
		if strings.TrimSpace(evidenceRoot) == "" {
			continue
		}
		clean := filepath.Clean(path)
		if filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
			return fmt.Errorf("evidence %s must stay inside evidence root", key)
		}
		fullPath := filepath.Join(evidenceRoot, clean)
		if _, err := os.Stat(fullPath); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("evidence %s path does not exist: %s", key, filepath.ToSlash(clean))
			}
			return err
		}
		if err := validateRecommendationEvidenceRequiredFields(key, fullPath); err != nil {
			return err
		}
	}
	return nil
}

func validateRecommendationEvidenceRequiredFields(key, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("evidence %s must be a JSON object: %w", key, err)
	}
	status, ok := raw["status"].(string)
	if !ok || strings.TrimSpace(status) == "" {
		return fmt.Errorf("evidence %s missing required field status", key)
	}
	return nil
}

func requiredRecommendationRunLinkEvidence(task FactoryTask) []string {
	seen := map[string]bool{}
	keys := []string{}
	add := func(key string) {
		key = strings.TrimSpace(key)
		if key == "" || seen[key] {
			return
		}
		seen[key] = true
		keys = append(keys, key)
	}
	for _, gate := range task.RequiredGates {
		add(gate)
	}
	for _, key := range []string{
		"implementation_evidence",
		"foundry_import",
		"checkpoint_bundle",
	} {
		add(key)
	}
	return keys
}

func recommendationNodeEvidence(workgraph Workgraph) []AtlasRecommendationNodeEvidence {
	evidence := make([]AtlasRecommendationNodeEvidence, 0, len(workgraph.Nodes))
	for _, node := range workgraph.Nodes {
		evidence = append(evidence, AtlasRecommendationNodeEvidence{
			NodeID:                 node.ID,
			TaskID:                 node.FactoryTask.ID,
			Status:                 node.Status,
			NodeGate:               evidenceStatus(node.FactoryTask.RequiredGates, "node_gate"),
			CandidateRecord:        evidenceStatus(node.FactoryTask.RequiredGates, "candidate_record"),
			RollbackRecord:         evidenceStatus(node.FactoryTask.RequiredGates, "rollback_record"),
			ImplementationEvidence: "recorded",
			Tests:                  evidenceStatus(node.FactoryTask.RequiredGates, "tests"),
			Verification:           evidenceStatus(node.FactoryTask.RequiredGates, "verification"),
			PublicSafetyWording:    evidenceStatus(node.FactoryTask.RequiredGates, "sentinel_public_safety"),
			PromoterReadback:       evidenceStatus(node.FactoryTask.RequiredGates, "promoter_no_promotion"),
			CommandReadback:        evidenceStatus(node.FactoryTask.RequiredGates, "command_readback"),
			RequiredGates:          append([]string(nil), node.FactoryTask.RequiredGates...),
			VerificationCommands:   append([]string(nil), node.FactoryTask.Verification...),
		})
	}
	return evidence
}

func evidenceStatus(values []string, want string) string {
	for _, value := range values {
		if value == want {
			return "recorded"
		}
	}
	return "missing"
}

func featureDepthRecommendationReadback(tasks []AtlasRecommendationTask, limit int) []string {
	if limit <= 0 || limit > len(tasks) {
		limit = len(tasks)
	}
	items := make([]string, 0, limit)
	for _, task := range tasks[:limit] {
		items = append(items, task.ID+": "+task.Task)
	}
	return items
}

func atlasOwnedRecommendationTasks(tasks []AOMissionFeatureDepthTask, limit int) []AOMissionFeatureDepthTask {
	selected := []AOMissionFeatureDepthTask{}
	for _, task := range tasks {
		if task.Owner != "ao-atlas" {
			continue
		}
		selected = append(selected, task)
		if limit > 0 && len(selected) >= limit {
			break
		}
	}
	return selected
}
