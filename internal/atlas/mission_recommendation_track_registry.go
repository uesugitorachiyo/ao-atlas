package atlas

import "fmt"

func DefaultAtlasRecommendationTrackRegistry() (AtlasRecommendationTrackRegistry, error) {
	registry := AtlasRecommendationTrackRegistry{
		Schema:                         AtlasRecommendationTrackRegistryContract,
		Status:                         "ready",
		DefaultTrack:                   "feature_depth",
		SaturatedFeatureDepthNextTrack: "refactoring",
		PriorityOrder:                  []string{"refactoring", "feature_depth", "rsi_boundary_hardening"},
		Tracks: []AtlasRecommendationTrackRegistryEntry{
			{
				Track:               "feature_depth",
				Rank:                1,
				Status:              "saturates_before_repeat_export",
				NextWhenCompleted:   "refactoring",
				AuthorityEffect:     "none",
				SchedulesWork:       false,
				MutatesRepositories: false,
			},
			{
				Track:               "refactoring",
				Rank:                2,
				Status:              "recommended_after_feature_depth_saturation",
				NextWhenCompleted:   "rsi_boundary_hardening",
				AuthorityEffect:     "none",
				SchedulesWork:       false,
				MutatesRepositories: false,
			},
			{
				Track:               "rsi_boundary_hardening",
				Rank:                3,
				Status:              "boundary_hardening_only_denied",
				NextWhenCompleted:   "none",
				AuthorityEffect:     "denial_preserved",
				SchedulesWork:       false,
				MutatesRepositories: false,
			},
		},
		FeatureDepthStatus:     "saturated_completed_routes_to_refactoring",
		RefactoringStatus:      "recommended_next",
		RSITrackStatus:         "boundary_hardening_only_denied",
		NoPromotionRequested:   true,
		PromotionGranted:       false,
		ClaimsAuthorityAdvance: false,
		RSIRemainsDenied:       true,
		SafeToExecute:          false,
		SchedulesWork:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
		MutatesRepositories:    false,
	}
	if err := ValidateAtlasRecommendationTrackRegistry(registry); err != nil {
		return AtlasRecommendationTrackRegistry{}, err
	}
	return registry, nil
}

func priorityOrderForRecommendedTrack(registry AtlasRecommendationTrackRegistry, recommendedTrack string) []string {
	if recommendedTrack == registry.SaturatedFeatureDepthNextTrack {
		return append([]string(nil), registry.PriorityOrder...)
	}
	order := []string{recommendedTrack}
	for _, track := range registry.PriorityOrder {
		if track != recommendedTrack {
			order = append(order, track)
		}
	}
	return order
}

func ValidateAtlasRecommendationTrackRegistry(registry AtlasRecommendationTrackRegistry) error {
	var errs []string
	requireContract(&errs, "recommendation_track_registry", registry.Schema, AtlasRecommendationTrackRegistryContract)
	if registry.Status != "ready" {
		errs = append(errs, "status must be ready")
	}
	if registry.DefaultTrack != "feature_depth" {
		errs = append(errs, "default_track must be feature_depth")
	}
	if registry.SaturatedFeatureDepthNextTrack != "refactoring" {
		errs = append(errs, "saturated_feature_depth_next_track must be refactoring")
	}
	requireList(&errs, "priority_order", registry.PriorityOrder)
	if len(registry.PriorityOrder) != 3 ||
		registry.PriorityOrder[0] != "refactoring" ||
		registry.PriorityOrder[1] != "feature_depth" ||
		registry.PriorityOrder[2] != "rsi_boundary_hardening" {
		errs = append(errs, "priority_order must be refactoring, feature_depth, rsi_boundary_hardening")
	}
	if len(registry.Tracks) != 3 {
		errs = append(errs, "tracks must include 3 entries")
	}
	seen := map[string]bool{}
	for _, track := range registry.Tracks {
		if !oneOf(track.Track, "feature_depth", "refactoring", "rsi_boundary_hardening") {
			errs = append(errs, fmt.Sprintf("track %q is invalid", track.Track))
		}
		if track.Rank <= 0 {
			errs = append(errs, fmt.Sprintf("track %s rank must be positive", track.Track))
		}
		requireField(&errs, "track.status", track.Status)
		requireField(&errs, "track.next_when_completed", track.NextWhenCompleted)
		requireField(&errs, "track.authority_effect", track.AuthorityEffect)
		if track.SchedulesWork {
			errs = append(errs, fmt.Sprintf("track %s schedules_work must be false", track.Track))
		}
		if track.MutatesRepositories {
			errs = append(errs, fmt.Sprintf("track %s mutates_repositories must be false", track.Track))
		}
		seen[track.Track] = true
	}
	for _, required := range []string{"feature_depth", "refactoring", "rsi_boundary_hardening"} {
		if !seen[required] {
			errs = append(errs, "tracks missing "+required)
		}
	}
	requireField(&errs, "feature_depth_status", registry.FeatureDepthStatus)
	requireField(&errs, "refactoring_status", registry.RefactoringStatus)
	if registry.RSITrackStatus != "boundary_hardening_only_denied" {
		errs = append(errs, "rsi_track_status must be boundary_hardening_only_denied")
	}
	if !registry.NoPromotionRequested {
		errs = append(errs, "no_promotion_requested must be true")
	}
	if registry.PromotionGranted {
		errs = append(errs, "promotion_granted must be false")
	}
	if registry.ClaimsAuthorityAdvance {
		errs = append(errs, "claims_authority_advance must be false")
	}
	if !registry.RSIRemainsDenied {
		errs = append(errs, "rsi_remains_denied must be true")
	}
	if registry.SafeToExecute {
		errs = append(errs, "safe_to_execute must be false")
	}
	if registry.SchedulesWork {
		errs = append(errs, "schedules_work must be false")
	}
	if registry.ExecutesWork {
		errs = append(errs, "executes_work must be false")
	}
	if registry.ApprovesWork {
		errs = append(errs, "approves_work must be false")
	}
	if registry.MutatesRepositories {
		errs = append(errs, "mutates_repositories must be false")
	}
	return joinErrors(errs)
}
