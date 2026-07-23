package app

import (
	"fmt"
	"strings"

	"github.com/laerciocrestani/openbench/internal/config"
	gitpkg "github.com/laerciocrestani/openbench/internal/git"
	"github.com/laerciocrestani/openbench/internal/ui"
)

// SyncOptions configures base-branch synchronization (fetch + fast-forward).
// Prune flags are used internally by RunHygiene via discoverPruneCandidates;
// RunSync ignores them.
type SyncOptions struct {
	Prune          bool // full prune (local + remote) — hygiene only
	PruneRemote    bool // deprecated remote-only; unused by new hygiene modes
	PruneLocalOnly bool // local-only prune — hygiene only
	Base           string
	DryRun         bool
	WorkDir        string // optional; when set, git runs in this directory
	Progress       Progress
}

// RunSync fetches origin and fast-forwards the local base branch.
// When already on the base, pulls --ff-only. When on a feature branch, updates
// the local base ref from origin without leaving the current branch.
func RunSync(opts SyncOptions) error {
	prog := opts.Progress
	if prog == nil {
		sess := ui.New("sync", opts.DryRun)
		sess.Header()
		prog = sess
	}

	repo, err := openRepo(opts.WorkDir)
	if err != nil {
		return err
	}
	if err := repo.IsRepo(); err != nil {
		return fmt.Errorf("diretório atual não é um repositório git")
	}

	clean, err := repo.IsClean()
	if err != nil {
		return err
	}
	if !clean {
		return fmt.Errorf("working tree com alterações — commit ou stash antes de sincronizar")
	}

	base := strings.TrimSpace(opts.Base)
	if base == "" {
		if cfg, err := config.Load(); err == nil {
			base = cfg.BaseBranch
		}
	}
	if base == "" {
		base = "main"
	}
	base = strings.TrimPrefix(base, "origin/")

	previous, err := repo.CurrentBranch()
	if err != nil {
		return err
	}

	fmt.Println()
	if sess, ok := prog.(*ui.Session); ok {
		sess.MetaRow("Base", base)
		if previous != base {
			sess.MetaRow("Branch", previous)
		}
		sess.Divider()
	}

	if err := prog.Step("Fetching origin", func() error {
		if opts.DryRun {
			prog.Detail("git fetch origin --prune")
			return nil
		}
		return repo.FetchPrune()
	}); err != nil {
		return err
	}

	if err := prog.Step("Updating "+base, func() error {
		if opts.DryRun {
			if previous == base {
				prog.Detail("git pull --ff-only origin " + base)
			} else {
				prog.Detail(fmt.Sprintf("git fetch origin %s:%s", base, base))
			}
			return nil
		}
		if previous == base {
			if err := repo.PullFFOnly(); err != nil {
				DiagnoseSyncFailure(base, err, prog)
				return err
			}
			return nil
		}
		updated, err := repo.UpdateLocalBranchFromOrigin(base)
		if err != nil {
			DiagnoseSyncFailure(base, err, prog)
			return err
		}
		if !updated {
			prog.Detail(base + " already up to date")
		}
		return nil
	}); err != nil {
		return err
	}

	prog.Success("Synced with origin/" + base)
	return nil
}

func discoverPruneCandidates(prog Progress, repo *gitpkg.Repo, opts SyncOptions, base string) (local, remote []string, err error) {
	err = prog.Step("Finding branches to prune", func() error {
		if opts.pruneLocal() {
			local, err = repo.LocalPruneCandidates(base)
			if err != nil {
				return err
			}
		}
		if opts.pruneRemote() {
			remote, err = repo.RemotePruneCandidates(base)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return local, remote, err
}

func refreshOriginAfterRemotePrune(prog Progress, repo *gitpkg.Repo, dryRun bool) error {
	return prog.Step("Refreshing origin", func() error {
		if dryRun {
			prog.Detail("git fetch origin --prune")
			return nil
		}
		return repo.FetchPrune()
	})
}

func pruneRemoteBranches(prog Progress, repo *gitpkg.Repo, names []string, dryRun bool) (int, error) {
	removed := 0
	for _, name := range names {
		if err := pruneRemote(prog, repo, name, dryRun); err != nil {
			return removed, err
		}
		removed++
	}
	return removed, nil
}

func pruneLocalBranches(prog Progress, repo *gitpkg.Repo, names []string, base string, dryRun bool) (int, error) {
	removed := 0
	for _, name := range names {
		ok, err := pruneLocal(prog, repo, name, base, dryRun)
		if err != nil {
			return removed, err
		}
		if ok {
			removed++
		}
	}
	return removed, nil
}

func (o SyncOptions) shouldPrune() bool {
	return o.Prune || o.PruneRemote || o.PruneLocalOnly
}

func (o SyncOptions) pruneLocal() bool {
	return o.Prune || o.PruneLocalOnly
}

func (o SyncOptions) pruneRemote() bool {
	return o.Prune || o.PruneRemote
}

func pruneLocal(prog Progress, repo *gitpkg.Repo, name, base string, dryRun bool) (bool, error) {
	issue, err := repo.LocalBranchPruneIssue(name)
	if err != nil {
		return false, err
	}

	force := false
	if issue != nil && issue.UpstreamGone {
		if dryRun {
			prog.Info(name + ": upstream removido no remoto — usaria git branch -D")
			return false, nil
		}
		force = true
	} else if issue != nil && issue.LocalAhead > 0 {
		if dryRun {
			prog.Info(fmt.Sprintf(
				"%s: diverge de %s (%d commit(s) local não enviado(s)) — usaria -D após confirmação",
				name, issue.Upstream, issue.LocalAhead,
			))
			for _, commit := range issue.LocalCommits {
				prog.Detail(commit)
			}
			return false, nil
		}

		sess, ok := prog.(*ui.Session)
		if !ok {
			rec := RecommendPruneBranchAction(issue)
			if rec.Action == PruneBranchKeep {
				prog.Info("Mantida: " + name + " — " + rec.Reason)
				return false, nil
			}
			force = true
		} else {
			action, err := promptPruneBranchConflict(sess, issue)
			if err != nil {
				return false, err
			}
			if action == PruneBranchKeep {
				prog.Info("Mantida: " + name)
				return false, nil
			}
			force = true
		}
	}

	if !force {
		absorbed, err := repo.BranchAbsorbedIntoBase(name, base)
		if err != nil {
			return false, err
		}
		if absorbed {
			if dryRun {
				prog.Info(name + ": alterações já estão na base (squash/rebase) — usaria git branch -D")
				return false, nil
			}
			force = true
		}
	}

	label := "Removing local " + name
	if force {
		label += " (forced)"
	}

	err = prog.Step(label, func() error {
		if dryRun {
			if force {
				prog.Detail("git branch -D " + name)
			} else {
				prog.Detail("git branch -d " + name)
			}
			return nil
		}
		if force {
			return repo.DeleteLocalBranchForce(name)
		}
		return repo.DeleteLocalBranch(name)
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func pruneRemote(prog Progress, repo *gitpkg.Repo, name string, dryRun bool) error {
	return prog.Step("Removing remote "+name, func() error {
		if dryRun {
			prog.Detail("git push origin --delete " + name)
			return nil
		}
		return repo.DeleteRemoteBranch(name)
	})
}
