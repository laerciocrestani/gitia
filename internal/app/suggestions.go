package app

import (
	gitpkg "github.com/laerciocrestani/gitai/internal/git"
	prpkg "github.com/laerciocrestani/gitai/internal/pr"
)

// NextStep representa uma ação sugerida ao usuário (CLI ou TUI).
type NextStep struct {
	Command string
	Note    string
	Plain   bool
}

func buildNextSteps(o *gitpkg.Overview, pr *prpkg.PRView, configured bool) []NextStep {
	var steps []NextStep

	if !configured {
		steps = append(steps, NextStep{Command: "gitai config"})
	}

	switch {
	case o.IsDirty() && o.Upstream != "":
		steps = append(steps, NextStep{
			Command: "gitai push",
			Note:    "inclui commit automático com IA",
		})
	case o.IsDirty():
		steps = append(steps, NextStep{Command: "gitai commit"})
	case o.Ahead > 0:
		steps = append(steps, NextStep{Command: "gitai push"})
	}

	if pr == nil && o.CommitsAheadOfBase > 0 && !o.IsDirty() {
		steps = append(steps, NextStep{Command: "gitai pr"})
	}
	if pr != nil {
		steps = append(steps, NextStep{Command: "gitai pr view"})
	}

	if len(o.Stashes) > 0 {
		steps = append(steps, NextStep{Command: "git stash pop"})
	}
	if o.Behind > 0 {
		steps = append(steps, NextStep{Command: "gitai sync"})
	}

	if len(steps) == 0 && !o.IsDirty() {
		steps = append(steps, NextStep{Plain: true, Command: "working tree clean"})
	}

	return steps
}
