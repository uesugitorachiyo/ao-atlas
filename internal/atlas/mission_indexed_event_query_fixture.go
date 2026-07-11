package atlas

import "fmt"

func BuildAtlasIndexedEventQueryFixture() (AtlasIndexedEventQueryFixture, error) {
	eventTypes := []string{"mission", "policy", "approval", "rollback", "readback"}
	fixture := AtlasIndexedEventQueryFixture{
		Schema:                 AtlasIndexedEventQueryFixtureContract,
		Status:                 "indexed_event_query_ready",
		EventTypes:             eventTypes,
		EventTypeCount:         len(eventTypes),
		MigrationRequired:      true,
		QueryIndexRequired:     true,
		QueryFields:            []string{"mission_id", "event_type", "created_at", "source_digest"},
		QueryFieldCount:        4,
		SchedulesWork:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
		ClaimsAuthorityAdvance: false,
		RSIRemainsDenied:       true,
	}
	if err := ValidateAtlasIndexedEventQueryFixture(fixture); err != nil {
		return AtlasIndexedEventQueryFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasIndexedEventQueryFixture(fixture AtlasIndexedEventQueryFixture) error {
	var errs []string
	requireContract(&errs, "indexed_event_query_fixture", fixture.Schema, AtlasIndexedEventQueryFixtureContract)
	if fixture.Status != "indexed_event_query_ready" {
		errs = append(errs, "status must be indexed_event_query_ready")
	}
	if fixture.EventTypeCount != len(fixture.EventTypes) {
		errs = append(errs, "event_type_count must match event_types")
	}
	if fixture.EventTypeCount != 5 {
		errs = append(errs, "event_type_count must be 5")
	}
	for _, required := range []string{"mission", "policy", "approval", "rollback", "readback"} {
		if !containsStringValue(fixture.EventTypes, required) {
			errs = append(errs, "event_types must include "+required)
		}
	}
	if !fixture.MigrationRequired {
		errs = append(errs, "migration_required must be true")
	}
	if !fixture.QueryIndexRequired {
		errs = append(errs, "query_index_required must be true")
	}
	if fixture.QueryFieldCount != len(fixture.QueryFields) {
		errs = append(errs, "query_field_count must match query_fields")
	}
	for i, field := range fixture.QueryFields {
		requireField(&errs, fmt.Sprintf("query_fields[%d]", i), field)
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}
