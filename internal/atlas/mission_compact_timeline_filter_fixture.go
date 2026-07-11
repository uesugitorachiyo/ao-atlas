package atlas

import "fmt"

func BuildAtlasCompactTimelineFilterFixture() (AtlasCompactTimelineFilterFixture, error) {
	filters := []AtlasCompactTimelineFilter{
		{Name: "stale", RecordStatus: "stale", DistinguishesFrom: "pending"},
		{Name: "duplicate", RecordStatus: "duplicate", DistinguishesFrom: "completed"},
		{Name: "pending", RecordStatus: "pending", DistinguishesFrom: "completed"},
		{Name: "denied", RecordStatus: "denied", DistinguishesFrom: "completed"},
		{Name: "completed", RecordStatus: "completed", DistinguishesFrom: "pending"},
	}
	fixture := AtlasCompactTimelineFilterFixture{
		Schema:                        AtlasCompactTimelineFilterFixtureContract,
		Status:                        "compact_timeline_filter_ready",
		Filters:                       filters,
		FilterCount:                   len(filters),
		StaleRecordsDistinguished:     true,
		DuplicateRecordsDistinguished: true,
		PendingRecordsDistinguished:   true,
		DeniedRecordsDistinguished:    true,
		CompletedRecordsDistinguished: true,
		SchedulesWork:                 false,
		ExecutesWork:                  false,
		ApprovesWork:                  false,
		ClaimsAuthorityAdvance:        false,
		RSIRemainsDenied:              true,
	}
	if err := ValidateAtlasCompactTimelineFilterFixture(fixture); err != nil {
		return AtlasCompactTimelineFilterFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasCompactTimelineFilterFixture(fixture AtlasCompactTimelineFilterFixture) error {
	var errs []string
	requireContract(&errs, "compact_timeline_filter_fixture", fixture.Schema, AtlasCompactTimelineFilterFixtureContract)
	if fixture.Status != "compact_timeline_filter_ready" {
		errs = append(errs, "status must be compact_timeline_filter_ready")
	}
	if fixture.FilterCount != len(fixture.Filters) {
		errs = append(errs, "filter_count must match filters")
	}
	if fixture.FilterCount != 5 {
		errs = append(errs, "filter_count must be 5")
	}
	if !fixture.StaleRecordsDistinguished {
		errs = append(errs, "stale_records_distinguished must be true")
	}
	if !fixture.DuplicateRecordsDistinguished {
		errs = append(errs, "duplicate_records_distinguished must be true")
	}
	if !fixture.PendingRecordsDistinguished {
		errs = append(errs, "pending_records_distinguished must be true")
	}
	if !fixture.DeniedRecordsDistinguished {
		errs = append(errs, "denied_records_distinguished must be true")
	}
	if !fixture.CompletedRecordsDistinguished {
		errs = append(errs, "completed_records_distinguished must be true")
	}
	seen := map[string]bool{}
	for i, filter := range fixture.Filters {
		prefix := fmt.Sprintf("filters[%d]", i)
		requireField(&errs, prefix+".name", filter.Name)
		requireField(&errs, prefix+".record_status", filter.RecordStatus)
		requireField(&errs, prefix+".distinguishes_from", filter.DistinguishesFrom)
		seen[filter.RecordStatus] = true
	}
	for _, required := range []string{"stale", "duplicate", "pending", "denied", "completed"} {
		if !seen[required] {
			errs = append(errs, "filters must include "+required)
		}
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}
