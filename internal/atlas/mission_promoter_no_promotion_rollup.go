package atlas

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type promoterNoPromotionDecision struct {
	noPromotionStatus  bool
	promotionRequested bool
	promotionGranted   bool
	promotionClaimed   bool
	authorityAdvance   bool
	rsiDenied          bool
}

func BuildAtlasPromoterNoPromotionRollup(nodeID, sourceReadbackPath string, evidenceRoots []string) (AtlasPromoterNoPromotionRollup, error) {
	nodeID = strings.TrimSpace(nodeID)
	sourceReadbackPath = strings.TrimSpace(sourceReadbackPath)
	if nodeID == "" {
		return AtlasPromoterNoPromotionRollup{}, fmt.Errorf("node id is required")
	}
	if sourceReadbackPath == "" {
		return AtlasPromoterNoPromotionRollup{}, fmt.Errorf("source readback path is required")
	}
	readback, err := LoadJSON[AtlasRecommendationReadback](sourceReadbackPath)
	if err != nil {
		return AtlasPromoterNoPromotionRollup{}, err
	}
	if err := ValidateAtlasRecommendationReadback(readback); err != nil {
		return AtlasPromoterNoPromotionRollup{}, err
	}
	readbackDigest, err := digestTextFileWithNormalizedLineEndings(sourceReadbackPath)
	if err != nil {
		return AtlasPromoterNoPromotionRollup{}, err
	}

	trimmedRoots := []string{}
	for _, root := range evidenceRoots {
		root = strings.TrimSpace(root)
		if root == "" {
			continue
		}
		trimmedRoots = append(trimmedRoots, root)
	}
	if len(trimmedRoots) == 0 {
		return AtlasPromoterNoPromotionRollup{}, fmt.Errorf("at least one evidence root is required")
	}

	rollup := AtlasPromoterNoPromotionRollup{
		Schema:                             AtlasPromoterNoPromotionRollupContract,
		NodeID:                             nodeID,
		SourceReadbackPath:                 publicArtifactRef(sourceReadbackPath),
		SourceReadbackDigest:               readbackDigest,
		SourceReadbackCompletedNodes:       readback.CompletedNodes,
		SourceReadbackReadyNodes:           readback.ReadyNodes,
		SourceReadbackFirstExecutableNode:  readback.FirstExecutableNode,
		SourceReadbackFinalResponseAllowed: readback.FinalResponseAllowed,
		EvidenceRoots:                      make([]string, 0, len(trimmedRoots)),
		WaveSummaries:                      []AtlasPromoterNoPromotionWaveSummary{},
		MissingPromoterNodes:               []string{},
		PromoterEvidenceFiles:              []string{},
		AggregatePromotionStatus:           "no_promotion_requested",
		SchedulesWork:                      false,
		ExecutesWork:                       false,
		ApprovesWork:                       false,
		ClaimsAuthorityAdvance:             false,
		RSIRemainsDenied:                   true,
	}
	for _, root := range trimmedRoots {
		var summary AtlasPromoterNoPromotionWaveSummary
		var promoterFiles []string
		var missingNodes []string
		var err error
		if publicArtifactRef(root) == readback.EvidenceRoot {
			workgraphPath := filepath.Join(filepath.Dir(sourceReadbackPath), "workgraph-after.json")
			summary, promoterFiles, missingNodes, err = buildAtlasPromoterNoPromotionWaveSummaryForReadback(root, workgraphPath, readback)
		} else {
			summary, promoterFiles, missingNodes, err = buildAtlasPromoterNoPromotionWaveSummary(root)
		}
		if err != nil {
			return AtlasPromoterNoPromotionRollup{}, err
		}
		rollup.EvidenceRoots = append(rollup.EvidenceRoots, publicArtifactRef(root))
		rollup.WaveSummaries = append(rollup.WaveSummaries, summary)
		rollup.CompletedNodesTotal += summary.CompletedNodes
		rollup.PromoterNoPromotionFiles += summary.PromoterNoPromotionFiles
		rollup.MissingPromoterNodesTotal += summary.MissingPromoterNodes
		rollup.NoPromotionStatusCount += summary.NoPromotionStatusCount
		rollup.PromotionRequestedCount += summary.PromotionRequestedCount
		rollup.PromotionGrantedCount += summary.PromotionGrantedCount
		rollup.PromotionClaimedCount += summary.PromotionClaimedCount
		rollup.AuthorityAdvanceClaimCount += summary.AuthorityAdvanceClaimCount
		rollup.RSIDeniedCount += summary.RSIDeniedCount
		rollup.PromoterEvidenceFiles = append(rollup.PromoterEvidenceFiles, promoterFiles...)
		rollup.MissingPromoterNodes = append(rollup.MissingPromoterNodes, missingNodes...)
	}
	rollup.AllCompletedNodesCovered = rollup.MissingPromoterNodesTotal == 0 && rollup.PromoterNoPromotionFiles == rollup.CompletedNodesTotal
	rollup.AllPromoterStatusesNoPromotion = rollup.NoPromotionStatusCount == rollup.PromoterNoPromotionFiles
	rollup.PromotionRequested = rollup.PromotionRequestedCount > 0
	rollup.PromotionGranted = rollup.PromotionGrantedCount > 0
	rollup.ClaimsAuthorityAdvance = rollup.AuthorityAdvanceClaimCount > 0
	rollup.RSIRemainsDenied = rollup.RSIDeniedCount == rollup.PromoterNoPromotionFiles
	rollup.NoPromotionInvariantHolds = rollup.AllCompletedNodesCovered &&
		rollup.AllPromoterStatusesNoPromotion &&
		rollup.PromotionRequestedCount == 0 &&
		rollup.PromotionGrantedCount == 0 &&
		rollup.PromotionClaimedCount == 0 &&
		rollup.AuthorityAdvanceClaimCount == 0 &&
		rollup.RSIRemainsDenied
	rollup.Status = "no_promotion_rollup_bound"
	if !rollup.NoPromotionInvariantHolds {
		rollup.Status = "no_promotion_rollup_failed"
		rollup.AggregatePromotionStatus = "promotion_blocked"
	}
	if err := ValidateAtlasPromoterNoPromotionRollup(rollup); err != nil {
		return AtlasPromoterNoPromotionRollup{}, err
	}
	return rollup, nil
}

func buildAtlasPromoterNoPromotionWaveSummaryForReadback(evidenceRoot, workgraphPath string, readback AtlasRecommendationReadback) (AtlasPromoterNoPromotionWaveSummary, []string, []string, error) {
	completedNodes := []string{}
	for _, evidence := range readback.NodeEvidence {
		if evidence.Status == "completed" {
			completedNodes = append(completedNodes, evidence.NodeID)
		}
	}
	return buildAtlasPromoterNoPromotionWaveSummaryForNodes(evidenceRoot, workgraphPath, completedNodes)
}

func buildAtlasPromoterNoPromotionWaveSummary(evidenceRoot string) (AtlasPromoterNoPromotionWaveSummary, []string, []string, error) {
	workgraphPath, workgraph, err := latestCompletedWorkgraph(evidenceRoot)
	if err != nil {
		return AtlasPromoterNoPromotionWaveSummary{}, nil, nil, err
	}
	completedNodes := []string{}
	for _, node := range workgraph.Nodes {
		if node.Status == "completed" {
			completedNodes = append(completedNodes, node.ID)
		}
	}
	return buildAtlasPromoterNoPromotionWaveSummaryForNodes(evidenceRoot, workgraphPath, completedNodes)
}

func buildAtlasPromoterNoPromotionWaveSummaryForNodes(evidenceRoot, workgraphPath string, completedNodes []string) (AtlasPromoterNoPromotionWaveSummary, []string, []string, error) {
	summary := AtlasPromoterNoPromotionWaveSummary{
		EvidenceRoot:        publicArtifactRef(evidenceRoot),
		SourceWorkgraphPath: publicArtifactRef(workgraphPath),
		CompletedNodes:      len(completedNodes),
	}
	promoterFiles := []string{}
	missingNodes := []string{}
	for _, nodeID := range completedNodes {
		promoterPath := filepath.Join(evidenceRoot, "nodes", nodeID, "promoter_no_promotion.json")
		if _, err := os.Stat(promoterPath); err != nil {
			missingNodes = append(missingNodes, nodeID)
			continue
		}
		decision, err := loadPromoterNoPromotionDecision(promoterPath)
		if err != nil {
			return AtlasPromoterNoPromotionWaveSummary{}, nil, nil, err
		}
		summary.PromoterNoPromotionFiles++
		if decision.noPromotionStatus {
			summary.NoPromotionStatusCount++
		}
		if decision.promotionRequested {
			summary.PromotionRequestedCount++
		}
		if decision.promotionGranted {
			summary.PromotionGrantedCount++
		}
		if decision.promotionClaimed {
			summary.PromotionClaimedCount++
		}
		if decision.authorityAdvance {
			summary.AuthorityAdvanceClaimCount++
		}
		if decision.rsiDenied {
			summary.RSIDeniedCount++
		}
		promoterFiles = append(promoterFiles, publicArtifactRef(promoterPath))
	}
	summary.MissingPromoterNodes = len(missingNodes)
	summary.AllCompletedNodesCovered = summary.MissingPromoterNodes == 0 && summary.PromoterNoPromotionFiles == summary.CompletedNodes
	summary.AllPromoterStatusesNoPromotion = summary.NoPromotionStatusCount == summary.PromoterNoPromotionFiles
	summary.RSIRemainsDenied = summary.RSIDeniedCount == summary.PromoterNoPromotionFiles
	summary.NoPromotionInvariantHolds = summary.AllCompletedNodesCovered &&
		summary.AllPromoterStatusesNoPromotion &&
		summary.PromotionRequestedCount == 0 &&
		summary.PromotionGrantedCount == 0 &&
		summary.PromotionClaimedCount == 0 &&
		summary.AuthorityAdvanceClaimCount == 0 &&
		summary.RSIRemainsDenied
	return summary, promoterFiles, missingNodes, nil
}

func latestCompletedWorkgraph(evidenceRoot string) (string, Workgraph, error) {
	nodesRoot := filepath.Join(evidenceRoot, "nodes")
	bestPath := ""
	bestCompleted := -1
	var best Workgraph
	err := filepath.WalkDir(nodesRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || entry.Name() != "workgraph-after.json" {
			return nil
		}
		workgraph, err := LoadJSON[Workgraph](path)
		if err != nil {
			return err
		}
		completed := 0
		for _, node := range workgraph.Nodes {
			if node.Status == "completed" {
				completed++
			}
		}
		if completed > bestCompleted || (completed == bestCompleted && filepath.ToSlash(path) > filepath.ToSlash(bestPath)) {
			bestPath = path
			bestCompleted = completed
			best = workgraph
		}
		return nil
	})
	if err != nil {
		return "", Workgraph{}, err
	}
	if bestPath == "" {
		return "", Workgraph{}, fmt.Errorf("no workgraph-after.json found under %s", nodesRoot)
	}
	return bestPath, best, nil
}

func loadPromoterNoPromotionDecision(path string) (promoterNoPromotionDecision, error) {
	var raw map[string]any
	if err := readJSONIfPossible(path, &raw); err != nil {
		return promoterNoPromotionDecision{}, err
	}
	status, _ := stringField(raw, "status")
	evidenceKey, _ := stringField(raw, "evidence_key")
	decision := promoterNoPromotionDecision{
		noPromotionStatus: oneOf(status, "no_promotion_requested", "no_promotion") || (status == "recorded" && evidenceKey == "promoter_no_promotion"),
	}
	for _, field := range []struct {
		key string
		set func(bool)
	}{
		{"promotion_requested", func(value bool) { decision.promotionRequested = value }},
		{"promotion_granted", func(value bool) { decision.promotionGranted = value }},
		{"promotion_claimed", func(value bool) { decision.promotionClaimed = value }},
		{"claims_authority_advance", func(value bool) { decision.authorityAdvance = value }},
		{"rsi_remains_denied", func(value bool) { decision.rsiDenied = value }},
	} {
		value, ok, err := boolField(raw, field.key)
		if err != nil {
			return promoterNoPromotionDecision{}, fmt.Errorf("%s: %w", path, err)
		}
		if ok {
			field.set(value)
		}
	}
	if unchanged, ok, err := boolField(raw, "highest_proven_live_class_unchanged"); err != nil {
		return promoterNoPromotionDecision{}, fmt.Errorf("%s: %w", path, err)
	} else if ok && !unchanged {
		decision.authorityAdvance = true
	}
	if nested, ok := raw["promotion_decision"].(map[string]any); ok {
		if value, present, err := boolField(nested, "requested"); err != nil {
			return promoterNoPromotionDecision{}, fmt.Errorf("%s: %w", path, err)
		} else if present {
			decision.promotionRequested = value
		}
		if value, present, err := boolField(nested, "granted"); err != nil {
			return promoterNoPromotionDecision{}, fmt.Errorf("%s: %w", path, err)
		} else if present {
			decision.promotionGranted = value
		}
		if value, present, err := boolField(nested, "highest_proven_live_class_changed"); err != nil {
			return promoterNoPromotionDecision{}, fmt.Errorf("%s: %w", path, err)
		} else if present && value {
			decision.authorityAdvance = true
		}
		if rsiStatus, present := stringField(nested, "rsi_status"); present && rsiStatus == "denied" {
			decision.rsiDenied = true
		}
	}
	if !decision.noPromotionStatus && !decision.promotionRequested && !decision.promotionGranted && decision.rsiDenied {
		decision.noPromotionStatus = status == "recorded"
	}
	return decision, nil
}

func stringField(raw map[string]any, key string) (string, bool) {
	value, ok := raw[key]
	if !ok {
		return "", false
	}
	text, ok := value.(string)
	return strings.TrimSpace(text), ok
}

func boolField(raw map[string]any, key string) (bool, bool, error) {
	value, ok := raw[key]
	if !ok {
		return false, false, nil
	}
	flag, ok := value.(bool)
	if !ok {
		return false, true, fmt.Errorf("%s must be boolean", key)
	}
	return flag, true, nil
}

func ValidateAtlasPromoterNoPromotionRollup(rollup AtlasPromoterNoPromotionRollup) error {
	var errs []string
	requireContract(&errs, "promoter_no_promotion_rollup", rollup.Schema, AtlasPromoterNoPromotionRollupContract)
	requireField(&errs, "node_id", rollup.NodeID)
	checkPublicPath(&errs, "node_id", rollup.NodeID, true)
	if !oneOf(rollup.Status, "no_promotion_rollup_bound", "no_promotion_rollup_failed") {
		errs = append(errs, "status must be no_promotion_rollup_bound or no_promotion_rollup_failed")
	}
	requireField(&errs, "source_readback_path", rollup.SourceReadbackPath)
	checkPublicPath(&errs, "source_readback_path", rollup.SourceReadbackPath, true)
	if !digestPattern.MatchString(rollup.SourceReadbackDigest) {
		errs = append(errs, "source_readback_digest must be sha256 digest")
	}
	if rollup.SourceReadbackCompletedNodes <= 0 {
		errs = append(errs, "source_readback_completed_nodes must be positive")
	}
	if rollup.SourceReadbackReadyNodes < 0 {
		errs = append(errs, "source_readback_ready_nodes must not be negative")
	}
	requireField(&errs, "source_readback_first_executable_node", rollup.SourceReadbackFirstExecutableNode)
	if rollup.SourceReadbackFinalResponseAllowed {
		errs = append(errs, "source_readback_final_response_allowed must be false while feature-depth nodes remain")
	}
	requireList(&errs, "evidence_roots", rollup.EvidenceRoots)
	if len(rollup.WaveSummaries) == 0 {
		errs = append(errs, "wave_summaries is required")
	}
	if len(rollup.EvidenceRoots) != len(rollup.WaveSummaries) {
		errs = append(errs, "evidence_roots length must match wave_summaries length")
	}
	checkPublicStrings(&errs, "evidence_roots", rollup.EvidenceRoots, true)
	checkPublicStrings(&errs, "missing_promoter_nodes", rollup.MissingPromoterNodes, true)
	checkPublicStrings(&errs, "promoter_evidence_files", rollup.PromoterEvidenceFiles, true)
	totalCompleted := 0
	totalFiles := 0
	totalMissing := 0
	totalStatuses := 0
	totalRequested := 0
	totalGranted := 0
	totalClaimed := 0
	totalAuthority := 0
	totalRSIDenied := 0
	for i, summary := range rollup.WaveSummaries {
		validatePromoterNoPromotionWaveSummary(&errs, i, summary)
		totalCompleted += summary.CompletedNodes
		totalFiles += summary.PromoterNoPromotionFiles
		totalMissing += summary.MissingPromoterNodes
		totalStatuses += summary.NoPromotionStatusCount
		totalRequested += summary.PromotionRequestedCount
		totalGranted += summary.PromotionGrantedCount
		totalClaimed += summary.PromotionClaimedCount
		totalAuthority += summary.AuthorityAdvanceClaimCount
		totalRSIDenied += summary.RSIDeniedCount
	}
	if rollup.CompletedNodesTotal != totalCompleted {
		errs = append(errs, "completed_nodes_total must match wave summaries")
	}
	if rollup.PromoterNoPromotionFiles != totalFiles || rollup.PromoterNoPromotionFiles != len(rollup.PromoterEvidenceFiles) {
		errs = append(errs, "promoter_no_promotion_files must match wave summaries and evidence files")
	}
	if rollup.MissingPromoterNodesTotal != totalMissing || rollup.MissingPromoterNodesTotal != len(rollup.MissingPromoterNodes) {
		errs = append(errs, "missing_promoter_nodes_total must match wave summaries and missing node list")
	}
	if rollup.NoPromotionStatusCount != totalStatuses {
		errs = append(errs, "no_promotion_status_count must match wave summaries")
	}
	if rollup.PromotionRequestedCount != totalRequested || rollup.PromotionRequested != (totalRequested > 0) {
		errs = append(errs, "promotion_requested count and flag must match")
	}
	if rollup.PromotionGrantedCount != totalGranted || rollup.PromotionGranted != (totalGranted > 0) {
		errs = append(errs, "promotion_granted count and flag must match")
	}
	if rollup.PromotionClaimedCount != totalClaimed {
		errs = append(errs, "promotion_claimed_count must match wave summaries")
	}
	if rollup.AuthorityAdvanceClaimCount != totalAuthority || rollup.ClaimsAuthorityAdvance != (totalAuthority > 0) {
		errs = append(errs, "authority advance count and flag must match")
	}
	if rollup.RSIDeniedCount != totalRSIDenied {
		errs = append(errs, "rsi_denied_count must match wave summaries")
	}
	expectedCovered := rollup.MissingPromoterNodesTotal == 0 && rollup.PromoterNoPromotionFiles == rollup.CompletedNodesTotal
	if rollup.AllCompletedNodesCovered != expectedCovered {
		errs = append(errs, "all_completed_nodes_covered must match counts")
	}
	expectedStatuses := rollup.NoPromotionStatusCount == rollup.PromoterNoPromotionFiles
	if rollup.AllPromoterStatusesNoPromotion != expectedStatuses {
		errs = append(errs, "all_promoter_statuses_no_promotion must match counts")
	}
	expectedRSI := rollup.RSIDeniedCount == rollup.PromoterNoPromotionFiles
	if rollup.RSIRemainsDenied != expectedRSI {
		errs = append(errs, "rsi_remains_denied must match rsi denied count")
	}
	expectedInvariant := expectedCovered && expectedStatuses && rollup.PromotionRequestedCount == 0 && rollup.PromotionGrantedCount == 0 && rollup.PromotionClaimedCount == 0 && rollup.AuthorityAdvanceClaimCount == 0 && expectedRSI
	if rollup.NoPromotionInvariantHolds != expectedInvariant {
		errs = append(errs, "no_promotion_invariant_holds must match coverage, promotion, authority, and RSI counts")
	}
	if rollup.Status == "no_promotion_rollup_bound" && !rollup.NoPromotionInvariantHolds {
		errs = append(errs, "bound rollup requires no-promotion invariant")
	}
	if rollup.AggregatePromotionStatus != "no_promotion_requested" && rollup.AggregatePromotionStatus != "promotion_blocked" {
		errs = append(errs, "aggregate_promotion_status must be no_promotion_requested or promotion_blocked")
	}
	validateNoAuthorityEffects(&errs, rollup.SchedulesWork, rollup.ExecutesWork, rollup.ApprovesWork, rollup.ClaimsAuthorityAdvance, rollup.RSIRemainsDenied)
	return joinErrors(errs)
}

func validatePromoterNoPromotionWaveSummary(errs *[]string, index int, summary AtlasPromoterNoPromotionWaveSummary) {
	prefix := fmt.Sprintf("wave_summaries[%d]", index)
	requireField(errs, prefix+".evidence_root", summary.EvidenceRoot)
	checkPublicPath(errs, prefix+".evidence_root", summary.EvidenceRoot, true)
	requireField(errs, prefix+".source_workgraph_path", summary.SourceWorkgraphPath)
	checkPublicPath(errs, prefix+".source_workgraph_path", summary.SourceWorkgraphPath, true)
	if summary.CompletedNodes <= 0 {
		*errs = append(*errs, prefix+".completed_nodes must be positive")
	}
	if summary.PromoterNoPromotionFiles < 0 || summary.MissingPromoterNodes < 0 || summary.NoPromotionStatusCount < 0 || summary.PromotionRequestedCount < 0 || summary.PromotionGrantedCount < 0 || summary.PromotionClaimedCount < 0 || summary.AuthorityAdvanceClaimCount < 0 || summary.RSIDeniedCount < 0 {
		*errs = append(*errs, prefix+" counts must not be negative")
	}
	if summary.PromoterNoPromotionFiles+summary.MissingPromoterNodes != summary.CompletedNodes {
		*errs = append(*errs, prefix+" promoter file and missing counts must cover completed nodes")
	}
	if summary.AllCompletedNodesCovered != (summary.MissingPromoterNodes == 0 && summary.PromoterNoPromotionFiles == summary.CompletedNodes) {
		*errs = append(*errs, prefix+".all_completed_nodes_covered must match counts")
	}
	if summary.AllPromoterStatusesNoPromotion != (summary.NoPromotionStatusCount == summary.PromoterNoPromotionFiles) {
		*errs = append(*errs, prefix+".all_promoter_statuses_no_promotion must match counts")
	}
	if summary.RSIRemainsDenied != (summary.RSIDeniedCount == summary.PromoterNoPromotionFiles) {
		*errs = append(*errs, prefix+".rsi_remains_denied must match count")
	}
	expectedInvariant := summary.AllCompletedNodesCovered && summary.AllPromoterStatusesNoPromotion && summary.PromotionRequestedCount == 0 && summary.PromotionGrantedCount == 0 && summary.PromotionClaimedCount == 0 && summary.AuthorityAdvanceClaimCount == 0 && summary.RSIRemainsDenied
	if summary.NoPromotionInvariantHolds != expectedInvariant {
		*errs = append(*errs, prefix+".no_promotion_invariant_holds must match counts")
	}
}

func WriteAtlasPromoterNoPromotionRollup(path string, rollup AtlasPromoterNoPromotionRollup) error {
	return WriteJSON(path, rollup)
}
