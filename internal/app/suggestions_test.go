package app

import (
	"testing"

	gitpkg "github.com/laerciocrestani/gitai/internal/git"
	prpkg "github.com/laerciocrestani/gitai/internal/pr"
)

func overview() *gitpkg.Overview {
	return &gitpkg.Overview{
		Upstream:           "origin/fix/foo",
		CommitsAheadOfBase: 2,
	}
}

func stepCommands(steps []NextStep) []string {
	out := make([]string, len(steps))
	for i, s := range steps {
		out[i] = s.Command
	}
	return out
}

func TestBuildNextSteps_dirtyWithUpstream(t *testing.T) {
	o := overview()
	o.Untracked = 1

	steps := buildNextSteps(o, nil, true)
	cmds := stepCommands(steps)

	if len(cmds) != 2 || cmds[0] != "gitai commit" || cmds[1] != "gitai push" {
		t.Fatalf("commands = %v, want [gitai commit gitai push]", cmds)
	}
	if steps[1].Note == "" {
		t.Fatal("expected commit note on gitai push")
	}
	if !steps[1].Muted {
		t.Fatal("expected gitai push to be muted")
	}
}

func TestBuildNextSteps_dirtyWithoutUpstream(t *testing.T) {
	o := overview()
	o.Upstream = ""
	o.Modified = 1

	steps := buildNextSteps(o, nil, true)
	if len(steps) != 1 || steps[0].Command != "gitai commit" {
		t.Fatalf("steps = %+v, want gitai commit only", steps)
	}
}

func TestBuildNextSteps_cleanAheadOfRemote(t *testing.T) {
	o := overview()
	o.Ahead = 2

	steps := buildNextSteps(o, nil, true)
	if len(steps) != 2 {
		t.Fatalf("len = %d, want 2", len(steps))
	}
	if steps[0].Command != "gitai push" || steps[0].Note != "" {
		t.Fatalf("first step = %+v, want gitai push without note", steps[0])
	}
	if steps[1].Command != "gitai pr" {
		t.Fatalf("second step = %+v, want gitai pr", steps[1])
	}
}

func TestBuildNextSteps_existingPR(t *testing.T) {
	o := overview()
	pr := &prpkg.PRView{Number: 87}

	steps := buildNextSteps(o, pr, true)
	cmds := stepCommands(steps)
	if len(cmds) != 1 || cmds[0] != "gitai pr view" {
		t.Fatalf("commands = %v, want [gitai pr view]", cmds)
	}
}

func TestBuildNextSteps_dirtyWithExistingPR(t *testing.T) {
	o := overview()
	o.Untracked = 1
	pr := &prpkg.PRView{Number: 87}

	steps := buildNextSteps(o, pr, true)
	cmds := stepCommands(steps)
	if len(cmds) != 3 || cmds[0] != "gitai commit" || cmds[1] != "gitai push" || cmds[2] != "gitai pr view" {
		t.Fatalf("commands = %v, want [gitai commit gitai push gitai pr view]", cmds)
	}
	if !steps[1].Muted {
		t.Fatal("expected gitai push to be muted")
	}
}

func TestBuildNextSteps_notConfigured(t *testing.T) {
	steps := buildNextSteps(overview(), nil, false)
	if steps[0].Command != "gitai config" {
		t.Fatalf("first step = %q, want gitai config", steps[0].Command)
	}
}
