package app

import (
	"strings"

	gitpkg "github.com/laerciocrestani/gitai/internal/git"
	prpkg "github.com/laerciocrestani/gitai/internal/pr"
)

// ResolveBranchStage picks the primary lifecycle badge for a branch.
func ResolveBranchStage(
	info gitpkg.BranchInfo,
	aheadOfBase int,
	merged bool,
	pr *prpkg.PRView,
	isCurrentDirty bool,
	base string,
) gitpkg.BranchStage {
	if info.Name == base {
		return gitpkg.StageBase
	}

	if info.Current && isCurrentDirty {
		return gitpkg.StageWIP
	}

	if info.Ahead > 0 {
		return gitpkg.StagePush
	}

	if info.Behind > 0 {
		return gitpkg.StageSync
	}

	if pr != nil {
		if pr.IsDraft {
			return gitpkg.StagePRDraft
		}
		if strings.EqualFold(pr.State, "OPEN") {
			return gitpkg.StagePROpen
		}
	}

	if merged {
		return gitpkg.StageMerged
	}

	if aheadOfBase > 0 {
		return gitpkg.StageReady
	}

	if info.Upstream == "" {
		return gitpkg.StageStale
	}

	return gitpkg.StageOK
}

// ListBranchSummaries returns local branches enriched with lifecycle status.
func ListBranchSummaries(snap *WorkspaceSnapshot) ([]gitpkg.BranchSummary, error) {
	branches, err := ListBranches()
	if err != nil {
		return nil, err
	}
	if len(branches) == 0 {
		return nil, nil
	}

	base := "main"
	isDirty := false
	if snap != nil && snap.Overview != nil {
		if snap.Overview.BaseBranch != "" {
			base = snap.Overview.BaseBranch
		}
		isDirty = snap.Overview.IsDirty()
	}

	repo, err := gitpkg.New()
	if err != nil {
		return nil, err
	}

	names := make([]string, len(branches))
	for i, b := range branches {
		names[i] = b.Name
	}

	mergedSet, err := repo.MergedBranchSet(base)
	if err != nil {
		mergedSet = map[string]bool{}
	}

	aheadOfBase, err := repo.CommitsAheadOfBaseByBranch(base, names)
	if err != nil {
		aheadOfBase = map[string]int{}
	}

	prByHead := map[string]prpkg.PRView{}
	if snap != nil && snap.HasGH {
		if client, err := prpkg.New(); err == nil {
			if prs, err := client.ListOpen(); err == nil {
				prByHead = prs
			}
		}
	}

	summaries := make([]gitpkg.BranchSummary, len(branches))
	for i, info := range branches {
		var pr *prpkg.PRView
		if view, ok := prByHead[info.Name]; ok {
			pr = &view
		}

		stage := ResolveBranchStage(
			info,
			aheadOfBase[info.Name],
			mergedSet[info.Name],
			pr,
			isDirty,
			base,
		)

		summary := gitpkg.BranchSummary{
			BranchInfo:         info,
			Stage:              stage,
			CommitsAheadOfBase: aheadOfBase[info.Name],
		}
		if pr != nil {
			summary.PRNumber = pr.Number
			summary.IsDraft = pr.IsDraft
		}
		summaries[i] = summary
	}

	return summaries, nil
}
