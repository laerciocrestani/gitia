package app

import (
	"testing"

	gitpkg "github.com/laerciocrestani/gitai/internal/git"
)

func TestRecommendPruneBranchAction_localAheadOnly(t *testing.T) {
	rec := RecommendPruneBranchAction(&gitpkg.BranchPruneIssue{
		LocalAhead:  2,
		RemoteAhead: 0,
	})
	if rec.Action != PruneBranchDeleteForce {
		t.Fatalf("action = %v, want force delete", rec.Action)
	}
}

func TestRecommendPruneBranchAction_remoteAheadOnly(t *testing.T) {
	rec := RecommendPruneBranchAction(&gitpkg.BranchPruneIssue{
		LocalAhead:  0,
		RemoteAhead: 1,
	})
	if rec.Action != PruneBranchKeep {
		t.Fatalf("action = %v, want keep", rec.Action)
	}
}

func TestRecommendPruneBranchAction_diverged(t *testing.T) {
	rec := RecommendPruneBranchAction(&gitpkg.BranchPruneIssue{
		LocalAhead:  1,
		RemoteAhead: 1,
	})
	if rec.Action != PruneBranchKeep {
		t.Fatalf("action = %v, want keep", rec.Action)
	}
}
