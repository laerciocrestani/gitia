package desktop

import (
	"fmt"

	gitpkg "github.com/laerciocrestani/openbench/internal/git"
)

// ChangedFileView is one row in the dashboard changed-files table.
type ChangedFileView struct {
	Path       string `json:"path"`
	Status     string `json:"status"`
	Insertions int    `json:"insertions"`
	Deletions  int    `json:"deletions"`
}

// FileDiffView is the side-panel mirror (before HEAD vs after working tree).
type FileDiffView struct {
	Path       string `json:"path"`
	Status     string `json:"status"`
	Before     string `json:"before"`
	After      string `json:"after"`
	Unified    string `json:"unified"`
	Binary     bool   `json:"binary"`
	Insertions int    `json:"insertions"`
	Deletions  int    `json:"deletions"`
}

// LoadFileDiff returns before/after content for a path in projectPath.
func LoadFileDiff(projectPath, filePath string) (*FileDiffView, error) {
	if projectPath == "" {
		return nil, fmt.Errorf("no project open")
	}
	repo, err := gitpkg.Open(projectPath)
	if err != nil {
		return nil, err
	}
	diff, err := repo.DiffFile(filePath)
	if err != nil {
		return nil, err
	}
	return &FileDiffView{
		Path:       diff.Path,
		Status:     diff.Status,
		Before:     diff.Before,
		After:      diff.After,
		Unified:    diff.Unified,
		Binary:     diff.Binary,
		Insertions: diff.Insertions,
		Deletions:  diff.Deletions,
	}, nil
}

func mapChangedFiles(changes []gitpkg.FileChange) []ChangedFileView {
	sorted := gitpkg.SortByChurn(changes)
	out := make([]ChangedFileView, 0, len(sorted))
	for _, c := range sorted {
		out = append(out, ChangedFileView{
			Path:       c.Path,
			Status:     c.Status,
			Insertions: c.Insertions,
			Deletions:  c.Deletions,
		})
	}
	return out
}
