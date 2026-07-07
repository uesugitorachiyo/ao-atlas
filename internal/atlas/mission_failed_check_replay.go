package atlas

import "fmt"

func BuildAtlasFailedCheckReplayFixture(inputPath string) (AtlasFailedCheckReplayFixture, error) {
	input, err := LoadJSON[AtlasFailedCheckReplayInput](inputPath)
	if err != nil {
		return AtlasFailedCheckReplayFixture{}, err
	}
	if err := ValidateAtlasFailedCheckReplayInput(input); err != nil {
		return AtlasFailedCheckReplayFixture{}, err
	}
	fixture := summarizeFailedCheckReplay(input.Cases)
	fixture.Schema = AtlasFailedCheckReplayFixtureContract
	fixture.Status = "replay_recorded"
	fixture.SourceInputPath = publicArtifactRef(inputPath)
	fixture.SourceInputDigest = digestValue(input)
	fixture.SchedulesWork = false
	fixture.ExecutesWork = false
	fixture.ApprovesWork = false
	fixture.ClaimsAuthorityAdvance = false
	fixture.RSIRemainsDenied = true
	if err := ValidateAtlasFailedCheckReplayFixture(fixture); err != nil {
		return AtlasFailedCheckReplayFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasFailedCheckReplayInput(input AtlasFailedCheckReplayInput) error {
	var errs []string
	requireContract(&errs, "failed_check_replay_input", input.Schema, AtlasFailedCheckReplayInputContract)
	if input.Status != "recorded" {
		errs = append(errs, "status must be recorded")
	}
	if len(input.Cases) == 0 {
		errs = append(errs, "cases must not be empty")
	}
	validateFailedCheckReplayInputCases(&errs, input.Cases)
	validateNoAuthorityEffects(&errs, input.SchedulesWork, input.ExecutesWork, input.ApprovesWork, input.ClaimsAuthorityAdvance, input.RSIRemainsDenied)
	return joinErrors(errs)
}

func ValidateAtlasFailedCheckReplayFixture(fixture AtlasFailedCheckReplayFixture) error {
	var errs []string
	requireContract(&errs, "failed_check_replay_fixture", fixture.Schema, AtlasFailedCheckReplayFixtureContract)
	if fixture.Status != "replay_recorded" {
		errs = append(errs, "status must be replay_recorded")
	}
	requireField(&errs, "source_input_path", fixture.SourceInputPath)
	checkPublicPath(&errs, "source_input_path", fixture.SourceInputPath, true)
	if !digestPattern.MatchString(fixture.SourceInputDigest) {
		errs = append(errs, "source_input_digest must be sha256 digest")
	}
	if len(fixture.Cases) == 0 {
		errs = append(errs, "cases must not be empty")
	}
	validateFailedCheckReplayFixtureCases(&errs, fixture.Cases)
	expected := summarizeReplayFixtureCases(fixture.Cases)
	if fixture.CaseCount != expected.CaseCount {
		errs = append(errs, "case_count must match cases length")
	}
	if fixture.MergeDeniedCases != expected.MergeDeniedCases {
		errs = append(errs, "merge_denied_cases must match cases")
	}
	if fixture.RetryAllowedCases != expected.RetryAllowedCases {
		errs = append(errs, "retry_allowed_cases must match cases")
	}
	if fixture.SafeToMerge != expected.SafeToMerge {
		errs = append(errs, "safe_to_merge must match merge decisions")
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}

func summarizeFailedCheckReplay(cases []AtlasFailedCheckReplayCase) AtlasFailedCheckReplayFixture {
	fixtureCases := make([]AtlasFailedCheckReplayFixtureCase, 0, len(cases))
	for _, input := range cases {
		retryDecision := "manual_repair_required"
		reason := failureKindReason(input.FailureKind)
		if input.Retryable {
			retryDecision = "retry_allowed"
			reason = "failed check blocks merge; retry is allowed for retryable timeout evidence"
		}
		fixtureCases = append(fixtureCases, AtlasFailedCheckReplayFixtureCase{
			ID:            input.ID,
			CheckName:     input.CheckName,
			Platform:      input.Platform,
			CheckStatus:   input.CheckStatus,
			FailureKind:   input.FailureKind,
			RetryDecision: retryDecision,
			MergeDecision: "merge_denied",
			Reason:        reason,
		})
	}
	return summarizeReplayFixtureCases(fixtureCases)
}

func summarizeReplayFixtureCases(cases []AtlasFailedCheckReplayFixtureCase) AtlasFailedCheckReplayFixture {
	fixture := AtlasFailedCheckReplayFixture{
		CaseCount:   len(cases),
		SafeToMerge: true,
		Cases:       append([]AtlasFailedCheckReplayFixtureCase(nil), cases...),
	}
	for _, item := range cases {
		if item.MergeDecision == "merge_denied" {
			fixture.MergeDeniedCases++
			fixture.SafeToMerge = false
		}
		if item.RetryDecision == "retry_allowed" {
			fixture.RetryAllowedCases++
		}
	}
	return fixture
}

func failureKindReason(kind string) string {
	switch kind {
	case "public_safety":
		return "public safety failure blocks merge until repaired and verified"
	case "schema_validation":
		return "schema validation failure blocks merge until evidence is repaired"
	default:
		return "failed check blocks merge until repaired and verified"
	}
}

func validateFailedCheckReplayInputCases(errs *[]string, cases []AtlasFailedCheckReplayCase) {
	seen := map[string]bool{}
	for i, item := range cases {
		prefix := fmt.Sprintf("cases[%d]", i)
		requireField(errs, prefix+".id", item.ID)
		checkPublicPath(errs, prefix+".id", item.ID, true)
		if seen[item.ID] {
			*errs = append(*errs, "cases ids must be unique")
		}
		seen[item.ID] = true
		requireField(errs, prefix+".check_name", item.CheckName)
		checkPublicPath(errs, prefix+".check_name", item.CheckName, true)
		requireField(errs, prefix+".platform", item.Platform)
		checkPublicPath(errs, prefix+".platform", item.Platform, true)
		if item.CheckStatus != "failed" {
			*errs = append(*errs, prefix+".check_status must be failed")
		}
		if !oneOf(item.FailureKind, "timeout", "public_safety", "schema_validation") {
			*errs = append(*errs, prefix+".failure_kind must be timeout, public_safety, or schema_validation")
		}
	}
}

func validateFailedCheckReplayFixtureCases(errs *[]string, cases []AtlasFailedCheckReplayFixtureCase) {
	seen := map[string]bool{}
	for i, item := range cases {
		prefix := fmt.Sprintf("cases[%d]", i)
		requireField(errs, prefix+".id", item.ID)
		checkPublicPath(errs, prefix+".id", item.ID, true)
		if seen[item.ID] {
			*errs = append(*errs, "cases ids must be unique")
		}
		seen[item.ID] = true
		requireField(errs, prefix+".check_name", item.CheckName)
		checkPublicPath(errs, prefix+".check_name", item.CheckName, true)
		requireField(errs, prefix+".platform", item.Platform)
		checkPublicPath(errs, prefix+".platform", item.Platform, true)
		if item.CheckStatus != "failed" {
			*errs = append(*errs, prefix+".check_status must be failed")
		}
		if !oneOf(item.FailureKind, "timeout", "public_safety", "schema_validation") {
			*errs = append(*errs, prefix+".failure_kind must be timeout, public_safety, or schema_validation")
		}
		if !oneOf(item.RetryDecision, "retry_allowed", "manual_repair_required") {
			*errs = append(*errs, prefix+".retry_decision must be retry_allowed or manual_repair_required")
		}
		if item.MergeDecision != "merge_denied" {
			*errs = append(*errs, prefix+".merge_decision must be merge_denied")
		}
		requireField(errs, prefix+".reason", item.Reason)
	}
}

func validateNoAuthorityEffects(errs *[]string, schedulesWork, executesWork, approvesWork, claimsAuthorityAdvance, rsiRemainsDenied bool) {
	if schedulesWork {
		*errs = append(*errs, "schedules_work must be false")
	}
	if executesWork {
		*errs = append(*errs, "executes_work must be false")
	}
	if approvesWork {
		*errs = append(*errs, "approves_work must be false")
	}
	if claimsAuthorityAdvance {
		*errs = append(*errs, "claims_authority_advance must be false")
	}
	if !rsiRemainsDenied {
		*errs = append(*errs, "rsi_remains_denied must be true")
	}
}
