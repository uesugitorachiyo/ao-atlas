package atlas

import "fmt"

func BuildAtlasMergeCheckBinding(inputPath string) (AtlasMergeCheckBinding, error) {
	input, err := LoadJSON[AtlasMergeCheckBindingInput](inputPath)
	if err != nil {
		return AtlasMergeCheckBinding{}, err
	}
	if err := ValidateAtlasMergeCheckBindingInput(input); err != nil {
		return AtlasMergeCheckBinding{}, err
	}
	binding := summarizeMergeCheckBinding(input.Rows)
	binding.Schema = AtlasMergeCheckBindingContract
	binding.Status = "required_checks_bound"
	binding.SourceInputPath = publicArtifactRef(inputPath)
	binding.SourceInputDigest = digestValue(input)
	binding.SchedulesWork = false
	binding.ExecutesWork = false
	binding.ApprovesWork = false
	binding.ClaimsAuthorityAdvance = false
	binding.RSIRemainsDenied = true
	if err := ValidateAtlasMergeCheckBinding(binding); err != nil {
		return AtlasMergeCheckBinding{}, err
	}
	return binding, nil
}

func ValidateAtlasMergeCheckBindingInput(input AtlasMergeCheckBindingInput) error {
	var errs []string
	requireContract(&errs, "merge_check_binding_input", input.Schema, AtlasMergeCheckBindingInputContract)
	if input.Status != "recorded" {
		errs = append(errs, "status must be recorded")
	}
	if len(input.Rows) == 0 {
		errs = append(errs, "rows must not be empty")
	}
	validateMergeCheckBindingRows(&errs, input.Rows, false)
	validateNoAuthorityEffects(&errs, input.SchedulesWork, input.ExecutesWork, input.ApprovesWork, input.ClaimsAuthorityAdvance, input.RSIRemainsDenied)
	return joinErrors(errs)
}

func ValidateAtlasMergeCheckBinding(binding AtlasMergeCheckBinding) error {
	var errs []string
	requireContract(&errs, "merge_check_binding", binding.Schema, AtlasMergeCheckBindingContract)
	if binding.Status != "required_checks_bound" {
		errs = append(errs, "status must be required_checks_bound")
	}
	requireField(&errs, "source_input_path", binding.SourceInputPath)
	checkPublicPath(&errs, "source_input_path", binding.SourceInputPath, true)
	if !digestPattern.MatchString(binding.SourceInputDigest) {
		errs = append(errs, "source_input_digest must be sha256 digest")
	}
	if len(binding.Rows) == 0 {
		errs = append(errs, "rows must not be empty")
	}
	validateMergeCheckBindingRows(&errs, binding.Rows, true)
	expected := summarizeMergeCheckBinding(binding.Rows)
	if binding.RowCount != expected.RowCount {
		errs = append(errs, "row_count must match rows length")
	}
	if binding.PassedRequiredCheckRows != expected.PassedRequiredCheckRows {
		errs = append(errs, "passed_required_check_rows must match rows")
	}
	if binding.UnboundMergeCommits != expected.UnboundMergeCommits {
		errs = append(errs, "unbound_merge_commits must match rows")
	}
	validateNoAuthorityEffects(&errs, binding.SchedulesWork, binding.ExecutesWork, binding.ApprovesWork, binding.ClaimsAuthorityAdvance, binding.RSIRemainsDenied)
	return joinErrors(errs)
}

func summarizeMergeCheckBinding(rows []AtlasMergeCheckBindingRow) AtlasMergeCheckBinding {
	binding := AtlasMergeCheckBinding{
		RowCount: len(rows),
		Rows:     make([]AtlasMergeCheckBindingRow, 0, len(rows)),
	}
	for _, row := range rows {
		row.RequiredChecksStatus = "failed"
		if row.RequiredCheckCount > 0 && row.PassedRequiredCheckCount == row.RequiredCheckCount {
			row.RequiredChecksStatus = "passed"
			binding.PassedRequiredCheckRows++
		}
		row.MergeCommitBound = len(row.MergeCommit) == 40
		if !row.MergeCommitBound {
			binding.UnboundMergeCommits++
		}
		binding.Rows = append(binding.Rows, row)
	}
	return binding
}

func validateMergeCheckBindingRows(errs *[]string, rows []AtlasMergeCheckBindingRow, requireBindingFields bool) {
	seenPRs := map[int]bool{}
	previousPR := 0
	for i, row := range rows {
		prefix := fmt.Sprintf("rows[%d]", i)
		requireField(errs, prefix+".node_id", row.NodeID)
		checkPublicPath(errs, prefix+".node_id", row.NodeID, true)
		if row.PRNumber <= 0 {
			*errs = append(*errs, prefix+".pr_number must be greater than zero")
		}
		if seenPRs[row.PRNumber] {
			*errs = append(*errs, "rows pr_number values must be unique")
		}
		seenPRs[row.PRNumber] = true
		if previousPR != 0 && row.PRNumber < previousPR {
			*errs = append(*errs, "rows must be sorted by pr_number")
		}
		previousPR = row.PRNumber
		requireField(errs, prefix+".merge_commit", row.MergeCommit)
		if len(row.MergeCommit) != 40 {
			*errs = append(*errs, prefix+".merge_commit must be a 40 character commit hash")
		}
		if row.RequiredCheckCount <= 0 {
			*errs = append(*errs, prefix+".required_check_count must be greater than zero")
		}
		if row.PassedRequiredCheckCount < 0 || row.PassedRequiredCheckCount > row.RequiredCheckCount {
			*errs = append(*errs, prefix+".passed_required_check_count must be between zero and required_check_count")
		}
		if requireBindingFields {
			expectedStatus := "failed"
			if row.PassedRequiredCheckCount == row.RequiredCheckCount {
				expectedStatus = "passed"
			}
			if row.RequiredChecksStatus != expectedStatus {
				*errs = append(*errs, prefix+".required_checks_status must match passed check count")
			}
			if !row.MergeCommitBound {
				*errs = append(*errs, prefix+".merge_commit_bound must be true")
			}
		}
	}
}
