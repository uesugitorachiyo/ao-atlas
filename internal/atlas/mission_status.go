package atlas

import (
	"fmt"
	"sort"
	"strings"
)

func BuildMissionStatus(intake Intake, workgraph Workgraph, links []RunLink) (MissionStatus, error) {
	if _, err := ValidateIntake(intake); err != nil {
		return MissionStatus{}, err
	}
	if err := ValidateWorkgraph(workgraph); err != nil {
		return MissionStatus{}, err
	}
	state, err := BuildWorkgraphState(workgraph)
	if err != nil {
		return MissionStatus{}, err
	}
	runLinks := map[string]string{}
	missingContextPacks := []string{}
	anyBlocked := false
	anyIncomplete := false
	for _, link := range links {
		if err := ValidateRunLink(link); err != nil {
			return MissionStatus{}, err
		}
		runLinks[link.TaskID] = link.Status
		if oneOf(link.Status, "blocked", "failed") {
			anyBlocked = true
		}
		if link.Status == "blocked" && strings.TrimSpace(link.Evidence["needs_context"]) != "" {
			missingContextPacks = append(missingContextPacks, link.TaskID)
		}
		if link.Status != "completed" {
			anyIncomplete = true
		}
	}
	missingHandoffs := state.MissingHandoffs(runLinks)
	completion := "in_progress"
	nextRecommendedAction := "handoff ready factory task to Foundry"
	nextActions := []string{"continue ready factory tasks through Foundry"}
	if anyBlocked || state.NodeCounts["blocked"] > 0 {
		completion = "blocked"
		if len(missingContextPacks) > 0 {
			nextRecommendedAction = "repack missing context before Foundry handoff"
			nextActions = []string{"emit context pack repack for needs_context run-link"}
		} else {
			nextRecommendedAction = "emit repair plan for blocked task"
			nextActions = []string{"emit repair plan or context repack for blocked task"}
		}
	} else if state.NodeCounts["ready"] == 0 && !anyIncomplete {
		completion = "completed"
		nextRecommendedAction = "record completion readback"
		nextActions = []string{"record completion readback and keep artifacts public-safe"}
	} else if len(missingHandoffs) > 0 {
		nextRecommendedAction = "emit Foundry handoff for ready nodes"
		nextActions = []string{"emit Foundry handoff/import material for ready nodes"}
	}
	finalAllowed, finalReason := atlasFinalResponseGate(completion, state.NodeCounts, missingHandoffs, nextRecommendedAction)
	status := MissionStatus{
		ContractVersion:          MissionStatusContract,
		IntakeID:                 intake.ID,
		WorkgraphID:              workgraph.ID,
		TargetInstance:           workgraph.TargetInstance,
		CompletionStatus:         completion,
		NodeCounts:               state.NodeCounts,
		RunLinks:                 runLinks,
		MissingContextPacks:      missingContextPacks,
		MissingHandoffs:          missingHandoffs,
		NextRecommendedAction:    nextRecommendedAction,
		NextActions:              nextActions,
		FinalResponseAllowed:     finalAllowed,
		FinalResponseReason:      finalReason,
		FinalStateReconciliation: buildAtlasFinalStateReconciliation(completion, state.NodeCounts, runLinks, finalAllowed, nextRecommendedAction),
		SchedulesWork:            false,
		ExecutesWork:             false,
	}
	if isAuthorityLadderWorkgraph(workgraph) {
		status.AuthorityLadder = BuildAuthorityLadderStatus(workgraph, links)
	}
	if err := ValidateMissionStatus(status); err != nil {
		return MissionStatus{}, err
	}
	return status, nil
}

func atlasFinalResponseGate(completion string, counts map[string]int, missingHandoffs []string, exactNextAction string) (bool, string) {
	if counts["ready"] > 0 || len(missingHandoffs) > 0 {
		return false, "ready nodes or exact next actions remain"
	}
	if completion == "completed" {
		return true, "workgraph is completed and no ready nodes remain"
	}
	if completion == "blocked" {
		return true, "workgraph is blocked for operator repair or support routing"
	}
	if strings.TrimSpace(exactNextAction) != "" {
		return false, "exact next action remains"
	}
	return true, "no ready work remains"
}

func buildAtlasFinalStateReconciliation(completion string, counts map[string]int, runLinks map[string]string, finalAllowed bool, nextAction string) *AtlasFinalStateReconciliation {
	foundryStatus := "not_bound"
	if len(runLinks) > 0 {
		foundryStatus = "run_links_recorded"
	}
	status := "ready"
	if !finalAllowed {
		status = "continuation_required"
	}
	return &AtlasFinalStateReconciliation{
		ContractVersion:       "ao.atlas.final-state-reconciliation.v0.1",
		Status:                status,
		WorkgraphStatus:       completion,
		FoundryRollupStatus:   foundryStatus,
		PromoterVerdictStatus: "not_bound",
		CommandReadbackStatus: "not_bound",
		ExactNextAction:       nextAction,
		SchedulesWork:         false,
		ExecutesWork:          false,
		ApprovesWork:          false,
	}
}

func isAuthorityLadderWorkgraph(workgraph Workgraph) bool {
	return strings.Contains(workgraph.ID, "authority-ladder")
}

func BuildAuthorityLadderStatus(workgraph Workgraph, links []RunLink) *AuthorityLadderStatus {
	runLinkStatus := map[string]string{}
	for _, link := range links {
		runLinkStatus[link.TaskID] = link.Status
	}
	order := mutationClassOrder()
	orderIndex := map[string]int{}
	for i, class := range order {
		orderIndex[class] = i
	}
	proven := map[string]bool{}
	dryRunReady := map[string]bool{}
	requiredEvidence := map[string]bool{}
	blockers := []string{}
	doNotAdvance := map[string]bool{}
	for _, node := range workgraph.Nodes {
		nodeClasses := classesInEvidence(node.FactoryTask.RequiredEvidence)
		for _, evidence := range node.FactoryTask.RequiredEvidence {
			if strings.HasPrefix(evidence, "live_rehearsal:") && (node.Status == "completed" || runLinkStatus[node.FactoryTask.ID] == "completed") {
				proven[strings.TrimPrefix(evidence, "live_rehearsal:")] = true
			}
			if strings.HasPrefix(evidence, "dry_run:") && oneOf(node.Status, "ready", "completed") {
				dryRunReady[strings.TrimPrefix(evidence, "dry_run:")] = true
			}
		}
		for _, limit := range node.FactoryTask.SafetyLimits {
			if strings.Contains(limit, "do_not_advance") {
				doNotAdvance[limit] = true
			}
		}
		for _, class := range nodeClasses {
			if class == "" {
				continue
			}
			if node.Status == "blocked" {
				for _, blocker := range node.Blockers {
					blockers = append(blockers, fmt.Sprintf("%s: %s", node.ID, blocker))
				}
			}
		}
	}
	current := "none"
	currentIndex := -1
	for _, class := range order {
		if proven[class] && orderIndex[class] > currentIndex {
			current = class
			currentIndex = orderIndex[class]
		}
	}
	next := ""
	for _, class := range order {
		if !proven[class] {
			next = class
			break
		}
	}
	if next == "" {
		next = "fully_unsupervised_complex_repo_mutation"
	}
	for _, node := range workgraph.Nodes {
		if !taskEvidenceMentionsClass(node.FactoryTask, next) {
			continue
		}
		for _, evidence := range node.FactoryTask.RequiredEvidence {
			if !strings.HasPrefix(evidence, "mutation_class:") {
				requiredEvidence[evidence] = true
			}
		}
	}
	denied := map[string]string{}
	denyFrom := currentIndex + 1
	if denyFrom < 0 {
		denyFrom = 0
	}
	for i := denyFrom; i < len(order); i++ {
		class := order[i]
		if class == next {
			continue
		}
		denied[class] = "denied until " + next + " live rehearsal, rollback proof, CI, Sentinel, Promoter, and Command evidence complete"
	}
	if next == "fully_unsupervised_complex_repo_mutation" {
		denied[next] = "denied until complex_repo_mutation live rehearsal evidence exists"
	}
	return &AuthorityLadderStatus{
		CurrentClass:        current,
		NextClass:           next,
		ProvenLiveClasses:   orderedSet(order, proven),
		DryRunReadyClasses:  orderedSet(order, dryRunReady),
		Blockers:            uniqueStrings(blockers),
		RequiredEvidence:    orderedEvidence(requiredEvidence),
		DeniedHigherClasses: denied,
		DoNotAdvanceGates:   orderedEvidence(doNotAdvance),
	}
}

func mutationClassOrder() []string {
	return []string{
		"docs_only_single_file",
		"docs_only_multi_file",
		"test_only",
		"low_risk_code",
		"multi_repo_low_risk",
		"complex_repo_mutation",
		"docs_config_only",
	}
}

func classesInEvidence(evidence []string) []string {
	classes := map[string]bool{}
	for _, item := range evidence {
		if strings.HasPrefix(item, "mutation_class:") {
			classes[strings.TrimPrefix(item, "mutation_class:")] = true
		}
		if strings.HasPrefix(item, "live_rehearsal:") {
			classes[strings.TrimPrefix(item, "live_rehearsal:")] = true
		}
		if strings.HasPrefix(item, "dry_run:") {
			classes[strings.TrimPrefix(item, "dry_run:")] = true
		}
	}
	return orderedEvidence(classes)
}

func taskEvidenceMentionsClass(task FactoryTask, class string) bool {
	for _, evidence := range task.RequiredEvidence {
		if strings.HasSuffix(evidence, ":"+class) {
			return true
		}
	}
	return false
}

func orderedSet(order []string, values map[string]bool) []string {
	result := []string{}
	for _, item := range order {
		if values[item] {
			result = append(result, item)
		}
	}
	return result
}

func orderedEvidence(values map[string]bool) []string {
	result := []string{}
	for item := range values {
		result = append(result, item)
	}
	sort.Strings(result)
	return result
}
