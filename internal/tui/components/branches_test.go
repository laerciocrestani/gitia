package components_test

import (
	"strings"
	"testing"

	gitpkg "github.com/laerciocrestani/gitai/internal/git"
	"github.com/laerciocrestani/gitai/internal/tui/components"
)

func TestRenderBranchesPanelShowsPosition(t *testing.T) {
	body := components.RenderBranchListLineNumbered(1, gitpkg.BranchSummary{
		BranchInfo: gitpkg.BranchInfo{Name: "feature/x"},
		Stage:      gitpkg.StageReady,
	}, true)
	out := components.RenderBranchesPanel(1, 5, "main", body, 70)
	if !strings.Contains(out, "2/5") {
		t.Fatalf("missing position in title: %q", out)
	}
	if !strings.Contains(out, "feature/x") {
		t.Fatalf("missing branch name: %q", out)
	}
	if !strings.Contains(out, "ready") {
		t.Fatalf("missing stage badge: %q", out)
	}
}

func TestRenderBranchStageBadge(t *testing.T) {
	cases := []struct {
		stage gitpkg.BranchStage
		want  string
	}{
		{gitpkg.StagePROpen, "PR #42"},
		{gitpkg.StageReady, "ready"},
		{gitpkg.StageOK, "ok"},
	}
	for _, tc := range cases {
		summary := gitpkg.BranchSummary{Stage: tc.stage}
		if tc.stage == gitpkg.StagePROpen {
			summary.PRNumber = 42
		}
		out := components.RenderBranchStageBadge(summary)
		if !strings.Contains(out, tc.want) {
			t.Fatalf("stage %q: want %q in %q", tc.stage, tc.want, out)
		}
	}
}

func TestRenderBranchDetailContextTitle(t *testing.T) {
	out := components.RenderBranchDetail(nil, nil, "feature/x", "main", 60, 2)
	if !strings.Contains(out, "Context · feature/x") {
		t.Fatalf("missing context title: %q", out)
	}
}
