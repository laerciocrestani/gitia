package app

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestLoadDiff_workingTree(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	t.Chdir(dir)

	writeFile(t, dir, "foo.txt", "hello\n")
	runGit(t, dir, "add", "foo.txt")

	snap, err := LoadWorkspaceSnapshot()
	if err != nil {
		t.Fatal(err)
	}

	title, diff, err := LoadDiff(snap)
	if err != nil {
		t.Fatal(err)
	}
	if title != "Working tree (staged + unstaged)" {
		t.Fatalf("title = %q", title)
	}
	if diff == "" {
		t.Fatal("expected non-empty diff")
	}
}

func initGitRepo(t *testing.T, dir string) {
	t.Helper()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "t@example.com")
	runGit(t, dir, "config", "user.name", "Test")
	writeFile(t, dir, "README.md", "init\n")
	runGit(t, dir, "add", "README.md")
	runGit(t, dir, "commit", "-m", "init")
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
