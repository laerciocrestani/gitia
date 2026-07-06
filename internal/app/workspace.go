package app

import (
	"fmt"
	"os/exec"

	"github.com/laerciocrestani/gitai/internal/config"
	gitpkg "github.com/laerciocrestani/gitai/internal/git"
	prpkg "github.com/laerciocrestani/gitai/internal/pr"
)

// WorkspaceSnapshot agrega o estado read-only do repositório para o dashboard TUI.
type WorkspaceSnapshot struct {
	Overview  *gitpkg.Overview
	OpenPR    *prpkg.PRView
	Config    *config.Config
	ConfigErr error
	NextSteps []NextStep
	HasGH     bool
}

// LoadWorkspaceSnapshot coleta overview, PR aberto, config e próximos passos.
func LoadWorkspaceSnapshot() (*WorkspaceSnapshot, error) {
	repo, err := gitpkg.New()
	if err != nil {
		return nil, err
	}
	if err := repo.IsRepo(); err != nil {
		return nil, fmt.Errorf("diretório atual não é um repositório git")
	}

	baseBranch := "main"
	cfg, cfgErr := config.Load()
	if cfgErr == nil {
		baseBranch = cfg.BaseBranch
	}

	overview, err := repo.Overview(baseBranch)
	if err != nil {
		return nil, err
	}

	snap := &WorkspaceSnapshot{
		Overview:  overview,
		Config:    cfg,
		ConfigErr: cfgErr,
		HasGH:     hasGH(),
	}

	if snap.HasGH {
		client, err := prpkg.New()
		if err == nil {
			snap.OpenPR, _ = client.ViewCurrent()
		}
	}

	snap.NextSteps = buildNextSteps(overview, snap.OpenPR, cfgErr == nil)
	return snap, nil
}

func hasGH() bool {
	_, err := exec.LookPath("gh")
	return err == nil
}
