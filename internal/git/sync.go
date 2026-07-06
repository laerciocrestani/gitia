package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func (r *Repo) IsClean() (bool, error) {
	out, err := r.run("status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) == "", nil
}

func (r *Repo) FetchPrune() error {
	return r.runInteractive("fetch", "origin", "--prune")
}

func (r *Repo) PullBase(base string) error {
	resolved, err := r.ResolveBase(base)
	if err != nil {
		return err
	}

	remoteBranch := resolved
	if !strings.Contains(remoteBranch, "/") {
		remoteBranch = "origin/" + strings.TrimPrefix(resolved, "origin/")
	}

	localBranch := strings.TrimPrefix(remoteBranch, "origin/")
	current, err := r.CurrentBranch()
	if err != nil {
		return err
	}

	if current != localBranch {
		if _, err := r.run("checkout", localBranch); err != nil {
			return fmt.Errorf("checkout %s: %w", localBranch, err)
		}
	}

	return r.runInteractive("pull", "--ff-only", "origin", localBranch)
}

func (r *Repo) MergedLocalBranches(base string) ([]string, error) {
	resolved, err := r.mergedRef(base)
	if err != nil {
		return nil, err
	}

	out, err := r.run("branch", "--merged", resolved, "--format=%(refname:short)")
	if err != nil {
		return nil, err
	}

	current, _ := r.CurrentBranch()
	protected := protectedBranches(base)

	var branches []string
	for _, name := range splitLines(out) {
		if name == "" || protected[name] || name == current {
			continue
		}
		branches = append(branches, name)
	}
	return branches, nil
}

func (r *Repo) MergedRemoteBranches(base string) ([]string, error) {
	resolved, err := r.mergedRef(base)
	if err != nil {
		return nil, err
	}

	out, err := r.run("branch", "-r", "--merged", resolved, "--format=%(refname:short)")
	if err != nil {
		return nil, err
	}

	protected := protectedBranches(base)
	for k := range protected {
		protected["origin/"+k] = true
	}
	protected["origin/HEAD"] = true

	var branches []string
	for _, name := range splitLines(out) {
		name = strings.TrimSpace(name)
		if name == "" || protected[name] || strings.Contains(name, "->") {
			continue
		}
		short := strings.TrimPrefix(name, "origin/")
		if short == "" || protected[short] {
			continue
		}
		branches = append(branches, short)
	}
	return uniqueStrings(branches), nil
}

func (r *Repo) DeleteLocalBranch(name string) error {
	_, err := r.run("branch", "-d", name)
	return err
}

func (r *Repo) DeleteRemoteBranch(name string) error {
	_, err := r.run("push", "origin", "--delete", name)
	return err
}

func (r *Repo) mergedRef(base string) (string, error) {
	resolved, err := r.ResolveBase(base)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(resolved, "origin/") {
		return resolved, nil
	}
	if _, err := r.run("rev-parse", "--verify", "origin/"+resolved); err == nil {
		return "origin/" + strings.TrimPrefix(resolved, "origin/"), nil
	}
	return resolved, nil
}

func (r *Repo) runInteractive(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = r.dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func protectedBranches(base string) map[string]bool {
	names := []string{
		base,
		strings.TrimPrefix(base, "origin/"),
		"main",
		"master",
		"develop",
		"development",
	}
	protected := make(map[string]bool, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name != "" {
			protected[name] = true
		}
	}
	return protected
}

func splitLines(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func uniqueStrings(items []string) []string {
	seen := make(map[string]bool, len(items))
	var out []string
	for _, item := range items {
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}
	return out
}
