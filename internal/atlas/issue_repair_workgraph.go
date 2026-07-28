package atlas

import (
	"fmt"
	"regexp"
	"strings"
)

const IssueRepairWorkgraphContract = "ao.atlas.issue-repair-workgraph.v0.1"

var issueRepairSourceSHAPattern = regexp.MustCompile(`^[0-9a-f]{40}([0-9a-f]{24})?$`)
var issueRepairRepositoryPattern = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9-]{0,38})/[A-Za-z0-9._-]+$`)

type IssueRepairCandidate struct {
	ID          string   `json:"id"`
	NodeID      string   `json:"node_id"`
	Repository  string   `json:"repository"`
	IssueNumber int      `json:"issue_number"`
	SourceSHA   string   `json:"source_sha"`
	Status      string   `json:"status"`
	RunLink     *RunLink `json:"run_link,omitempty"`
}

type IssueRepairWorkgraph struct {
	ContractVersion string                 `json:"contract_version"`
	MissionID       string                 `json:"mission_id"`
	Workgraph       Workgraph              `json:"workgraph"`
	Candidates      []IssueRepairCandidate `json:"candidates"`
	SchedulesWork   bool                   `json:"schedules_work"`
	ExecutesWork    bool                   `json:"executes_work"`
	ApprovesWork    bool                   `json:"approves_work"`
}

type IssueRepairWorkgraphState struct {
	IssueRepairWorkgraph
	WorkgraphState
	candidateByNodeID map[string]IssueRepairCandidate
}

func BuildIssueRepairWorkgraphState(workgraph IssueRepairWorkgraph) (IssueRepairWorkgraphState, error) {
	if err := ValidateIssueRepairWorkgraph(workgraph); err != nil {
		return IssueRepairWorkgraphState{}, err
	}
	genericState, err := BuildWorkgraphState(workgraph.Workgraph)
	if err != nil {
		return IssueRepairWorkgraphState{}, err
	}
	state := IssueRepairWorkgraphState{
		IssueRepairWorkgraph: workgraph,
		WorkgraphState:       genericState,
		candidateByNodeID:    make(map[string]IssueRepairCandidate, len(workgraph.Candidates)),
	}
	for _, candidate := range workgraph.Candidates {
		state.candidateByNodeID[candidate.NodeID] = candidate
	}
	return state, nil
}

func ValidateIssueRepairWorkgraph(workgraph IssueRepairWorkgraph) error {
	var errs []string
	requireContract(&errs, "issue_repair_workgraph", workgraph.ContractVersion, IssueRepairWorkgraphContract)
	requireField(&errs, "mission_id", workgraph.MissionID)
	if err := ValidateWorkgraph(workgraph.Workgraph); err != nil {
		errs = append(errs, "workgraph: "+err.Error())
	}
	if len(workgraph.Candidates) == 0 {
		errs = append(errs, "candidates must not be empty")
	}

	nodesByID := make(map[string]WorkgraphNode, len(workgraph.Workgraph.Nodes))
	taskIDs := make(map[string]bool, len(workgraph.Workgraph.Nodes))
	for _, node := range workgraph.Workgraph.Nodes {
		nodesByID[node.ID] = node
		if taskIDs[node.FactoryTask.ID] {
			errs = append(errs, "workgraph node "+node.ID+" factory_task.id must be unique")
		}
		taskIDs[node.FactoryTask.ID] = true
	}
	candidateIDs := make(map[string]bool, len(workgraph.Candidates))
	candidateNodes := make(map[string]bool, len(workgraph.Candidates))
	logicalIssues := make(map[string]bool, len(workgraph.Candidates))
	for i, candidate := range workgraph.Candidates {
		prefix := fmt.Sprintf("candidates[%d]", i)
		requireField(&errs, prefix+".id", candidate.ID)
		requireField(&errs, prefix+".node_id", candidate.NodeID)
		requireField(&errs, prefix+".repository", candidate.Repository)
		requireField(&errs, prefix+".source_sha", candidate.SourceSHA)
		checkPublicPath(&errs, prefix+".repository", candidate.Repository, true)
		if !issueRepairRepositoryPattern.MatchString(candidate.Repository) {
			errs = append(errs, prefix+".repository must be canonical owner/repository")
		}
		if candidate.IssueNumber < 1 {
			errs = append(errs, prefix+".issue_number must be positive")
		}
		if candidate.SourceSHA != "" && !issueRepairSourceSHAPattern.MatchString(candidate.SourceSHA) {
			errs = append(errs, prefix+".source_sha must be 40 or 64 lowercase hex characters")
		}
		if !oneOf(candidate.Status, "ready", "completed", "blocked", "failed") {
			errs = append(errs, prefix+".status must be ready, completed, blocked, or failed")
		}
		if candidate.Status == "ready" && candidate.RunLink != nil {
			errs = append(errs, prefix+".run_link must be absent while ready")
		}
		if candidate.Status != "ready" && candidate.RunLink == nil {
			errs = append(errs, prefix+".run_link must bind a terminal outcome")
		}
		if candidate.RunLink != nil {
			if err := ValidateRunLink(*candidate.RunLink); err != nil {
				errs = append(errs, prefix+".run_link: "+err.Error())
			} else if candidate.RunLink.Digest != digestRunLink(*candidate.RunLink) {
				errs = append(errs, prefix+".run_link digest does not match its content")
			}
			if candidate.Status != candidate.RunLink.Status {
				errs = append(errs, prefix+".status must match terminal run-link")
			}
		}
		if candidateIDs[candidate.ID] {
			errs = append(errs, prefix+".id must be unique")
		}
		candidateIDs[candidate.ID] = true
		if candidateNodes[candidate.NodeID] {
			errs = append(errs, prefix+".node_id must be unique")
		}
		candidateNodes[candidate.NodeID] = true
		if _, ok := nodesByID[candidate.NodeID]; !ok {
			errs = append(errs, prefix+".node_id does not reference a workgraph node")
		} else {
			node := nodesByID[candidate.NodeID]
			nodeStatus := node.Status
			if candidate.Status == "ready" && nodeStatus != "ready" ||
				candidate.Status == "completed" && nodeStatus != "completed" ||
				oneOf(candidate.Status, "blocked", "failed") && nodeStatus != "blocked" {
				errs = append(errs, prefix+".status does not match workgraph node status")
			}
			if candidate.RunLink != nil && candidate.RunLink.TaskID != node.FactoryTask.ID {
				errs = append(errs, prefix+".run_link.task_id must match workgraph node factory_task.id")
			}
		}
		logicalIssue := fmt.Sprintf("%s#%d", strings.ToLower(candidate.Repository), candidate.IssueNumber)
		if logicalIssues[logicalIssue] {
			errs = append(errs, prefix+".repository and issue_number must be unique")
		}
		logicalIssues[logicalIssue] = true
	}
	for _, node := range workgraph.Workgraph.Nodes {
		if !candidateNodes[node.ID] {
			errs = append(errs, "workgraph node "+node.ID+" does not have an issue-repair candidate")
		}
	}
	if workgraph.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if workgraph.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if workgraph.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	return joinErrors(errs)
}

func (state IssueRepairWorkgraphState) NextCandidate() (IssueRepairCandidate, WorkgraphNode, bool) {
	node, ok := state.WorkgraphState.NextReadyNode()
	if !ok {
		return IssueRepairCandidate{}, WorkgraphNode{}, false
	}
	candidate, ok := state.candidateByNodeID[node.ID]
	if !ok {
		return IssueRepairCandidate{}, WorkgraphNode{}, false
	}
	return candidate, node, true
}

func (state IssueRepairWorkgraphState) ApplyRunLink(link RunLink) (IssueRepairWorkgraph, string, error) {
	if err := ValidateRunLink(link); err != nil {
		return IssueRepairWorkgraph{}, "", err
	}
	if link.Digest != digestRunLink(link) {
		return IssueRepairWorkgraph{}, "", fmt.Errorf("run-link digest does not match its content")
	}
	if !oneOf(link.Status, "completed", "blocked", "failed") {
		return IssueRepairWorkgraph{}, "", fmt.Errorf("run-link status must be completed, blocked, or failed")
	}
	if link.Status == "completed" {
		updated, nodeID, err := state.WorkgraphState.CompleteWithRunLink(link)
		if err != nil {
			return IssueRepairWorkgraph{}, "", err
		}
		return state.withUpdatedWorkgraph(updated, nodeID, link)
	}

	matchedTask := false
	for i, node := range state.IssueRepairWorkgraph.Workgraph.Nodes {
		if node.FactoryTask.ID != link.TaskID {
			continue
		}
		matchedTask = true
		nodeState, ok := state.NodeState(node.ID)
		if !ok || !nodeState.ExecutableReady {
			continue
		}
		updated := state.IssueRepairWorkgraph.Workgraph
		updated.Nodes = append([]WorkgraphNode(nil), state.IssueRepairWorkgraph.Workgraph.Nodes...)
		updated.Nodes[i].Status = "blocked"
		updated.Nodes[i].Blockers = []string{
			fmt.Sprintf("run-link %s recorded terminal status %s", link.Digest, link.Status),
		}
		if err := ValidateWorkgraph(updated); err != nil {
			return IssueRepairWorkgraph{}, "", err
		}
		return state.withUpdatedWorkgraph(updated, node.ID, link)
	}
	if !matchedTask {
		return IssueRepairWorkgraph{}, "", fmt.Errorf("no matching issue-repair node for run-link task_id %q", link.TaskID)
	}
	return IssueRepairWorkgraph{}, "", fmt.Errorf("no dependency-ready matching issue-repair node for run-link task_id %q", link.TaskID)
}

func (state IssueRepairWorkgraphState) withUpdatedWorkgraph(updated Workgraph, nodeID string, link RunLink) (IssueRepairWorkgraph, string, error) {
	candidate, ok := state.candidateByNodeID[nodeID]
	if !ok {
		return IssueRepairWorkgraph{}, "", fmt.Errorf("workgraph node %q does not map to an issue-repair candidate", nodeID)
	}
	result := state.IssueRepairWorkgraph
	result.Workgraph = updated
	result.Candidates = append([]IssueRepairCandidate(nil), state.Candidates...)
	for i := range result.Candidates {
		if result.Candidates[i].ID == candidate.ID {
			result.Candidates[i].Status = link.Status
			result.Candidates[i].RunLink = cloneIssueRepairRunLink(link)
			break
		}
	}
	if err := ValidateIssueRepairWorkgraph(result); err != nil {
		return IssueRepairWorkgraph{}, "", err
	}
	return result, candidate.ID, nil
}

func cloneIssueRepairRunLink(link RunLink) *RunLink {
	cloned := link
	cloned.Evidence = make(map[string]string, len(link.Evidence))
	for key, value := range link.Evidence {
		cloned.Evidence[key] = value
	}
	return &cloned
}
