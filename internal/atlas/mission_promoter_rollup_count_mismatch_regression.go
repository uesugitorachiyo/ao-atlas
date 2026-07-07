package atlas

import (
	"fmt"
	"strings"
)

type promoterRollupCountMismatchRegressionCaseSpec struct {
	id                    string
	mutation              string
	expectedErrorContains string
	mutate                func(*AtlasPromoterNoPromotionRollup)
}

func BuildAtlasPromoterRollupCountMismatchRegression(nodeID, sourceRollupPath string) (AtlasPromoterRollupCountMismatchRegression, error) {
	nodeID = strings.TrimSpace(nodeID)
	sourceRollupPath = strings.TrimSpace(sourceRollupPath)
	if nodeID == "" {
		return AtlasPromoterRollupCountMismatchRegression{}, fmt.Errorf("node id is required")
	}
	if sourceRollupPath == "" {
		return AtlasPromoterRollupCountMismatchRegression{}, fmt.Errorf("source rollup path is required")
	}
	source, err := LoadJSON[AtlasPromoterNoPromotionRollup](sourceRollupPath)
	if err != nil {
		return AtlasPromoterRollupCountMismatchRegression{}, err
	}
	if err := ValidateAtlasPromoterNoPromotionRollup(source); err != nil {
		return AtlasPromoterRollupCountMismatchRegression{}, err
	}
	sourceDigest, err := digestTextFileWithNormalizedLineEndings(sourceRollupPath)
	if err != nil {
		return AtlasPromoterRollupCountMismatchRegression{}, err
	}

	specs := []promoterRollupCountMismatchRegressionCaseSpec{
		{
			id:                    "completed_nodes_total_mismatch",
			mutation:              "increment completed_nodes_total without updating wave summaries",
			expectedErrorContains: "completed_nodes_total",
			mutate: func(rollup *AtlasPromoterNoPromotionRollup) {
				rollup.CompletedNodesTotal++
			},
		},
		{
			id:                    "promoter_files_mismatch",
			mutation:              "increment promoter_no_promotion_files without adding evidence files",
			expectedErrorContains: "promoter_no_promotion_files",
			mutate: func(rollup *AtlasPromoterNoPromotionRollup) {
				rollup.PromoterNoPromotionFiles++
			},
		},
		{
			id:                    "missing_nodes_mismatch",
			mutation:              "increment missing_promoter_nodes_total without adding a missing node",
			expectedErrorContains: "missing_promoter_nodes_total",
			mutate: func(rollup *AtlasPromoterNoPromotionRollup) {
				rollup.MissingPromoterNodesTotal++
			},
		},
		{
			id:                    "no_promotion_status_mismatch",
			mutation:              "decrement no_promotion_status_count while promoter files stay unchanged",
			expectedErrorContains: "no_promotion_status_count",
			mutate: func(rollup *AtlasPromoterNoPromotionRollup) {
				rollup.NoPromotionStatusCount--
			},
		},
		{
			id:                    "rsi_denied_mismatch",
			mutation:              "decrement rsi_denied_count while promoter files stay unchanged",
			expectedErrorContains: "rsi_denied_count",
			mutate: func(rollup *AtlasPromoterNoPromotionRollup) {
				rollup.RSIDeniedCount--
			},
		},
	}

	regression := AtlasPromoterRollupCountMismatchRegression{
		Schema:                 AtlasPromoterRollupCountMismatchRegressionContract,
		NodeID:                 nodeID,
		Status:                 "mismatch_regression_recorded",
		SourceRollupPath:       publicArtifactRef(sourceRollupPath),
		SourceRollupDigest:     sourceDigest,
		CaseCount:              len(specs),
		Cases:                  make([]AtlasPromoterRollupCountMismatchRegressionCase, 0, len(specs)),
		SchedulesWork:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
		ClaimsAuthorityAdvance: false,
		RSIRemainsDenied:       true,
	}
	for _, spec := range specs {
		mutated := clonePromoterNoPromotionRollup(source)
		spec.mutate(&mutated)
		err := ValidateAtlasPromoterNoPromotionRollup(mutated)
		actual := ""
		if err != nil {
			actual = err.Error()
		}
		rejected := err != nil && strings.Contains(actual, spec.expectedErrorContains)
		if rejected {
			regression.RejectedCases++
		}
		regression.Cases = append(regression.Cases, AtlasPromoterRollupCountMismatchRegressionCase{
			ID:                    spec.id,
			Mutation:              spec.mutation,
			ExpectedErrorContains: spec.expectedErrorContains,
			Rejected:              rejected,
			ActualError:           actual,
		})
		switch spec.id {
		case "completed_nodes_total_mismatch":
			regression.CompletedNodesMismatchRejected = rejected
		case "promoter_files_mismatch":
			regression.PromoterFilesMismatchRejected = rejected
		case "missing_nodes_mismatch":
			regression.MissingNodesMismatchRejected = rejected
		case "no_promotion_status_mismatch":
			regression.NoPromotionStatusMismatchRejected = rejected
		case "rsi_denied_mismatch":
			regression.RSIDeniedMismatchRejected = rejected
		}
	}
	if regression.RejectedCases != regression.CaseCount {
		regression.Status = "mismatch_regression_failed"
	}
	if err := ValidateAtlasPromoterRollupCountMismatchRegression(regression); err != nil {
		return AtlasPromoterRollupCountMismatchRegression{}, err
	}
	return regression, nil
}

func clonePromoterNoPromotionRollup(source AtlasPromoterNoPromotionRollup) AtlasPromoterNoPromotionRollup {
	clone := source
	clone.EvidenceRoots = append([]string(nil), source.EvidenceRoots...)
	clone.WaveSummaries = append([]AtlasPromoterNoPromotionWaveSummary(nil), source.WaveSummaries...)
	clone.MissingPromoterNodes = append([]string(nil), source.MissingPromoterNodes...)
	clone.PromoterEvidenceFiles = append([]string(nil), source.PromoterEvidenceFiles...)
	return clone
}

func ValidateAtlasPromoterRollupCountMismatchRegression(regression AtlasPromoterRollupCountMismatchRegression) error {
	var errs []string
	requireContract(&errs, "promoter_rollup_count_mismatch_regression", regression.Schema, AtlasPromoterRollupCountMismatchRegressionContract)
	requireField(&errs, "node_id", regression.NodeID)
	checkPublicPath(&errs, "node_id", regression.NodeID, true)
	if !oneOf(regression.Status, "mismatch_regression_recorded", "mismatch_regression_failed") {
		errs = append(errs, "status must be mismatch_regression_recorded or mismatch_regression_failed")
	}
	requireField(&errs, "source_rollup_path", regression.SourceRollupPath)
	checkPublicPath(&errs, "source_rollup_path", regression.SourceRollupPath, true)
	if !digestPattern.MatchString(regression.SourceRollupDigest) {
		errs = append(errs, "source_rollup_digest must be sha256 digest")
	}
	if regression.CaseCount != 5 || len(regression.Cases) != regression.CaseCount {
		errs = append(errs, "case_count must be 5 and match cases length")
	}
	if regression.RejectedCases != regression.CaseCount {
		errs = append(errs, "rejected_cases must match case_count")
	}
	expectedCaseIDs := map[string]bool{
		"completed_nodes_total_mismatch": false,
		"promoter_files_mismatch":        false,
		"missing_nodes_mismatch":         false,
		"no_promotion_status_mismatch":   false,
		"rsi_denied_mismatch":            false,
	}
	for i, c := range regression.Cases {
		prefix := fmt.Sprintf("cases[%d]", i)
		requireField(&errs, prefix+".id", c.ID)
		requireField(&errs, prefix+".mutation", c.Mutation)
		requireField(&errs, prefix+".expected_error_contains", c.ExpectedErrorContains)
		requireField(&errs, prefix+".actual_error", c.ActualError)
		if _, ok := expectedCaseIDs[c.ID]; !ok {
			errs = append(errs, prefix+".id is not an expected mismatch case")
		} else if expectedCaseIDs[c.ID] {
			errs = append(errs, prefix+".id is duplicated")
		} else {
			expectedCaseIDs[c.ID] = true
		}
		if !c.Rejected {
			errs = append(errs, prefix+".rejected must be true")
		}
		if !strings.Contains(c.ActualError, c.ExpectedErrorContains) {
			errs = append(errs, prefix+".actual_error must contain expected_error_contains")
		}
	}
	for id, seen := range expectedCaseIDs {
		if !seen {
			errs = append(errs, "missing regression case "+id)
		}
	}
	if !regression.CompletedNodesMismatchRejected {
		errs = append(errs, "completed_nodes_mismatch_rejected must be true")
	}
	if !regression.PromoterFilesMismatchRejected {
		errs = append(errs, "promoter_files_mismatch_rejected must be true")
	}
	if !regression.MissingNodesMismatchRejected {
		errs = append(errs, "missing_nodes_mismatch_rejected must be true")
	}
	if !regression.NoPromotionStatusMismatchRejected {
		errs = append(errs, "no_promotion_status_mismatch_rejected must be true")
	}
	if !regression.RSIDeniedMismatchRejected {
		errs = append(errs, "rsi_denied_mismatch_rejected must be true")
	}
	if regression.Status == "mismatch_regression_recorded" && regression.RejectedCases != regression.CaseCount {
		errs = append(errs, "recorded mismatch regression requires all cases rejected")
	}
	validateNoAuthorityEffects(&errs, regression.SchedulesWork, regression.ExecutesWork, regression.ApprovesWork, regression.ClaimsAuthorityAdvance, regression.RSIRemainsDenied)
	return joinErrors(errs)
}

func WriteAtlasPromoterRollupCountMismatchRegression(path string, regression AtlasPromoterRollupCountMismatchRegression) error {
	return WriteJSON(path, regression)
}
