package atlas

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

func BuildAtlasBlueprintCanonicalPreservationFixture() (AtlasBlueprintCanonicalPreservationFixture, error) {
	root, err := findRepoRoot()
	if err != nil {
		return AtlasBlueprintCanonicalPreservationFixture{}, err
	}
	workspaceRootRef := "examples/valid/blueprint-import-low-risk-code/blueprint-pack"
	packPath := filepath.Join(root, filepath.FromSlash(workspaceRootRef))
	outDir, err := os.MkdirTemp("", "atlas-blueprint-canonical-preservation-*")
	if err != nil {
		return AtlasBlueprintCanonicalPreservationFixture{}, err
	}
	defer os.RemoveAll(outDir)

	result, err := BuildBlueprintImport(BlueprintImportPaths{
		PackPath:            packPath,
		AuthorizationPath:   filepath.Join(root, "examples", "valid", "blueprint-import-low-risk-code", "build-authorization.json"),
		InstancePath:        filepath.Join(root, "examples", "valid", "stack-instance.json"),
		MutationClassesPath: filepath.Join(root, "examples", "valid", "mutation-classes.json"),
		OutDir:              outDir,
	})
	if err != nil {
		return AtlasBlueprintCanonicalPreservationFixture{}, err
	}
	foundryImport, err := LoadJSON[FoundryImport](filepath.Join(outDir, "foundry-import", "foundry-import.json"))
	if err != nil {
		return AtlasBlueprintCanonicalPreservationFixture{}, err
	}
	packDigest, err := digestDirectory(packPath)
	if err != nil {
		return AtlasBlueprintCanonicalPreservationFixture{}, err
	}
	foundryPackDigest := sourceDigestForRef(foundryImport.SourceArtifacts, workspaceRootRef)
	files, err := blueprintCanonicalFiles(packPath, workspaceRootRef)
	if err != nil {
		return AtlasBlueprintCanonicalPreservationFixture{}, err
	}
	fixture := AtlasBlueprintCanonicalPreservationFixture{
		Schema:                           AtlasBlueprintCanonicalPreservationFixtureContract,
		Status:                           "preserved",
		WorkspaceRootRef:                 workspaceRootRef,
		BlueprintPackDigest:              packDigest,
		ImportRecordBlueprintPackDigest:  result.Record.BlueprintPack.Digest,
		FoundrySourceBlueprintPackDigest: foundryPackDigest,
		DigestPreserved:                  packDigest == result.Record.BlueprintPack.Digest && packDigest == foundryPackDigest,
		CanonicalBytesPreserved:          true,
		CanonicalFiles:                   files,
		SchedulesWork:                    false,
		ExecutesWork:                     false,
		ApprovesWork:                     false,
		ClaimsAuthorityAdvance:           false,
		RSIRemainsDenied:                 true,
	}
	fixture.CanonicalFileCount = len(fixture.CanonicalFiles)
	if err := ValidateAtlasBlueprintCanonicalPreservationFixture(fixture); err != nil {
		return AtlasBlueprintCanonicalPreservationFixture{}, err
	}
	return fixture, nil
}

func sourceDigestForRef(refs []SourceRef, ref string) string {
	for _, source := range refs {
		if source.Ref == ref {
			return source.Digest
		}
	}
	return ""
}

func blueprintCanonicalFiles(packPath string, workspaceRootRef string) ([]AtlasBlueprintCanonicalPreservationFile, error) {
	files := []AtlasBlueprintCanonicalPreservationFile{}
	err := filepath.WalkDir(packPath, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(packPath, path)
		if err != nil {
			return err
		}
		digest, err := digestFile(path)
		if err != nil {
			return err
		}
		files = append(files, AtlasBlueprintCanonicalPreservationFile{
			Path:      filepath.ToSlash(filepath.Join(workspaceRootRef, rel)),
			Digest:    digest,
			SizeBytes: info.Size(),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})
	return files, nil
}

func ValidateAtlasBlueprintCanonicalPreservationFixture(fixture AtlasBlueprintCanonicalPreservationFixture) error {
	var errs []string
	requireContract(&errs, "blueprint_canonical_preservation_fixture", fixture.Schema, AtlasBlueprintCanonicalPreservationFixtureContract)
	if fixture.Status != "preserved" {
		errs = append(errs, "status must be preserved")
	}
	requireField(&errs, "workspace_root_ref", fixture.WorkspaceRootRef)
	validateRejectedTicketDigest(&errs, "blueprint_pack_digest", fixture.BlueprintPackDigest)
	if fixture.BlueprintPackDigest != fixture.ImportRecordBlueprintPackDigest {
		errs = append(errs, "import_record_blueprint_pack_digest must match blueprint_pack_digest")
	}
	if fixture.BlueprintPackDigest != fixture.FoundrySourceBlueprintPackDigest {
		errs = append(errs, "foundry_source_blueprint_pack_digest must match blueprint_pack_digest")
	}
	if !fixture.DigestPreserved {
		errs = append(errs, "digest_preserved must be true")
	}
	if !fixture.CanonicalBytesPreserved {
		errs = append(errs, "canonical_bytes_preserved must be true")
	}
	if fixture.CanonicalFileCount != len(fixture.CanonicalFiles) {
		errs = append(errs, "canonical_file_count must match canonical_files")
	}
	if fixture.CanonicalFileCount == 0 {
		errs = append(errs, "canonical_file_count must be positive")
	}
	for i, file := range fixture.CanonicalFiles {
		prefix := fmt.Sprintf("canonical_files[%d]", i)
		requireField(&errs, prefix+".path", file.Path)
		validateRejectedTicketDigest(&errs, prefix+".digest", file.Digest)
		if file.SizeBytes <= 0 {
			errs = append(errs, prefix+".size_bytes must be positive")
		}
	}
	validateNoAuthorityEffects(&errs, fixture.SchedulesWork, fixture.ExecutesWork, fixture.ApprovesWork, fixture.ClaimsAuthorityAdvance, fixture.RSIRemainsDenied)
	return joinErrors(errs)
}
