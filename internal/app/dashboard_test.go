package app

import (
	"strings"
	"testing"

	"github.com/laerciocrestani/openbench/internal/config"
	gitpkg "github.com/laerciocrestani/openbench/internal/git"
)

func TestBuildChangeSummary(t *testing.T) {
	o := &gitpkg.Overview{
		FileChanges: []gitpkg.FileChange{
			{Path: "internal/app/foo.go", Insertions: 10, Deletions: 2},
			{Path: "internal/app/foo_test.go", Insertions: 5, Deletions: 0},
			{Path: "web/app.js", Insertions: 3, Deletions: 1},
		},
	}

	s := BuildChangeSummary(o)
	if s.FileCount != 3 {
		t.Fatalf("file count = %d", s.FileCount)
	}
	if s.Insertions != 18 || s.Deletions != 3 {
		t.Fatalf("stats = +%d -%d", s.Insertions, s.Deletions)
	}
	if s.Languages["Go"] != 1 || s.Languages["Tests"] != 1 || s.Languages["JS"] != 1 {
		t.Fatalf("languages = %#v", s.Languages)
	}
	if s.DominantDir != "internal" {
		t.Fatalf("dominant dir = %q", s.DominantDir)
	}
}

func TestBuildTUINextAction_commit(t *testing.T) {
	snap := &WorkspaceSnapshot{
		Overview: &gitpkg.Overview{
			Modified: 1,
			FileChanges: []gitpkg.FileChange{
				{Path: "internal/ui/header.go", Status: "modified"},
			},
		},
		NextSteps: []NextStep{{Command: "ob commit"}},
		Config:    &config.Config{APIKey: "test", Provider: config.ProviderGemini},
	}

	action := BuildTUINextAction(snap)
	if action.Key != "c" {
		t.Fatalf("key = %q", action.Key)
	}
	if !strings.Contains(action.Message, "Commit") {
		t.Fatalf("message = %q", action.Message)
	}
}

func TestRepoDisplayName_prefersRootOverRemote(t *testing.T) {
	o := &gitpkg.Overview{
		Root:      "/Users/dev/openbench",
		RemoteURL: "https://github.com/user/gitai.git",
	}
	if got := repoDisplayName(o); got != "openbench" {
		t.Fatalf("repoDisplayName = %q want openbench", got)
	}
}

func TestBuildHeaderContext(t *testing.T) {
	snap := &WorkspaceSnapshot{
		Overview: &gitpkg.Overview{
			Branch:     "main",
			HeadHash:   "abc1234",
			BaseBranch: "main",
		},
		Config: &config.Config{APIKey: "key", Provider: config.ProviderGemini, Model: "gemini-2.5-flash-lite"},
	}
	ctx := BuildHeaderContext(snap)
	if ctx.Repo == "" || ctx.Branch != "main" || !ctx.AIReady {
		t.Fatalf("ctx = %#v", ctx)
	}
	if !ctx.OnBase {
		t.Fatal("expected on base branch")
	}
}

func TestCanPushAndPR(t *testing.T) {
	pushSnap := &WorkspaceSnapshot{Overview: &gitpkg.Overview{Ahead: 1}}
	if !CanPush(pushSnap) {
		t.Fatal("expected CanPush")
	}

	prSnap := &WorkspaceSnapshot{
		Overview:  &gitpkg.Overview{CommitsAheadOfBase: 2},
		HasGH:     true,
		ConfigErr: nil,
	}
	if !CanPR(prSnap) {
		t.Fatal("expected CanPR")
	}
}
