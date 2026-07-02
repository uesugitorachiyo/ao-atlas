package atlas

import (
	"strings"
	"time"
)

func validateBlueprintAuthorization(auth BlueprintBuildAuthorization, rules BlueprintCandidateRules, packDigest string) ([]string, []string) {
	missing := []string{}
	blockers := []string{}
	if auth.SchemaVersion != "ao.blueprint.build-authorization.v0.1" {
		missing = append(missing, "build_authorization_schema")
		blockers = append(blockers, "Blueprint authorization schema must be ao.blueprint.build-authorization.v0.1")
	}
	if auth.Status != "ready" {
		missing = append(missing, "build_authorization_ready")
		blockers = append(blockers, "Blueprint authorization status must be ready")
	}
	if auth.Score < 100 {
		missing = append(missing, "build_authorization_score")
		blockers = append(blockers, "Blueprint authorization score must be 100")
	}
	if !auth.ApprovedByUser {
		missing = append(missing, "user_approval")
		blockers = append(blockers, "Blueprint authorization must be approved by user")
	}
	if len(auth.BlockingAssumptions) > 0 {
		missing = append(missing, "blocking_assumptions")
		blockers = append(blockers, "Blueprint authorization must not contain blocking assumptions")
	}
	if auth.ProjectID != rules.ProjectID {
		missing = append(missing, "authorization_scope")
		blockers = append(blockers, "Blueprint authorization project_id must match candidate rules")
	}
	if auth.MutationClass != "" && auth.MutationClass != rules.MutationClass {
		missing = append(missing, "authorization_mutation_class")
		blockers = append(blockers, "Blueprint authorization mutation_class must match candidate rules")
	}
	if auth.BlueprintPackDigest != "" && auth.BlueprintPackDigest != packDigest {
		missing = append(missing, "blueprint_pack_digest")
		blockers = append(blockers, "Blueprint authorization blueprint_pack_digest is stale or mismatched")
	}
	if auth.NextAllowedAction != "" && !oneOf(auth.NextAllowedAction, "ao-atlas", "ao-foundry") {
		missing = append(missing, "next_allowed_action")
		blockers = append(blockers, "Blueprint authorization next_allowed_action must route to AO Atlas or legacy AO Foundry intake")
	}
	if rules.MutationClass == "low_risk_code" {
		if !strings.Contains(auth.Scope, "low_risk_code") || !strings.Contains(auth.Scope, "dry_run") {
			missing = append(missing, "authorization_scope")
			blockers = append(blockers, "low_risk_code authorization must be scoped to low_risk_code dry_run")
		}
	}
	if strings.TrimSpace(auth.ExpiresAtUTC) != "" {
		expires, err := time.Parse(time.RFC3339, auth.ExpiresAtUTC)
		if err != nil || !expires.After(time.Now().UTC()) {
			missing = append(missing, "build_authorization_freshness")
			blockers = append(blockers, "Blueprint authorization is stale or has invalid expires_at_utc")
		}
	}
	return uniqueStrings(missing), uniqueStrings(blockers)
}

func mutationModelIncludes(model MutationClassModel, class string) bool {
	for _, definition := range model.Classes {
		if definition.Name == class {
			return true
		}
	}
	return false
}
