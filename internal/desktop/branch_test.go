package desktop

import (
	"path/filepath"
	"testing"
)

func TestListBranches_openbenchRepo(t *testing.T) {
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	branches, err := ListBranches(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(branches) == 0 {
		t.Fatal("expected at least one local branch")
	}
	var current int
	for _, b := range branches {
		if b.Name == "" {
			t.Fatal("branch with empty name")
		}
		if b.Current {
			current++
		}
	}
	if current != 1 {
		t.Fatalf("expected exactly one current branch, got %d", current)
	}
}

func TestListBranches_emptyPath(t *testing.T) {
	if _, err := ListBranches(""); err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestCheckoutBranch_emptyName(t *testing.T) {
	if _, err := CheckoutBranch("/tmp", ""); err == nil {
		t.Fatal("expected error for empty branch name")
	}
}
