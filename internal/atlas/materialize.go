package atlas

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func MaterializeFactoryDryRun(task FactoryTask, outDir string) (FactoryMaterialization, error) {
	if err := ValidateFactoryTask(task); err != nil {
		return FactoryMaterialization{}, err
	}
	if strings.TrimSpace(outDir) == "" {
		return FactoryMaterialization{}, fmt.Errorf("--out is required")
	}
	files := []string{
		"README.md",
		"task.json",
		"verification.txt",
		"materialization.json",
		filepath.ToSlash(filepath.Join("evidence", "README.md")),
		filepath.ToSlash(filepath.Join("context", "README.md")),
	}
	taskData, err := json.Marshal(task)
	if err != nil {
		return FactoryMaterialization{}, err
	}
	materialization := FactoryMaterialization{
		ContractVersion: FactoryMaterializationContract,
		TaskID:          task.ID,
		Mode:            "dry_run",
		OutputRoot:      "selected-output-directory",
		Files:           files,
		ExecutesWork:    false,
		SchedulesWork:   false,
		TaskDigest:      DigestBytes(taskData),
	}
	if err := ValidateFactoryMaterialization(materialization); err != nil {
		return FactoryMaterialization{}, err
	}
	if err := os.MkdirAll(filepath.Join(outDir, "evidence"), 0o755); err != nil {
		return FactoryMaterialization{}, err
	}
	if err := os.MkdirAll(filepath.Join(outDir, "context"), 0o755); err != nil {
		return FactoryMaterialization{}, err
	}
	writes := map[string]string{
		"README.md":                            taskReadme(task),
		"verification.txt":                     strings.Join(task.Verification, "\n") + "\n",
		filepath.Join("evidence", "README.md"): "# Evidence\n\nPlace required evidence here after governed execution. This dry-run skeleton does not execute work.\n",
		filepath.Join("context", "README.md"):  "# Context\n\nPlace bounded context packs here. Do not copy whole source repositories or private machine-local state.\n",
	}
	for rel, content := range writes {
		if err := writeTextInside(outDir, rel, content); err != nil {
			return FactoryMaterialization{}, err
		}
	}
	if err := WriteJSON(filepath.Join(outDir, "task.json"), task); err != nil {
		return FactoryMaterialization{}, err
	}
	if err := WriteJSON(filepath.Join(outDir, "materialization.json"), materialization); err != nil {
		return FactoryMaterialization{}, err
	}
	return materialization, nil
}

func writeTextInside(root, rel, content string) error {
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func taskReadme(task FactoryTask) string {
	return fmt.Sprintf(`# %s

Objective: %s

Target factory repo: %s
Factory folder: %s

This is an AO Atlas dry-run factory skeleton. It does not schedule, execute, approve, publish, push, tag, upload, or call providers.
`, task.ID, task.Objective, task.TargetFactoryRepo, task.FactoryFolder)
}
