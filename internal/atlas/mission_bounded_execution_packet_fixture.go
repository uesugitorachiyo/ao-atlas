package atlas

import "fmt"

func BuildAtlasBoundedExecutionPacketFixture() (AtlasBoundedExecutionPacketFixture, error) {
	required := []string{"isolated_worktree", "exact_digest_approval", "verified_diff", "reviewed_pr_evidence", "rollback_receipt"}
	fixture := AtlasBoundedExecutionPacketFixture{
		Schema:                      AtlasBoundedExecutionPacketFixtureContract,
		Status:                      "bounded_execution_packet_ready",
		RequiredEvidence:            required,
		RequiredEvidenceCount:       len(required),
		IsolatedWorktreeRequired:    true,
		ExactDigestApprovalRequired: true,
		VerifiedDiffRequired:        true,
		ReviewedPREvidenceRequired:  true,
		RollbackReceiptRequired:     true,
		SchedulesWork:               false,
		ExecutesWork:                false,
		ApprovesWork:                false,
		ClaimsAuthorityAdvance:      false,
		RSIRemainsDenied:            true,
	}
	if err := ValidateAtlasBoundedExecutionPacketFixture(fixture); err != nil {
		return AtlasBoundedExecutionPacketFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasBoundedExecutionPacketFixture(fixture AtlasBoundedExecutionPacketFixture) error {
	var errs []string
	requireContract(&errs, "bounded_execution_packet_fixture", fixture.Schema, AtlasBoundedExecutionPacketFixtureContract)
	if fixture.Status != "bounded_execution_packet_ready" {
		errs = append(errs, "status must be bounded_execution_packet_ready")
	}
	if fixture.RequiredEvidenceCount != len(fixture.RequiredEvidence) {
		errs = append(errs, "required_evidence_count must match required_evidence")
	}
	if fixture.RequiredEvidenceCount != 5 || !containsAll(fixture.RequiredEvidence, []string{"isolated_worktree", "exact_digest_approval", "verified_diff", "reviewed_pr_evidence", "rollback_receipt"}) {
		errs = append(errs, "required_evidence must cover isolated_worktree, exact_digest_approval, verified_diff, reviewed_pr_evidence, and rollback_receipt")
	}
	for i, evidence := range fixture.RequiredEvidence {
		requireField(&errs, fmt.Sprintf("required_evidence[%d]", i), evidence)
	}
	if !fixture.IsolatedWorktreeRequired {
		errs = append(errs, "isolated_worktree_required must be true")
	}
	if !fixture.ExactDigestApprovalRequired {
		errs = append(errs, "exact_digest_approval_required must be true")
	}
	if !fixture.VerifiedDiffRequired {
		errs = append(errs, "verified_diff_required must be true")
	}
	if !fixture.ReviewedPREvidenceRequired {
		errs = append(errs, "reviewed_pr_evidence_required must be true")
	}
	if !fixture.RollbackReceiptRequired {
		errs = append(errs, "rollback_receipt_required must be true")
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}
