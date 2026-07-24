package gha

import "testing"

func TestIsDefaultBranch(t *testing.T) {
	if !IsDefaultBranch("main", "main") {
		t.Fatal("main/main")
	}
	if !IsDefaultBranch("MAIN", "main") {
		t.Fatal("case")
	}
	if !IsDefaultBranch("master", "") {
		t.Fatal("master alias")
	}
	if IsDefaultBranch("feature/x", "main") {
		t.Fatal("feature should not match")
	}
}

func TestDefaultBranchWarning(t *testing.T) {
	w := DefaultBranchWarning("main", "main")
	if w == "" || !containsFold(w, "CI") {
		t.Fatalf("warning=%q", w)
	}
	if DefaultBranchWarning("feat", "main") != "" {
		t.Fatal("expected empty")
	}
}

func containsFold(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		indexFold(s, sub) >= 0)
}

func indexFold(s, sub string) int {
	ls, lsub := toLower(s), toLower(sub)
	for i := 0; i+len(lsub) <= len(ls); i++ {
		if ls[i:i+len(lsub)] == lsub {
			return i
		}
	}
	return -1
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}
