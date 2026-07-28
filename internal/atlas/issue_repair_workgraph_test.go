package atlas

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestIssueRepairWorkgraphContinuesAfterIndependentCandidateIsBlocked(t *testing.T) {
	workgraph := fixtureIssueRepairWorkgraph(t)
	state, err := BuildIssueRepairWorkgraphState(workgraph)
	if err != nil {
		t.Fatal(err)
	}

	first, node, ok := state.NextCandidate()
	if !ok || first.ID != "candidate-101" || node.ID != "issue-101" {
		t.Fatalf("unexpected first candidate: ok=%t candidate=%+v node=%+v", ok, first, node)
	}

	link := mustIssueRepairRunLink(t, node.FactoryTask.ID, "blocked")
	updated, candidateID, err := state.ApplyRunLink(link)
	if err != nil {
		t.Fatal(err)
	}
	if candidateID != first.ID {
		t.Fatalf("updated candidate %q, want %q", candidateID, first.ID)
	}

	updatedState, err := BuildIssueRepairWorkgraphState(updated)
	if err != nil {
		t.Fatal(err)
	}
	blocked, ok := updatedState.NodeState("issue-101")
	if !ok || blocked.Status != "blocked" || updatedState.NodeCounts["blocked"] != 1 {
		t.Fatalf("blocked outcome was not preserved: ok=%t state=%+v counts=%+v", ok, blocked, updatedState.NodeCounts)
	}
	next, nextNode, ok := updatedState.NextCandidate()
	if !ok || next.ID != "candidate-102" || nextNode.ID != "issue-102" {
		t.Fatalf("independent candidate did not remain safely executable: ok=%t candidate=%+v node=%+v", ok, next, nextNode)
	}
}

func TestIssueRepairWorkgraphDoesNotRunCandidateDependentOnBlockedCandidate(t *testing.T) {
	workgraph := fixtureIssueRepairWorkgraph(t)
	workgraph.Workgraph.Nodes[1].Dependencies = []string{"issue-101"}
	state, err := BuildIssueRepairWorkgraphState(workgraph)
	if err != nil {
		t.Fatal(err)
	}

	updated, _, err := state.ApplyRunLink(mustIssueRepairRunLink(t, "repair-issue-101", "failed"))
	if err != nil {
		t.Fatal(err)
	}
	updatedState, err := BuildIssueRepairWorkgraphState(updated)
	if err != nil {
		t.Fatal(err)
	}
	if updatedState.Candidates[0].Status != "failed" ||
		updatedState.Candidates[0].RunLink == nil ||
		updatedState.Candidates[0].RunLink.Digest == "" {
		t.Fatalf("failed candidate outcome was not preserved: %+v", updatedState.Candidates[0])
	}
	if next, node, ok := updatedState.NextCandidate(); ok {
		t.Fatalf("dependent candidate became executable after failed prerequisite: candidate=%+v node=%+v", next, node)
	}
}

func TestIssueRepairWorkgraphCompletesCandidateThroughGenericRunLink(t *testing.T) {
	state, err := BuildIssueRepairWorkgraphState(fixtureIssueRepairWorkgraph(t))
	if err != nil {
		t.Fatal(err)
	}

	updated, candidateID, err := state.ApplyRunLink(mustIssueRepairRunLink(t, "repair-issue-101", "completed"))
	if err != nil {
		t.Fatal(err)
	}
	if candidateID != "candidate-101" || updated.Workgraph.Nodes[0].Status != "completed" {
		t.Fatalf("completed transition did not use typed candidate mapping: candidate=%q workgraph=%+v", candidateID, updated.Workgraph)
	}
}

func TestIssueRepairWorkgraphRejectsMalformedCandidateMapping(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*IssueRepairWorkgraph)
		wantErr string
	}{
		{
			name: "unknown node",
			mutate: func(workgraph *IssueRepairWorkgraph) {
				workgraph.Candidates[0].NodeID = "missing-node"
			},
			wantErr: "does not reference a workgraph node",
		},
		{
			name: "duplicate candidate node",
			mutate: func(workgraph *IssueRepairWorkgraph) {
				workgraph.Candidates[1].NodeID = workgraph.Candidates[0].NodeID
			},
			wantErr: "node_id must be unique",
		},
		{
			name: "ambiguous task identity",
			mutate: func(workgraph *IssueRepairWorkgraph) {
				workgraph.Workgraph.Nodes[1].FactoryTask.ID = workgraph.Workgraph.Nodes[0].FactoryTask.ID
			},
			wantErr: "factory_task.id must be unique",
		},
		{
			name: "duplicate logical issue",
			mutate: func(workgraph *IssueRepairWorkgraph) {
				workgraph.Candidates[1].Repository = strings.ToUpper(workgraph.Candidates[0].Repository)
				workgraph.Candidates[1].IssueNumber = workgraph.Candidates[0].IssueNumber
			},
			wantErr: "repository and issue_number must be unique",
		},
		{
			name: "noncanonical repository identity",
			mutate: func(workgraph *IssueRepairWorkgraph) {
				workgraph.Candidates[1].Repository = workgraph.Candidates[0].Repository + " "
			},
			wantErr: "repository must be canonical owner/repository",
		},
		{
			name: "missing source provenance",
			mutate: func(workgraph *IssueRepairWorkgraph) {
				workgraph.Candidates[0].SourceSHA = ""
			},
			wantErr: "source_sha",
		},
		{
			name: "authority widened",
			mutate: func(workgraph *IssueRepairWorkgraph) {
				workgraph.ExecutesWork = true
			},
			wantErr: "executes_work must be false",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			workgraph := fixtureIssueRepairWorkgraph(t)
			test.mutate(&workgraph)
			_, err := BuildIssueRepairWorkgraphState(workgraph)
			if err == nil || !strings.Contains(err.Error(), test.wantErr) {
				t.Fatalf("expected %q rejection, got %v", test.wantErr, err)
			}
		})
	}
}

func TestIssueRepairWorkgraphRejectsPersistedOutcomeStatusTampering(t *testing.T) {
	state, err := BuildIssueRepairWorkgraphState(fixtureIssueRepairWorkgraph(t))
	if err != nil {
		t.Fatal(err)
	}
	updated, _, err := state.ApplyRunLink(mustIssueRepairRunLink(t, "repair-issue-101", "failed"))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "issue-repair-workgraph.json")
	if err := WriteJSON(path, updated); err != nil {
		t.Fatal(err)
	}
	persisted, err := LoadJSON[IssueRepairWorkgraph](path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := BuildIssueRepairWorkgraphState(persisted); err != nil {
		t.Fatalf("persisted valid outcome was rejected: %v", err)
	}
	persisted.Candidates[0].Status = "blocked"
	_, err = BuildIssueRepairWorkgraphState(persisted)
	if err == nil || !strings.Contains(err.Error(), "status must match terminal run-link") {
		t.Fatalf("expected persisted outcome status tampering rejection, got %v", err)
	}
}

func TestIssueRepairWorkgraphRejectsTamperedRunLinkDigest(t *testing.T) {
	state, err := BuildIssueRepairWorkgraphState(fixtureIssueRepairWorkgraph(t))
	if err != nil {
		t.Fatal(err)
	}
	link := mustIssueRepairRunLink(t, "repair-issue-101", "blocked")
	link.Status = "completed"
	_, _, err = state.ApplyRunLink(link)
	if err == nil || !strings.Contains(err.Error(), "digest does not match") {
		t.Fatalf("expected tampered run-link rejection, got %v", err)
	}
}

func TestIssueRepairWorkgraphDoesNotApplyBlockedRunLinkRetryToNextCandidate(t *testing.T) {
	state, err := BuildIssueRepairWorkgraphState(fixtureIssueRepairWorkgraph(t))
	if err != nil {
		t.Fatal(err)
	}
	link := mustIssueRepairRunLink(t, "repair-issue-101", "blocked")
	updated, _, err := state.ApplyRunLink(link)
	if err != nil {
		t.Fatal(err)
	}
	updatedState, err := BuildIssueRepairWorkgraphState(updated)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = updatedState.ApplyRunLink(link)
	if err == nil || !strings.Contains(err.Error(), "no dependency-ready matching") {
		t.Fatalf("expected blocked retry to refuse a second transition, got %v", err)
	}
	next, node, ok := updatedState.NextCandidate()
	if !ok || next.ID != "candidate-102" || node.ID != "issue-102" {
		t.Fatalf("blocked retry changed next candidate: ok=%t candidate=%+v node=%+v", ok, next, node)
	}
}

func TestIssueRepairWorkgraphRejectsNonTerminalRunLink(t *testing.T) {
	state, err := BuildIssueRepairWorkgraphState(fixtureIssueRepairWorkgraph(t))
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = state.ApplyRunLink(mustIssueRepairRunLink(t, "repair-issue-101", "running"))
	if err == nil || !strings.Contains(err.Error(), "completed, blocked, or failed") {
		t.Fatalf("expected non-terminal run-link rejection, got %v", err)
	}
}

func TestIssueRepairWorkgraphPreservesNoAuthorityBoundary(t *testing.T) {
	state, err := BuildIssueRepairWorkgraphState(fixtureIssueRepairWorkgraph(t))
	if err != nil {
		t.Fatal(err)
	}
	if state.SchedulesWork || state.ExecutesWork || state.ApprovesWork {
		t.Fatalf("typed planning state widened authority: %+v", state.IssueRepairWorkgraph)
	}
}

func fixtureIssueRepairWorkgraph(t *testing.T) IssueRepairWorkgraph {
	t.Helper()
	workgraph := fixtureWorkgraph()
	workgraph.ID = "issue-repair-wave"
	workgraph.Nodes = []WorkgraphNode{workgraph.Nodes[1], workgraph.Nodes[3]}
	workgraph.Nodes[0].ID = "issue-101"
	workgraph.Nodes[0].FactoryTask.ID = "repair-issue-101"
	workgraph.Nodes[0].Dependencies = nil
	workgraph.Nodes[1].ID = "issue-102"
	workgraph.Nodes[1].FactoryTask.ID = "repair-issue-102"
	workgraph.Nodes[1].Dependencies = nil
	return IssueRepairWorkgraph{
		ContractVersion: IssueRepairWorkgraphContract,
		MissionID:       "mission-e2539bc826abdbc0",
		Workgraph:       workgraph,
		Candidates: []IssueRepairCandidate{
			{
				ID:          "candidate-101",
				NodeID:      "issue-101",
				Repository:  "uesugitorachiyo/ao2",
				IssueNumber: 101,
				SourceSHA:   strings.Repeat("a", 40),
				Status:      "ready",
			},
			{
				ID:          "candidate-102",
				NodeID:      "issue-102",
				Repository:  "uesugitorachiyo/ao-command",
				IssueNumber: 102,
				SourceSHA:   strings.Repeat("b", 40),
				Status:      "ready",
			},
		},
		SchedulesWork: false,
		ExecutesWork:  false,
		ApprovesWork:  false,
	}
}

func mustIssueRepairRunLink(t *testing.T, taskID, status string) RunLink {
	t.Helper()
	link, err := BuildRunLink(taskID, status, map[string]string{
		"result": "evidence/" + taskID + ".json",
	})
	if err != nil {
		t.Fatal(err)
	}
	return link
}
