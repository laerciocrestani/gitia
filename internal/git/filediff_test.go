package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestDiffFile_modified(t *testing.T) {
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "test")

	path := filepath.Join(dir, "hello.txt")
	if err := os.WriteFile(path, []byte("line1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "hello.txt")
	runGit(t, dir, "commit", "-m", "init")

	if err := os.WriteFile(path, []byte("line1\nline2\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	repo, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	diff, err := repo.DiffFile("hello.txt")
	if err != nil {
		t.Fatal(err)
	}
	if diff.Before != "line1\n" {
		t.Fatalf("before = %q", diff.Before)
	}
	if diff.After != "line1\nline2\n" {
		t.Fatalf("after = %q", diff.After)
	}
	if diff.Binary {
		t.Fatal("expected text file")
	}
}

func TestDiffFile_rejectsTraversal(t *testing.T) {
	dir := t.TempDir()
	runGit(t, dir, "init")
	repo, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.DiffFile("../outside"); err == nil {
		t.Fatal("expected error")
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}
