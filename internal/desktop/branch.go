package desktop

import (
	"fmt"
	"strings"

	gitpkg "github.com/laerciocrestani/openbench/internal/git"
)

// BranchView is one row in the desktop branch picker.
type BranchView struct {
	Name     string `json:"name"`
	Current  bool   `json:"current"`
	Upstream string `json:"upstream"`
	Ahead    int    `json:"ahead"`
	Behind   int    `json:"behind"`
}

// ListBranches returns local branches for the open project.
func ListBranches(projectPath string) ([]BranchView, error) {
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
	branches, err := repo.ListBranches()
	if err != nil {
		return nil, err
	}
	out := make([]BranchView, 0, len(branches))
	for _, b := range branches {
		out = append(out, BranchView{
			Name:     b.Name,
			Current:  b.Current,
			Upstream: b.Upstream,
			Ahead:    b.Ahead,
			Behind:   b.Behind,
		})
	}
	return out, nil
}

// CheckoutBranch switches to name and returns a refreshed dashboard.
func CheckoutBranch(projectPath, name string) (*Dashboard, error) {
	if strings.TrimSpace(projectPath) == "" {
		return nil, fmt.Errorf("no project open")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("nome da branch vazio")
	}
	repo, err := gitpkg.Open(projectPath)
	if err != nil {
		return nil, err
	}
	if err := repo.Checkout(name); err != nil {
		return nil, err
	}
	return LoadDashboard(projectPath)
}
