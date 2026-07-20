package git

import "testing"

func TestFileChangeNeedsAdd(t *testing.T) {
	tests := []struct {
		status string
		want   bool
	}{
		{"untracked", true},
		{"modified", true},
		{"staged+modified", true},
		{"deleted", true},
		{"staged", false},
		{"new", false},
	}

	for _, tc := range tests {
		f := FileChange{Path: "x.go", Status: tc.status}
		if got := f.NeedsAdd(); got != tc.want {
			t.Errorf("NeedsAdd(%q) = %v, want %v", tc.status, got, tc.want)
		}
	}
}

func TestFilterAddable(t *testing.T) {
	changes := []FileChange{
		{Path: "a.go", Status: "untracked"},
		{Path: "b.go", Status: "staged"},
		{Path: "c.go", Status: "modified"},
	}
	got := FilterAddable(changes)
	if len(got) != 2 {
		t.Fatalf("FilterAddable len = %d, want 2", len(got))
	}
}

func TestSortByChurn(t *testing.T) {
	changes := []FileChange{
		{Path: "small.go", Insertions: 1, Deletions: 0},
		{Path: "big.go", Insertions: 10, Deletions: 5},
		{Path: "mid.go", Insertions: 3, Deletions: 3},
	}
	got := SortByChurn(changes)
	if got[0].Path != "big.go" || got[1].Path != "mid.go" || got[2].Path != "small.go" {
		t.Fatalf("order = %v, %v, %v", got[0].Path, got[1].Path, got[2].Path)
	}
}
