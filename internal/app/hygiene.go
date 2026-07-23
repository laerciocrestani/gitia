package app

import (
	"fmt"
	"strings"

	"github.com/laerciocrestani/openbench/internal/config"
	"github.com/laerciocrestani/openbench/internal/ui"
)

// Hygiene mode identifiers (English CLI/API values).
const (
	HygieneModeFull  = "full"  // delete local + remote merged/absorbed branches
	HygieneModeLocal = "local" // delete local only; keep remote
)

// HygieneOptions configures branch cleanup (fetch refs + prune). Does not
// fast-forward the local base tip — use RunSync for that.
type HygieneOptions struct {
	Mode     string // full | local
	Base     string
	DryRun   bool
	WorkDir  string
	Progress Progress
}

// RunHygiene fetches origin refs and prunes merged/absorbed branches.
func RunHygiene(opts HygieneOptions) error {
	prog := opts.Progress
	if prog == nil {
		sess := ui.New("hygiene", opts.DryRun)
		sess.Header()
		prog = sess
	}

	mode := strings.TrimSpace(opts.Mode)
	if mode == "" {
		mode = HygieneModeFull
	}
	localOnly := false
	switch mode {
	case HygieneModeFull:
	case HygieneModeLocal:
		localOnly = true
	default:
		return fmt.Errorf("modo de hygiene inválido: %s (use full ou local)", mode)
	}

	repo, err := openRepo(opts.WorkDir)
	if err != nil {
		return err
	}
	if err := repo.IsRepo(); err != nil {
		return fmt.Errorf("diretório atual não é um repositório git")
	}

	// Dirty working tree is OK: hygiene only deletes other branches / remotes.

	base := strings.TrimSpace(opts.Base)
	if base == "" {
		if cfg, err := config.Load(); err == nil {
			base = cfg.BaseBranch
		}
	}
	if base == "" {
		base = "main"
	}

	fmt.Println()
	if sess, ok := prog.(*ui.Session); ok {
		sess.MetaRow("Base", base)
		sess.MetaRow("Mode", mode)
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

	pruneOpts := SyncOptions{
		PruneLocalOnly: localOnly,
		Prune:          !localOnly,
		Base:           base,
		DryRun:         opts.DryRun,
		WorkDir:        opts.WorkDir,
		Progress:       prog,
	}

	local, remote, err := discoverPruneCandidates(prog, repo, pruneOpts, base)
	if err != nil {
		return err
	}

	if len(local) == 0 && len(remote) == 0 {
		prog.Info("No branches to prune")
		prog.Success("Hygiene complete")
		return nil
	}

	if sess, ok := prog.(*ui.Session); ok {
		sess.Section("Prune")
	}

	remoteRemoved := 0
	if pruneOpts.pruneRemote() {
		remoteRemoved, err = pruneRemoteBranches(prog, repo, remote, opts.DryRun)
		if err != nil {
			return err
		}
		if remoteRemoved > 0 || (opts.DryRun && len(remote) > 0) {
			if err := refreshOriginAfterRemotePrune(prog, repo, opts.DryRun); err != nil {
				return err
			}
		}
		if pruneOpts.pruneLocal() && remoteRemoved > 0 {
			local, err = repo.LocalPruneCandidates(base)
			if err != nil {
				return err
			}
		}
	}

	localRemoved := 0
	if pruneOpts.pruneLocal() {
		localRemoved, err = pruneLocalBranches(prog, repo, local, base, opts.DryRun)
		if err != nil {
			return err
		}
	}

	msg := "Hygiene"
	if localRemoved > 0 {
		msg += fmt.Sprintf(" · %d local removed", localRemoved)
	}
	if remoteRemoved > 0 {
		msg += fmt.Sprintf(" · %d remote removed", remoteRemoved)
	}
	if localRemoved == 0 && remoteRemoved == 0 {
		msg += " · nothing removed"
	}
	prog.Success(msg)
	return nil
}

// CountHygieneCandidates returns local/remote prune candidate counts for UI pulse.
func CountHygieneCandidates(workDir, base string) (local, remote int, err error) {
	repo, err := openRepo(workDir)
	if err != nil {
		return 0, 0, err
	}
	if err := repo.IsRepo(); err != nil {
		return 0, 0, err
	}
	base = strings.TrimSpace(base)
	if base == "" {
		base = "main"
	}
	locs, err := repo.LocalPruneCandidates(base)
	if err != nil {
		return 0, 0, err
	}
	rems, err := repo.RemotePruneCandidates(base)
	if err != nil {
		return 0, 0, err
	}
	return len(locs), len(rems), nil
}
