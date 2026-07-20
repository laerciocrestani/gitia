package desktop

import (
	"path/filepath"
	"testing"
)

func TestStatusLabel(t *testing.T) {
	if got := statusLabel(false, 0, 0, 0); got != "clean" {
		t.Fatalf("clean: got %q", got)
	}
	if got := statusLabel(true, 1, 2, 3); got != "1 staged · 2 modified · 3 untracked" {
		t.Fatalf("dirty: got %q", got)
	}
}

func TestLoadDashboard_openbenchRepo(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	dash, err := LoadDashboard(root)
	if err != nil {
		t.Fatal(err)
	}
	if dash.RepoName == "" || dash.Branch == "" {
		t.Fatalf("incomplete dashboard: %+v", dash)
	}
	if dash.StatusLabel == "" {
		t.Fatal("missing status label")
	}
}
