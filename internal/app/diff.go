package app

import (
	"fmt"
	"strings"

	gitpkg "github.com/laerciocrestani/gitai/internal/git"
)

// LoadDiff retorna diff do working tree ou da branch vs base, com título descritivo.
func LoadDiff(snap *WorkspaceSnapshot) (title, diff string, err error) {
	if snap == nil || snap.Overview == nil {
		return "", "", fmt.Errorf("snapshot inválido")
	}

	repo, err := gitpkg.New()
	if err != nil {
		return "", "", err
	}

	base := snap.Overview.BaseBranch
	if base == "" {
		base = "main"
	}

	if snap.Overview.IsDirty() {
		working, err := repo.DiffForCommit()
		if err != nil {
			return "", "", err
		}
		if strings.TrimSpace(working) != "" {
			return "Working tree (staged + unstaged)", working, nil
		}
	}

	resolved, err := repo.ResolveBase(base)
	if err != nil {
		return "", "", err
	}

	branchDiff, err := repo.DiffBranch(resolved)
	if err != nil {
		return "", "", err
	}
	if strings.TrimSpace(branchDiff) == "" {
		return "", "", fmt.Errorf("nenhum diff disponível")
	}

	label := fmt.Sprintf("Branch vs %s", strings.TrimPrefix(resolved, "origin/"))
	return label, branchDiff, nil
}
