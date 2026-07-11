package atlas

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

func BuildAtlasCanonicalJSONVectorSmokeChecks(vectors AtlasCanonicalJSONVectors) (AtlasCanonicalJSONVectorSmokeChecks, error) {
	if err := ValidateAtlasCanonicalJSONVectors(vectors); err != nil {
		return AtlasCanonicalJSONVectorSmokeChecks{}, err
	}
	var checks []AtlasCanonicalJSONVectorSmokeCheck
	for _, vector := range vectors.Vectors {
		fields, err := canonicalJSONVectorFieldNames(vector.CanonicalJSON)
		if err != nil {
			return AtlasCanonicalJSONVectorSmokeChecks{}, err
		}
		for _, language := range vectors.Languages {
			checks = append(checks, AtlasCanonicalJSONVectorSmokeCheck{
				Language:       language,
				VectorID:       vector.ID,
				RecordClass:    vector.RecordClass,
				ExpectedDigest: vector.Digest,
				FieldCount:     len(fields),
				FieldNames:     append([]string(nil), fields...),
				Command:        canonicalJSONVectorSmokeCommand(language),
				DependencyFree: true,
			})
		}
	}
	fixture := AtlasCanonicalJSONVectorSmokeChecks{
		Schema:                 AtlasCanonicalJSONVectorSmokeChecksContract,
		Status:                 "smoke_checks_ready",
		SourceSchema:           vectors.Schema,
		DigestAlgorithm:        vectors.DigestAlgorithm,
		VectorCount:            vectors.VectorCount,
		LanguageCount:          vectors.LanguageCount,
		SmokeCheckCount:        len(checks),
		Languages:              append([]string(nil), vectors.Languages...),
		GoDependencyFree:       true,
		RustDependencyFree:     true,
		Checks:                 checks,
		SchedulesWork:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
		ClaimsAuthorityAdvance: false,
		RSIRemainsDenied:       true,
	}
	if err := ValidateAtlasCanonicalJSONVectorSmokeChecks(fixture); err != nil {
		return AtlasCanonicalJSONVectorSmokeChecks{}, err
	}
	return fixture, nil
}

func canonicalJSONVectorFieldNames(raw string) ([]string, error) {
	var object map[string]any
	if err := json.Unmarshal([]byte(raw), &object); err != nil {
		return nil, fmt.Errorf("canonical_json must parse as object: %w", err)
	}
	if len(object) == 0 {
		return nil, fmt.Errorf("canonical_json object must not be empty")
	}
	fields := make([]string, 0, len(object))
	for field := range object {
		fields = append(fields, field)
	}
	sort.Strings(fields)
	return fields, nil
}

func canonicalJSONVectorSmokeCommand(language string) string {
	switch language {
	case "go":
		return "go test dependency-free sha256 over canonical_json bytes"
	case "rust":
		return "rustc dependency-free sha256 over canonical_json bytes"
	default:
		return language + " dependency-free sha256 over canonical_json bytes"
	}
}

func ValidateAtlasCanonicalJSONVectorSmokeChecks(fixture AtlasCanonicalJSONVectorSmokeChecks) error {
	var errs []string
	requireContract(&errs, "canonical_json_vector_smoke_checks", fixture.Schema, AtlasCanonicalJSONVectorSmokeChecksContract)
	if fixture.Status != "smoke_checks_ready" {
		errs = append(errs, "status must be smoke_checks_ready")
	}
	requireContract(&errs, "source_schema", fixture.SourceSchema, AtlasCanonicalJSONVectorsContract)
	if fixture.DigestAlgorithm != "sha256.canonical-json.v1" {
		errs = append(errs, "digest_algorithm must be sha256.canonical-json.v1")
	}
	if fixture.VectorCount != 5 {
		errs = append(errs, "vector_count must be 5")
	}
	if fixture.LanguageCount != len(fixture.Languages) {
		errs = append(errs, "language_count must match languages")
	}
	if fixture.LanguageCount != 2 || !containsAll(fixture.Languages, []string{"go", "rust"}) {
		errs = append(errs, "languages must cover go and rust")
	}
	if fixture.SmokeCheckCount != len(fixture.Checks) {
		errs = append(errs, "smoke_check_count must match checks")
	}
	if fixture.SmokeCheckCount != fixture.VectorCount*fixture.LanguageCount {
		errs = append(errs, "smoke_check_count must cover every vector/language pair")
	}
	if !fixture.GoDependencyFree {
		errs = append(errs, "go_dependency_free must be true")
	}
	if !fixture.RustDependencyFree {
		errs = append(errs, "rust_dependency_free must be true")
	}
	seen := map[string]bool{}
	for i, check := range fixture.Checks {
		prefix := fmt.Sprintf("checks[%d]", i)
		requireField(&errs, prefix+".language", check.Language)
		requireField(&errs, prefix+".vector_id", check.VectorID)
		requireField(&errs, prefix+".record_class", check.RecordClass)
		validateRejectedTicketDigest(&errs, prefix+".expected_digest", check.ExpectedDigest)
		if !containsAll(fixture.Languages, []string{check.Language}) {
			errs = append(errs, prefix+".language must be declared in languages")
		}
		if check.FieldCount != len(check.FieldNames) {
			errs = append(errs, prefix+".field_count must match field_names")
		}
		if check.FieldCount == 0 {
			errs = append(errs, prefix+".field_count must be positive")
		}
		for j, field := range check.FieldNames {
			requireField(&errs, fmt.Sprintf("%s.field_names[%d]", prefix, j), field)
		}
		requireField(&errs, prefix+".command", check.Command)
		if !strings.Contains(check.Command, "dependency-free") {
			errs = append(errs, prefix+".command must declare dependency-free smoke check")
		}
		if !check.DependencyFree {
			errs = append(errs, prefix+".dependency_free must be true")
		}
		key := check.VectorID + "\x00" + check.Language
		if seen[key] {
			errs = append(errs, prefix+".vector/language pair must be unique")
		}
		seen[key] = true
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}

func containsAll(haystack, needles []string) bool {
	have := map[string]bool{}
	for _, value := range haystack {
		have[value] = true
	}
	for _, value := range needles {
		if !have[value] {
			return false
		}
	}
	return true
}
