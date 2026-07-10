package app

import (
	"testing"

	gitpkg "github.com/laerciocrestani/gitai/internal/git"
	prpkg "github.com/laerciocrestani/gitai/internal/pr"
)

func TestResolveBranchStage(t *testing.T) {
	base := "main"
	prOpen := &prpkg.PRView{State: "OPEN", Number: 7}
	prDraft := &prpkg.PRView{State: "OPEN", Number: 3, IsDraft: true}

	cases := []struct {
		name       string
		info       gitpkg.BranchInfo
		ahead      int
		merged     bool
		pr         *prpkg.PRView
		dirty      bool
		want       gitpkg.BranchStage
	}{
		{
			name: "base branch",
			info: gitpkg.BranchInfo{Name: "main"},
			want: gitpkg.StageBase,
		},
		{
			name:  "wip on current branch",
			info:  gitpkg.BranchInfo{Name: "feat/x", Current: true},
			dirty: true,
			want:  gitpkg.StageWIP,
		},
		{
			name: "unpushed commits",
			info: gitpkg.BranchInfo{Name: "feat/x", Ahead: 2},
			want: gitpkg.StagePush,
		},
		{
			name: "behind remote",
			info: gitpkg.BranchInfo{Name: "feat/x", Behind: 1},
			want: gitpkg.StageSync,
		},
		{
			name: "open pr",
			info: gitpkg.BranchInfo{Name: "feat/x"},
			pr:   prOpen,
			want: gitpkg.StagePROpen,
		},
		{
			name: "draft pr",
			info: gitpkg.BranchInfo{Name: "feat/x"},
			pr:   prDraft,
			want: gitpkg.StagePRDraft,
		},
		{
			name:   "merged branch",
			info:   gitpkg.BranchInfo{Name: "feat/x"},
			merged: true,
			want:   gitpkg.StageMerged,
		},
		{
			name:  "ready for pr",
			info:  gitpkg.BranchInfo{Name: "feat/x"},
			ahead: 3,
			want:  gitpkg.StageReady,
		},
		{
			name: "aligned with base",
			info: gitpkg.BranchInfo{Name: "feat/x", Upstream: "origin/feat/x"},
			want: gitpkg.StageOK,
		},
		{
			name: "local only",
			info: gitpkg.BranchInfo{Name: "feat/x"},
			want: gitpkg.StageStale,
		},
		{
			name: "push beats pr",
			info: gitpkg.BranchInfo{Name: "feat/x", Ahead: 1},
			pr:   prOpen,
			want: gitpkg.StagePush,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ResolveBranchStage(tc.info, tc.ahead, tc.merged, tc.pr, tc.dirty, base)
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}
