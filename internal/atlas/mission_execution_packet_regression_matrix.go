package atlas

import "fmt"

func BuildAtlasExecutionPacketRegressionMatrix() (AtlasExecutionPacketRegressionMatrix, error) {
	cases := []AtlasExecutionPacketRegressionCase{
		{
			Name:                       "default_packet",
			PacketState:                "default",
			ExpectedStatus:             "denied",
			ProviderInvocationAllowed:  false,
			SilentChangedResultAllowed: false,
			ChangedResultClaimed:       false,
		},
		{
			Name:                       "malformed_packet",
			PacketState:                "malformed",
			ExpectedStatus:             "rejected",
			ProviderInvocationAllowed:  false,
			SilentChangedResultAllowed: false,
			ChangedResultClaimed:       false,
		},
	}
	matrix := AtlasExecutionPacketRegressionMatrix{
		Schema:                     AtlasExecutionPacketRegressionMatrixContract,
		Status:                     "execution_packet_regression_matrix_ready",
		Cases:                      cases,
		CaseCount:                  len(cases),
		ProviderInvocationAllowed:  false,
		SilentChangedResultAllowed: false,
		SchedulesWork:              false,
		ExecutesWork:               false,
		ApprovesWork:               false,
		ClaimsAuthorityAdvance:     false,
		RSIRemainsDenied:           true,
	}
	if err := ValidateAtlasExecutionPacketRegressionMatrix(matrix); err != nil {
		return AtlasExecutionPacketRegressionMatrix{}, err
	}
	return matrix, nil
}

func ValidateAtlasExecutionPacketRegressionMatrix(matrix AtlasExecutionPacketRegressionMatrix) error {
	var errs []string
	requireContract(&errs, "execution_packet_regression_matrix", matrix.Schema, AtlasExecutionPacketRegressionMatrixContract)
	if matrix.Status != "execution_packet_regression_matrix_ready" {
		errs = append(errs, "status must be execution_packet_regression_matrix_ready")
	}
	if matrix.CaseCount != len(matrix.Cases) {
		errs = append(errs, "case_count must match cases")
	}
	if matrix.CaseCount != 2 {
		errs = append(errs, "case_count must be 2")
	}
	if matrix.ProviderInvocationAllowed {
		errs = append(errs, "provider_invocation_allowed must be false")
	}
	if matrix.SilentChangedResultAllowed {
		errs = append(errs, "silent_changed_result_allowed must be false")
	}
	seen := map[string]bool{}
	for i, tc := range matrix.Cases {
		prefix := fmt.Sprintf("cases[%d]", i)
		requireField(&errs, prefix+".name", tc.Name)
		requireField(&errs, prefix+".packet_state", tc.PacketState)
		requireField(&errs, prefix+".expected_status", tc.ExpectedStatus)
		if tc.ProviderInvocationAllowed {
			errs = append(errs, prefix+".provider_invocation_allowed must be false")
		}
		if tc.SilentChangedResultAllowed {
			errs = append(errs, prefix+".silent_changed_result_allowed must be false")
		}
		if tc.ChangedResultClaimed {
			errs = append(errs, prefix+".changed_result_claimed must be false")
		}
		seen[tc.PacketState] = true
	}
	for _, state := range []string{"default", "malformed"} {
		if !seen[state] {
			errs = append(errs, "cases must include "+state+" packet")
		}
	}
	validateNoAuthorityEffects(&errs, matrix.SchedulesWork, matrix.ExecutesWork, matrix.ApprovesWork, matrix.ClaimsAuthorityAdvance, matrix.RSIRemainsDenied)
	return joinErrors(errs)
}
