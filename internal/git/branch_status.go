package git

import (
	"fmt"
	"strconv"
	"strings"
)

// BranchStage describes the lifecycle stage of a branch.
type BranchStage string

const (
	StageBase    BranchStage = "base"
	StageWIP     BranchStage = "wip"
	StageMerged  BranchStage = "merged"
	StagePROpen  BranchStage = "pr_open"
	StagePRDraft BranchStage = "pr_draft"
	StagePush    BranchStage = "push"
	StageSync    BranchStage = "sync"
	StageReady   BranchStage = "ready"
	StageOK      BranchStage = "ok"
	StageStale   BranchStage = "stale"
)

// BranchSummary extends BranchInfo with lifecycle status for the branch list.
type BranchSummary struct {
	BranchInfo
	Stage              BranchStage
	CommitsAheadOfBase int
	PRNumber           int
	IsDraft            bool
}

// MergedBranchSet returns local branches whose tips are reachable from base.
func (r *Repo) MergedBranchSet(base string) (map[string]bool, error) {
	resolved, err := r.mergedRef(base)
	if err != nil {
		return nil, err
	}

	out, err := r.run("branch", "--merged", resolved, "--format=%(refname:short)")
	if err != nil {
		return nil, err
	}

	set := make(map[string]bool)
	for _, name := range splitLines(out) {
		if name != "" {
			set[name] = true
		}
	}
	return set, nil
}

// CommitsAheadOfBaseByBranch counts commits on each branch that are not in base.
func (r *Repo) CommitsAheadOfBaseByBranch(base string, branches []string) (map[string]int, error) {
	resolved, err := r.ResolveBase(base)
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int, len(branches))
	for _, name := range branches {
		if name == "" {
			continue
		}
		out, err := r.run("rev-list", "--count", fmt.Sprintf("%s..%s", resolved, name))
		if err != nil {
			counts[name] = 0
			continue
		}
		counts[name], _ = strconv.Atoi(strings.TrimSpace(out))
	}
	return counts, nil
}
