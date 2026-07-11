package atlas

import "slices"

func BuildAtlasLocalPlatformFixture() (AtlasLocalPlatformFixture, error) {
	fixture := AtlasLocalPlatformFixture{
		Schema:                   AtlasLocalPlatformFixtureContract,
		Status:                   "local_platform_fixture_ready",
		Platforms:                []string{"darwin", "linux", "windows"},
		PlatformCount:            3,
		PathModes:                []string{"posix", "windows"},
		LineEndingModes:          []string{"lf", "crlf"},
		LineEndingModeCount:      2,
		DeterministicInstall:     true,
		PathNormalizationChecked: true,
		RollbackReceiptRequired:  true,
		LiveProviderCalls:        false,
		SchedulesWork:            false,
		ExecutesWork:             false,
		ApprovesWork:             false,
		ClaimsAuthorityAdvance:   false,
		RSIRemainsDenied:         true,
	}
	if err := ValidateAtlasLocalPlatformFixture(fixture); err != nil {
		return AtlasLocalPlatformFixture{}, err
	}
	return fixture, nil
}

func ValidateAtlasLocalPlatformFixture(fixture AtlasLocalPlatformFixture) error {
	var errs []string
	requireContract(&errs, "local_platform_fixture", fixture.Schema, AtlasLocalPlatformFixtureContract)
	if fixture.Status != "local_platform_fixture_ready" {
		errs = append(errs, "status must be local_platform_fixture_ready")
	}
	if fixture.PlatformCount != len(fixture.Platforms) {
		errs = append(errs, "platform_count must match platforms")
	}
	if fixture.PlatformCount != 3 {
		errs = append(errs, "platform_count must be 3")
	}
	for _, platform := range []string{"darwin", "linux", "windows"} {
		if !slices.Contains(fixture.Platforms, platform) {
			errs = append(errs, "platforms must include "+platform)
		}
	}
	for _, mode := range []string{"posix", "windows"} {
		if !slices.Contains(fixture.PathModes, mode) {
			errs = append(errs, "path_modes must include "+mode)
		}
	}
	if fixture.LineEndingModeCount != len(fixture.LineEndingModes) {
		errs = append(errs, "line_ending_mode_count must match line_ending_modes")
	}
	for _, mode := range []string{"lf", "crlf"} {
		if !slices.Contains(fixture.LineEndingModes, mode) {
			errs = append(errs, "line_ending_modes must include "+mode)
		}
	}
	if !fixture.DeterministicInstall {
		errs = append(errs, "deterministic_install must be true")
	}
	if !fixture.PathNormalizationChecked {
		errs = append(errs, "path_normalization_checked must be true")
	}
	if !fixture.RollbackReceiptRequired {
		errs = append(errs, "rollback_receipt_required must be true")
	}
	if fixture.LiveProviderCalls {
		errs = append(errs, "live_provider_calls must be false")
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}
