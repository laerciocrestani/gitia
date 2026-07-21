package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStatusFingerprint_changesWithWorkingTree(t *testing.T) {
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "test")
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("one\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "a.txt")
	runGit(t, dir, "commit", "-m", "init")

	repo, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	fp1, err := repo.StatusFingerprint()
	if err != nil {
		t.Fatal(err)
	}
	fp2, err := repo.StatusFingerprint()
	if err != nil {
		t.Fatal(err)
	}
	if fp1 != fp2 {
		t.Fatalf("stable fingerprint expected, got %q vs %q", fp1, fp2)
	}

	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("two\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	fp3, err := repo.StatusFingerprint()
	if err != nil {
		t.Fatal(err)
	}
	if fp3 == fp1 {
		t.Fatal("expected fingerprint to change after edit")
	}
}

func TestStatusFingerprintAt(t *testing.T) {
	dir := t.TempDir()
	runGit(t, dir, "init")
	runGit(t, dir, "config", "user.email", "test@example.com")
	runGit(t, dir, "config", "user.name", "test")
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("one\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	runGit(t, dir, "add", "a.txt")
	runGit(t, dir, "commit", "-m", "init")

	fp, err := StatusFingerprintAt(dir)
	if err != nil {
		t.Fatal(err)
	}
	if fp == "" {
		t.Fatal("empty fingerprint")
	}
}
