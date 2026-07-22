package desktop

import (
	"fmt"
	"strings"

	gitpkg "github.com/laerciocrestani/openbench/internal/git"
)

// PushOutcome is returned after pushing the current branch.
type PushOutcome struct {
	Path    string `json:"path"`
	Branch  string `json:"branch"`
	Message string `json:"message"`
}

// PushCurrentBranch pushes existing commits on HEAD to origin (no commit/AI).
// Use when the branch is already ahead and the user only needs git push.
func PushCurrentBranch(projectPath string) (*PushOutcome, error) {
	if strings.TrimSpace(projectPath) == "" {
		return nil, fmt.Errorf("no project open")
	}
	repo, err := gitpkg.Open(projectPath)
	if err != nil {
		return nil, err
	}
	if err := repo.IsRepo(); err != nil {
		return nil, fmt.Errorf("diretório atual não é um repositório git")
	}

	branch, err := repo.CurrentBranch()
	if err != nil {
		return nil, err
	}
	if branch == "" || branch == "HEAD" {
		return nil, fmt.Errorf("checkout em detached HEAD — faça checkout de uma branch antes do push")
	}

	overview, err := repo.Overview("")
	if err != nil {
		return nil, err
	}
	hasUpstream := strings.TrimSpace(overview.Upstream) != ""
	if hasUpstream && overview.Ahead <= 0 {
		return nil, fmt.Errorf("nada para enviar — branch já está sincronizada com o remote (↑0)")
	}

	if err := repo.Push(); err != nil {
		return nil, err
	}
	msg := fmt.Sprintf("push de %s ok", branch)
	if hasUpstream && overview.Ahead > 0 {
		msg = fmt.Sprintf("push de %s (↑%d) ok", branch, overview.Ahead)
	}
	return &PushOutcome{
		Path:    projectPath,
		Branch:  branch,
		Message: msg,
	}, nil
}
