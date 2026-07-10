package atlas

import "fmt"

type AtlasP0BRollbackAudit struct {
	Schema                       string                       `json:"schema"`
	NodeID                       string                       `json:"node_id"`
	Status                       string                       `json:"status"`
	Source                       string                       `json:"source"`
	CoveredNodeStart             int                          `json:"covered_node_start"`
	CoveredNodeEnd               int                          `json:"covered_node_end"`
	EntryCount                   int                          `json:"entry_count"`
	AllRollbackRecordsPresent    bool                         `json:"all_rollback_records_present"`
	AllRollbackRecordsReady      bool                         `json:"all_rollback_records_ready"`
	MissingRollbackRecordCount   int                          `json:"missing_rollback_record_count"`
	ReleaseOrDeployRollbackCount int                          `json:"release_or_deploy_rollback_count"`
	CompletedNodesBefore         int                          `json:"completed_nodes_before"`
	ReadyNodesBefore             int                          `json:"ready_nodes_before"`
	FinalResponseAllowed         bool                         `json:"final_response_allowed"`
	ExpectedNextNode             string                       `json:"expected_next_node"`
	Entries                      []AtlasP0BRollbackAuditEntry `json:"entries"`
	PromotionRequested           bool                         `json:"promotion_requested"`
	PromotionGranted             bool                         `json:"promotion_granted"`
	ClaimsAuthorityAdvance       bool                         `json:"claims_authority_advance"`
	RSIRemainsDenied             bool                         `json:"rsi_remains_denied"`
	SchedulesWork                bool                         `json:"schedules_work"`
	ExecutesWork                 bool                         `json:"executes_work"`
	ApprovesWork                 bool                         `json:"approves_work"`
}

type AtlasP0BRollbackAuditEntry struct {
	NodeNumber              int    `json:"node_number"`
	NodeID                  string `json:"node_id"`
	RollbackRecordPath      string `json:"rollback_record_path"`
	Status                  string `json:"status"`
	RollbackCommand         string `json:"rollback_command"`
	RollbackScopeCount      int    `json:"rollback_scope_count"`
	RequiresReleaseOrDeploy bool   `json:"requires_release_or_deploy"`
}

func ValidateAtlasP0BRollbackAudit(audit AtlasP0BRollbackAudit) error {
	var errs []string
	requireContract(&errs, "p0b_rollback_audit", audit.Schema, AtlasP0BRollbackAuditContract)
	if audit.Status != "complete" {
		errs = append(errs, "status must be complete")
	}
	requireField(&errs, "node_id", audit.NodeID)
	checkPublicPath(&errs, "node_id", audit.NodeID, true)
	requireField(&errs, "source", audit.Source)
	checkPublicPath(&errs, "source", audit.Source, true)
	if audit.CoveredNodeStart <= 0 || audit.CoveredNodeEnd < audit.CoveredNodeStart {
		errs = append(errs, "covered node range must be positive and ordered")
	}
	expectedCount := audit.CoveredNodeEnd - audit.CoveredNodeStart + 1
	if audit.EntryCount != len(audit.Entries) || audit.EntryCount != expectedCount {
		errs = append(errs, "entry_count must match covered node range and entries length")
	}
	if audit.MissingRollbackRecordCount < 0 || audit.ReleaseOrDeployRollbackCount < 0 || audit.CompletedNodesBefore < 0 || audit.ReadyNodesBefore < 0 {
		errs = append(errs, "summary counts must be non-negative")
	}
	if audit.FinalResponseAllowed {
		errs = append(errs, "final_response_allowed must be false while P0-B work remains")
	}
	requireField(&errs, "expected_next_node", audit.ExpectedNextNode)
	checkPublicPath(&errs, "expected_next_node", audit.ExpectedNextNode, true)
	seenNodes := map[int]bool{}
	readyCount := 0
	releaseRollbackCount := 0
	for i, entry := range audit.Entries {
		prefix := fmt.Sprintf("entries[%d]", i)
		validateP0BRollbackAuditEntry(&errs, prefix, entry)
		if entry.NodeNumber < audit.CoveredNodeStart || entry.NodeNumber > audit.CoveredNodeEnd {
			errs = append(errs, prefix+".node_number must be inside covered range")
		}
		if seenNodes[entry.NodeNumber] {
			errs = append(errs, prefix+".node_number must be unique")
		}
		seenNodes[entry.NodeNumber] = true
		if entry.Status == "ready" {
			readyCount++
		}
		if entry.RequiresReleaseOrDeploy {
			releaseRollbackCount++
		}
	}
	if audit.AllRollbackRecordsPresent != (audit.MissingRollbackRecordCount == 0 && len(audit.Entries) == expectedCount) {
		errs = append(errs, "all_rollback_records_present must match entries")
	}
	if audit.AllRollbackRecordsReady != (readyCount == len(audit.Entries) && len(audit.Entries) > 0) {
		errs = append(errs, "all_rollback_records_ready must match entries")
	}
	if audit.ReleaseOrDeployRollbackCount != releaseRollbackCount {
		errs = append(errs, "release_or_deploy_rollback_count must match entries")
	}
	if audit.PromotionRequested {
		errs = append(errs, "promotion_requested must be false")
	}
	if audit.PromotionGranted {
		errs = append(errs, "promotion_granted must be false")
	}
	validateNoAuthorityEffects(&errs, audit.SchedulesWork, audit.ExecutesWork, audit.ApprovesWork, audit.ClaimsAuthorityAdvance, audit.RSIRemainsDenied)
	return joinErrors(errs)
}

func validateP0BRollbackAuditEntry(errs *[]string, prefix string, entry AtlasP0BRollbackAuditEntry) {
	requireField(errs, prefix+".node_id", entry.NodeID)
	checkPublicPath(errs, prefix+".node_id", entry.NodeID, true)
	requireField(errs, prefix+".rollback_record_path", entry.RollbackRecordPath)
	checkPublicPath(errs, prefix+".rollback_record_path", entry.RollbackRecordPath, true)
	if entry.Status != "ready" {
		*errs = append(*errs, prefix+".status must be ready")
	}
	requireField(errs, prefix+".rollback_command", entry.RollbackCommand)
	checkPublicPath(errs, prefix+".rollback_command", entry.RollbackCommand, false)
	if entry.RollbackScopeCount <= 0 {
		*errs = append(*errs, prefix+".rollback_scope_count must be greater than zero")
	}
	if entry.RequiresReleaseOrDeploy {
		*errs = append(*errs, prefix+".requires_release_or_deploy must be false")
	}
}
