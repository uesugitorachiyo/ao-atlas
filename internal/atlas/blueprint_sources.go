package atlas

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func readJSONIfPossible(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, value)
}

func digestFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return DigestBytes(data), nil
}

func digestDirectory(root string) (string, error) {
	hash := sha256.New()
	cleanRoot := filepath.Clean(root)
	err := filepath.WalkDir(cleanRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if shouldSkipBlueprintDigestDir(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(cleanRoot, path)
		if err != nil {
			return err
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		hash.Write([]byte(filepath.ToSlash(rel)))
		hash.Write([]byte{0})
		hash.Write(body)
		hash.Write([]byte{0})
		return nil
	})
	if err != nil {
		return "", err
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), nil
}

func digestValue(value any) string {
	data, _ := json.Marshal(value)
	return DigestBytes(data)
}

func digestPersistedJSON(value any) (string, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "", err
	}
	data = append(data, '\n')
	return DigestBytes(data), nil
}

func shouldSkipBlueprintDigestDir(name string) bool {
	switch name {
	case ".git", "tmp", "target", ".idea", ".vscode", "__pycache__":
		return true
	default:
		return false
	}
}

func contextDigestKey(ref string) string {
	switch filepath.ToSlash(ref) {
	case "implementation-spec.md":
		return "implementation_spec"
	case "quality-profile.md":
		return "quality_profile"
	case "candidate-rules.json":
		return "candidate_rules"
	default:
		return ""
	}
}

func publicArtifactRef(path string) string {
	if strings.TrimSpace(path) == "" {
		return ""
	}
	clean := filepath.Clean(path)
	abs, err := filepath.Abs(clean)
	if err == nil {
		if root, rootErr := findRepoRoot(); rootErr == nil {
			if rel, relErr := filepath.Rel(root, abs); relErr == nil && !strings.HasPrefix(rel, "..") {
				return filepath.ToSlash(rel)
			}
		}
	}
	return filepath.ToSlash(filepath.Join("excluded", "local-artifacts", filepath.Base(clean)))
}

func findRepoRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(cwd, "go.mod")); err == nil {
			return cwd, nil
		}
		parent := filepath.Dir(cwd)
		if parent == cwd {
			return "", errors.New("repo root not found")
		}
		cwd = parent
	}
}

func copyStringMap(values map[string]string) map[string]string {
	result := map[string]string{}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		result[key] = values[key]
	}
	return result
}
