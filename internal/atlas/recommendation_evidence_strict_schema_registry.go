package atlas

import "fmt"

var strictRecommendationEvidenceGenericSchemas = map[string]struct{}{
	"ao.atlas.consolidation-candidate-record.v0.1":         {},
	"ao.atlas.consolidation-checkpoint-bundle.v0.1":        {},
	"ao.atlas.consolidation-command-readback.v0.1":         {},
	"ao.atlas.consolidation-evidence-catalog-plan.v0.1":    {},
	"ao.atlas.consolidation-evidence-volume-baseline.v0.1": {},
	"ao.atlas.consolidation-foundry-import.v0.1":           {},
	"ao.atlas.consolidation-implementation-evidence.v0.1":  {},
	"ao.atlas.consolidation-node-gate.v0.1":                {},
	"ao.atlas.consolidation-promoter-no-promotion.v0.1":    {},
	"ao.atlas.consolidation-rollback-record.v0.1":          {},
	"ao.atlas.consolidation-sentinel-public-safety.v0.1":   {},
	"ao.atlas.consolidation-test-evidence.v0.1":            {},
	"ao.atlas.consolidation-verification-evidence.v0.1":    {},
}

func validateRecommendationEvidenceTypedFileStrict(path, schema string) (string, error) {
	validator, err := validateRecommendationEvidenceTypedFile(path, schema)
	if err != nil {
		return "strict:" + validator, err
	}
	if validator == "generic:schema-marker" {
		if _, ok := strictRecommendationEvidenceGenericSchemas[schema]; !ok {
			return "strict:unknown-schema", fmt.Errorf("unknown recommendation evidence schema %q", schema)
		}
		return "strict:generic:schema-marker", nil
	}
	return "strict:" + validator, nil
}
