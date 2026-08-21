package atlas

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVerifyRunLinkEvidenceRejectsWrongRootIdentity(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "node.json"), []byte("{\"status\":\"passed\"}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	link, err := BuildEvidenceBoundRunLink(
		"task-01",
		"completed",
		map[string]string{"node": "node.json"},
		root,
		"evidence-root-a",
	)
	if err != nil {
		t.Fatal(err)
	}
	err = VerifyRunLinkEvidence(link, root, "evidence-root-b")
	if err == nil || !strings.Contains(err.Error(), "evidence root identity mismatch") {
		t.Fatalf("expected wrong evidence-root identity rejection, got %v", err)
	}
}

func TestBuildEvidenceBoundRunLinkRejectsMissingEvidence(t *testing.T) {
	root := t.TempDir()
	_, err := BuildEvidenceBoundRunLink(
		"task-01",
		"completed",
		map[string]string{"node": "missing.json"},
		root,
		"evidence-root",
	)
	if err == nil || !os.IsNotExist(unwrapPathError(err)) {
		t.Fatalf("expected missing evidence rejection, got %v", err)
	}
}

func TestBuildEvidenceBoundRunLinkRejectsSymlinkEvidence(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target.json")
	if err := os.WriteFile(target, []byte("{\"status\":\"passed\"}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	requireTestSymlink(t, "target.json", filepath.Join(root, "link.json"))
	_, err := BuildEvidenceBoundRunLink(
		"task-01",
		"completed",
		map[string]string{"node": "link.json"},
		root,
		"evidence-root",
	)
	if err == nil || !strings.Contains(err.Error(), "must not be a symlink") {
		t.Fatalf("expected symlink evidence rejection, got %v", err)
	}
}

func TestBuildEvidenceBoundRunLinkRejectsTraversal(t *testing.T) {
	root := t.TempDir()
	_, err := BuildEvidenceBoundRunLink(
		"task-01",
		"completed",
		map[string]string{"node": "../outside.json"},
		root,
		"evidence-root",
	)
	if err == nil || !strings.Contains(err.Error(), "must remain beneath evidence root") {
		t.Fatalf("expected traversal rejection, got %v", err)
	}
}

func unwrapPathError(err error) error {
	for err != nil {
		pathErr, ok := err.(*os.PathError)
		if ok {
			return pathErr.Err
		}
		unwrapper, ok := err.(interface{ Unwrap() error })
		if !ok {
			return err
		}
		err = unwrapper.Unwrap()
	}
	return nil
}
