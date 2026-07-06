package app

import (
	"fmt"

	"github.com/laerciocrestani/gitai/internal/config"
	gitpkg "github.com/laerciocrestani/gitai/internal/git"
	"github.com/laerciocrestani/gitai/internal/ui"
)

type SyncOptions struct {
	Prune       bool
	PruneRemote bool
	Base        string
	DryRun      bool
	Progress    Progress
}

func RunSync(opts SyncOptions) error {
	prog := opts.Progress
	if prog == nil {
		sess := ui.New("sync", opts.DryRun)
		sess.Header()
		prog = sess
	}

	repo, err := gitpkg.New()
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

	base := opts.Base
	if base == "" {
		if cfg, err := config.Load(); err == nil {
			base = cfg.BaseBranch
		}
	}
	if base == "" {
		base = "main"
	}

	previous, err := repo.CurrentBranch()
	if err != nil {
		return err
	}

	fmt.Println()
	if sess, ok := prog.(*ui.Session); ok {
		sess.MetaRow("Base", base)
		if previous != base && !opts.shouldPrune() {
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

	if err := prog.Step("Pulling "+base, func() error {
		if opts.DryRun {
			prog.Detail(fmt.Sprintf("git checkout %s && git pull --ff-only origin %s", base, base))
			return nil
		}
		return repo.PullBase(base)
	}); err != nil {
		return err
	}

	if !opts.shouldPrune() {
		prog.Success("Synced with origin/" + base)
		return nil
	}

	var local, remote []string
	if err := prog.Step("Finding merged branches", func() error {
		var err error
		if opts.pruneLocal() {
			local, err = repo.MergedLocalBranches(base)
			if err != nil {
				return err
			}
		}
		if opts.pruneRemote() {
			remote, err = repo.MergedRemoteBranches(base)
			if err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	if len(local) == 0 && len(remote) == 0 {
		prog.Info("No merged branches to prune")
		prog.Success("Synced with origin/" + base)
		return nil
	}

	if sess, ok := prog.(*ui.Session); ok {
		sess.Section("Prune")
	}
	for _, name := range local {
		if err := pruneLocal(prog, repo, name, opts.DryRun); err != nil {
			return err
		}
	}
	for _, name := range remote {
		if err := pruneRemote(prog, repo, name, opts.DryRun); err != nil {
			return err
		}
	}

	msg := "Synced"
	if len(local) > 0 {
		msg += fmt.Sprintf(" · %d local removed", len(local))
	}
	if len(remote) > 0 {
		msg += fmt.Sprintf(" · %d remote removed", len(remote))
	}
	prog.Success(msg)
	return nil
}

func (o SyncOptions) shouldPrune() bool {
	return o.Prune || o.PruneRemote
}

func (o SyncOptions) pruneLocal() bool {
	return o.Prune
}

func (o SyncOptions) pruneRemote() bool {
	return o.Prune || o.PruneRemote
}

func pruneLocal(prog Progress, repo *gitpkg.Repo, name string, dryRun bool) error {
	return prog.Step("Removing local "+name, func() error {
		if dryRun {
			prog.Detail("git branch -d " + name)
			return nil
		}
		return repo.DeleteLocalBranch(name)
	})
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
