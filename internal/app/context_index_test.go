package app

import (
	"testing"

	"github.com/laerciocrestani/openbench/internal/config"
	gitpkg "github.com/laerciocrestani/openbench/internal/git"
)

func TestBuildCommitContextIndex_nilWhenClean(t *testing.T) {
	if got := BuildCommitContextIndex(nil, nil); got != nil {
		t.Fatalf("expected nil, got %+v", got)
	}
	o := &gitpkg.Overview{}
	if got := BuildCommitContextIndex(o, nil); got != nil {
		t.Fatalf("expected nil for empty overview, got %+v", got)
	}
}

func TestBuildCommitContextIndex_okSmallDiff(t *testing.T) {
	o := &gitpkg.Overview{
		Modified: 1,
		FileChanges: []gitpkg.FileChange{
			{Path: "internal/app/foo.go", Status: "modified", Insertions: 10, Deletions: 2},
		},
	}
	cfg := &config.Config{MaxDiffBytes: 120000}
	idx := BuildCommitContextIndex(o, cfg)
	if idx == nil {
		t.Fatal("expected index")
	}
	if idx.Level != ContextLevelOK {
		t.Fatalf("level: got %s want %s", idx.Level, ContextLevelOK)
	}
	if idx.RecommendCommit {
		t.Fatal("should not recommend commit for small diff")
	}
	if idx.FileCount != 1 || idx.Insertions != 10 || idx.Deletions != 2 {
		t.Fatalf("counts: %+v", idx)
	}
}

func TestBuildCommitContextIndex_attentionOnAreas(t *testing.T) {
	o := &gitpkg.Overview{
		Modified: 2,
		FileChanges: []gitpkg.FileChange{
			{Path: "internal/app/a.go", Status: "modified", Insertions: 5, Deletions: 0},
			{Path: "frontend/src/App.tsx", Status: "modified", Insertions: 5, Deletions: 0},
		},
	}
	idx := BuildCommitContextIndex(o, &config.Config{MaxDiffBytes: 120000})
	if idx == nil {
		t.Fatal("expected index")
	}
	if idx.AreaCount < 2 {
		t.Fatalf("areaCount: got %d", idx.AreaCount)
	}
	if idx.Level != ContextLevelAttention {
		t.Fatalf("level: got %s want %s", idx.Level, ContextLevelAttention)
	}
	if !idx.RecommendCommit {
		t.Fatal("expected recommend commit")
	}
}

func TestBuildCommitContextIndex_criticalNearTruncate(t *testing.T) {
	o := &gitpkg.Overview{
		Modified: 1,
		FileChanges: []gitpkg.FileChange{
			{Path: "big.go", Status: "modified", Insertions: 3000, Deletions: 0},
		},
	}
	cfg := &config.Config{MaxDiffBytes: 50_000}
	idx := BuildCommitContextIndex(o, cfg)
	if idx == nil {
		t.Fatal("expected index")
	}
	if !idx.NearTruncate {
		t.Fatalf("expected nearTruncate, estimated=%d max=%d", idx.EstimatedBytes, idx.MaxDiffBytes)
	}
	if idx.Level != ContextLevelCritical {
		t.Fatalf("level: got %s want %s", idx.Level, ContextLevelCritical)
	}
}

func TestBuildCommitContextIndex_untrackedFloor(t *testing.T) {
	o := &gitpkg.Overview{
		Untracked: 1,
		FileChanges: []gitpkg.FileChange{
			{Path: "newfile.go", Status: "untracked", Insertions: 0, Deletions: 0},
		},
	}
	idx := BuildCommitContextIndex(o, &config.Config{MaxDiffBytes: 120000})
	if idx == nil {
		t.Fatal("expected index")
	}
	if idx.Insertions < contextUntrackedLineFloor {
		t.Fatalf("untracked floor: insertions=%d", idx.Insertions)
	}
}
