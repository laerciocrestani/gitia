package app

import (
	"testing"
)

func TestLoadWorkspaceSnapshot_requiresGitRepo(t *testing.T) {
	t.Chdir(t.TempDir())
	_, err := LoadWorkspaceSnapshot()
	if err == nil {
		t.Fatal("expected error outside git repo")
	}
}
