package git

import (
	"os/exec"
	"strings"
	"testing"
)

func TestWrapGitAuthError(t *testing.T) {
	err := wrapGitAuthError([]string{"push"}, "Username for 'https://github.com':", exec.ErrNotFound)
	if err == nil {
		t.Fatal("expected error")
	}
	got := err.Error()
	if !strings.Contains(got, "gh auth setup-git") || !strings.Contains(got, "Username for") {
		t.Fatalf("unexpected message: %q", got)
	}
}

func TestWrapGitAuthErrorPassthrough(t *testing.T) {
	err := wrapGitAuthError([]string{"status"}, "fatal: not a git repository", exec.ErrNotFound)
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), "gh auth setup-git") {
		t.Fatalf("should not mention gh auth: %q", err)
	}
}
