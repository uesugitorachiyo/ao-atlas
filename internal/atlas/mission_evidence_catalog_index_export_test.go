package atlas

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestMonth6EvidenceCatalogIndexExport(t *testing.T) {
	root := repoRoot(t)
	recordedPath := filepath.Join(root, "examples", "valid", "evidence-catalog-index-export.json")
	outPath := filepath.Join(t.TempDir(), "evidence-catalog-index-export.json")

	var out bytes.Buffer
	code := Run([]string{
		"mission", "recommendations", "evidence-catalog-index-export",
		"--out", outPath,
	}, &out, &out)
	if code != 0 {
		t.Fatalf("evidence-catalog-index-export command failed: %s", out.String())
	}
	for _, want := range []string{
		"status=evidence_catalog_index_export_ready",
		"index_entry_count=3",
		"bulk_campaign_artifacts_cataloged=true",
		"source_artifacts_retained=true",
		"uploads_artifacts=false",
		"deletes_source_artifacts=false",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("evidence catalog index export output missing %q: %s", want, out.String())
		}
	}

	recorded := mustLoadJSON[map[string]any](t, recordedPath)
	generated := mustLoadJSON[map[string]any](t, outPath)
	if digestValue(generated) != digestValue(recorded) {
		t.Fatalf("evidence catalog index export changed\nwant %s\ngot  %s", digestValue(recorded), digestValue(generated))
	}
	if generated["bulk_campaign_artifacts_cataloged"] != true ||
		generated["source_artifacts_retained"] != true ||
		generated["uploads_artifacts"] != false ||
		generated["deletes_source_artifacts"] != false ||
		generated["executes_work"] != false ||
		generated["claims_authority_advance"] != false ||
		generated["rsi_remains_denied"] != true {
		t.Fatalf("evidence catalog index export lost safety state: %#v", generated)
	}
}

func TestMonth6EvidenceCatalogIndexExportUsesTypedValidator(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "examples", "valid", "evidence-catalog-index-export.json")
	validator, err := validateRecommendationEvidenceTypedFile(path, "ao.atlas.evidence-catalog-index-export.v0.1")
	if err != nil {
		t.Fatal(err)
	}
	if validator != "typed:evidence-catalog-index-export" {
		t.Fatalf("expected typed evidence catalog index export validator, got %s", validator)
	}
}

func TestMonth6EvidenceCatalogIndexExportRejectsUpload(t *testing.T) {
	fixture, err := BuildAtlasEvidenceCatalogIndexExport()
	if err != nil {
		t.Fatal(err)
	}
	fixture.UploadsArtifacts = true
	if err := ValidateAtlasEvidenceCatalogIndexExport(fixture); err == nil || !strings.Contains(err.Error(), "uploads_artifacts must be false") {
		t.Fatalf("expected upload rejection, got %v", err)
	}
}
