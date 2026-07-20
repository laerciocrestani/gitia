package git

import "sort"

// NeedsAdd reports whether the file has unstaged changes that git add can stage.
func (f FileChange) NeedsAdd() bool {
	switch f.Status {
	case "untracked", "modified", "staged+modified", "deleted", "changed", "renamed":
		return true
	default:
		return false
	}
}

// TotalChurn is insertions + deletions (used to rank files by change volume).
func (f FileChange) TotalChurn() int {
	return f.Insertions + f.Deletions
}

// FilterAddable returns files that can be staged with git add.
func FilterAddable(changes []FileChange) []FileChange {
	var out []FileChange
	for _, f := range changes {
		if f.NeedsAdd() {
			out = append(out, f)
		}
	}
	return out
}

// SortByChurn returns a copy of changes ordered by total +/- descending, then path.
func SortByChurn(changes []FileChange) []FileChange {
	sorted := make([]FileChange, len(changes))
	copy(sorted, changes)
	sort.Slice(sorted, func(i, j int) bool {
		ti := sorted[i].TotalChurn()
		tj := sorted[j].TotalChurn()
		if ti != tj {
			return ti > tj
		}
		return sorted[i].Path < sorted[j].Path
	})
	return sorted
}
