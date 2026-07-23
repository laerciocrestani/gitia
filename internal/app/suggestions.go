package app

import (
	gitpkg "github.com/laerciocrestani/openbench/internal/git"
	dockerpkg "github.com/laerciocrestani/openbench/internal/docker"
	prpkg "github.com/laerciocrestani/openbench/internal/pr"
)

// NextStep representa uma ação sugerida ao usuário (CLI ou TUI).
type NextStep struct {
	Command string
	Note    string
	Plain   bool
	Muted   bool
}

func buildNextSteps(o *gitpkg.Overview, pr *prpkg.PRView, docker *dockerpkg.Overview, configured bool) []NextStep {
	var steps []NextStep

	if !configured {
		steps = append(steps, NextStep{Command: "ob config"})
	}

	if docker != nil && docker.CanUp() && !dockerpkg.HasRunningContainers(docker.Containers) {
		steps = append(steps, NextStep{Command: "ob docker up", Note: "subir ambiente local"})
	}
	if docker != nil && docker.Available && !docker.DaemonRunning {
		steps = append(steps, NextStep{Command: "ob docker status", Note: "Docker daemon parado"})
	}

	switch {
	case o.IsDirty() && o.Upstream != "":
		steps = append(steps, NextStep{Command: "ob commit"})
		steps = append(steps, NextStep{
			Command: "ob push",
			Note:    "inclui commit automático com IA",
			Muted:   true,
		})
	case o.IsDirty():
		steps = append(steps, NextStep{Command: "ob commit"})
	case o.Ahead > 0:
		steps = append(steps, NextStep{Command: "ob push"})
	}

	if pr == nil && o.CommitsAheadOfBase > 0 && !o.IsDirty() {
		steps = append(steps, NextStep{Command: "ob pr"})
	}
	if pr != nil {
		steps = append(steps, NextStep{Command: "ob pr view"})
	}

	if len(o.Stashes) > 0 {
		steps = append(steps, NextStep{Command: "git stash pop"})
	}
	if o.Behind > 0 || o.BaseBehind > 0 {
		steps = append(steps, NextStep{Command: "ob sync"})
	}

	if len(steps) == 0 && !o.IsDirty() {
		steps = append(steps, NextStep{Plain: true, Command: "working tree clean"})
	}

	return steps
}
