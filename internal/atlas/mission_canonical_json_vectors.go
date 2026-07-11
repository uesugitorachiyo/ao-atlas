package atlas

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func BuildAtlasCanonicalJSONVectors() (AtlasCanonicalJSONVectors, error) {
	vectors := []AtlasCanonicalJSONVector{
		{ID: "mission-record-minimal", RecordClass: "mission", CanonicalJSON: `{"id":"mission-1","status":"active"}`, Consumers: []string{"go", "rust"}},
		{ID: "approval-record-minimal", RecordClass: "approval", CanonicalJSON: `{"approved":false,"ticket":"ticket-1"}`, Consumers: []string{"go", "rust"}},
		{ID: "policy-record-minimal", RecordClass: "policy", CanonicalJSON: `{"policy_digest":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","version":"v1"}`, Consumers: []string{"go", "rust"}},
		{ID: "rollback-record-minimal", RecordClass: "rollback", CanonicalJSON: `{"rollback_available":true,"target":"change-1"}`, Consumers: []string{"go", "rust"}},
		{ID: "readback-record-minimal", RecordClass: "readback", CanonicalJSON: `{"completed_nodes":1,"final_response_allowed":false}`, Consumers: []string{"go", "rust"}},
	}
	for i := range vectors {
		vectors[i].Digest = canonicalJSONVectorDigest(vectors[i].CanonicalJSON)
	}
	fixture := AtlasCanonicalJSONVectors{
		Schema:                 AtlasCanonicalJSONVectorsContract,
		Status:                 "canonical_json_vectors_ready",
		DigestAlgorithm:        "sha256.canonical-json.v1",
		Languages:              []string{"go", "rust"},
		Vectors:                vectors,
		SchedulesWork:          false,
		ExecutesWork:           false,
		ApprovesWork:           false,
		ClaimsAuthorityAdvance: false,
		RSIRemainsDenied:       true,
	}
	fixture.VectorCount = len(fixture.Vectors)
	fixture.LanguageCount = len(fixture.Languages)
	if err := ValidateAtlasCanonicalJSONVectors(fixture); err != nil {
		return AtlasCanonicalJSONVectors{}, err
	}
	return fixture, nil
}

func canonicalJSONVectorDigest(canonicalJSON string) string {
	sum := sha256.Sum256([]byte(canonicalJSON))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func ValidateAtlasCanonicalJSONVectors(fixture AtlasCanonicalJSONVectors) error {
	var errs []string
	requireContract(&errs, "canonical_json_vectors", fixture.Schema, AtlasCanonicalJSONVectorsContract)
	if fixture.Status != "canonical_json_vectors_ready" {
		errs = append(errs, "status must be canonical_json_vectors_ready")
	}
	if fixture.DigestAlgorithm != "sha256.canonical-json.v1" {
		errs = append(errs, "digest_algorithm must be sha256.canonical-json.v1")
	}
	if fixture.VectorCount != len(fixture.Vectors) {
		errs = append(errs, "vector_count must match vectors")
	}
	if fixture.LanguageCount != len(fixture.Languages) {
		errs = append(errs, "language_count must match languages")
	}
	if fixture.VectorCount != 5 {
		errs = append(errs, "vector_count must be 5")
	}
	for i, lang := range fixture.Languages {
		requireField(&errs, fmt.Sprintf("languages[%d]", i), lang)
	}
	for i, vector := range fixture.Vectors {
		prefix := fmt.Sprintf("vectors[%d]", i)
		requireField(&errs, prefix+".id", vector.ID)
		requireField(&errs, prefix+".record_class", vector.RecordClass)
		requireField(&errs, prefix+".canonical_json", vector.CanonicalJSON)
		validateRejectedTicketDigest(&errs, prefix+".digest", vector.Digest)
		if vector.Digest != canonicalJSONVectorDigest(vector.CanonicalJSON) {
			errs = append(errs, prefix+".digest must match canonical_json")
		}
		if len(vector.Consumers) != fixture.LanguageCount {
			errs = append(errs, prefix+".consumers must cover every language")
		}
		for j, consumer := range vector.Consumers {
			requireField(&errs, fmt.Sprintf("%s.consumers[%d]", prefix, j), consumer)
		}
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}
