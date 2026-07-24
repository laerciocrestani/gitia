package gha

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/laerciocrestani/openbench/internal/config"
	gitpkg "github.com/laerciocrestani/openbench/internal/git"
)

// ResolveDefaultBranch returns the repo default branch (origin/HEAD, config, or main).
func ResolveDefaultBranch(dir string) string {
	cmd := exec.Command("git", "symbolic-ref", "--short", "refs/remotes/origin/HEAD")
	cmd.Dir = dir
	if out, err := cmd.Output(); err == nil {
		ref := strings.TrimSpace(string(out))
		ref = strings.TrimPrefix(ref, "origin/")
		if ref != "" {
			return ref
		}
	}
	if cfg, err := config.Load(); err == nil && strings.TrimSpace(cfg.BaseBranch) != "" {
		return strings.TrimSpace(cfg.BaseBranch)
	}
	repo, err := gitpkg.Open(dir)
	if err == nil {
		if b, err := repo.ResolveBase("main"); err == nil && b != "" {
			return strings.TrimPrefix(b, "origin/")
		}
	}
	return "main"
}

// IsDefaultBranch reports whether branch is the default (main/master/configured).
func IsDefaultBranch(branch, defaultBranch string) bool {
	b := strings.TrimSpace(branch)
	d := strings.TrimSpace(defaultBranch)
	if b == "" {
		return false
	}
	if d != "" {
		return strings.EqualFold(b, d)
	}
	lb := strings.ToLower(b)
	return lb == "main" || lb == "master"
}

// DefaultBranchWarning returns a user-facing warning when pushing the default branch.
func DefaultBranchWarning(branch, defaultBranch string) string {
	if !IsDefaultBranch(branch, defaultBranch) {
		return ""
	}
	name := strings.TrimSpace(branch)
	if name == "" {
		name = defaultBranch
	}
	return fmt.Sprintf(
		"Você está em %s (branch default). O push vai disparar CI automaticamente no GitHub Actions.",
		name,
	)
}
