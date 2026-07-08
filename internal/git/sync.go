package git

import (
	"fmt"
	"strings"
)

func (r *Repo) IsClean() (bool, error) {
	out, err := r.run("status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) == "", nil
}

// FetchPrune updates remote refs and prunes stale tracking branches.
func (r *Repo) FetchPrune() error {
	_, err := r.run("fetch", "origin", "--prune")
	return err
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

	_, err = r.run("pull", "--ff-only", "origin", localBranch)
	return err
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

	// strip=2 → origin/<branch>; refname:short inclui "origin" (namespace do remote).
	out, err := r.run("branch", "-r", "--merged", resolved, "--format=%(refname:strip=2)")
	if err != nil {
		return nil, err
	}

	protected := protectedBranches(base)
	for k := range protected {
		protected["origin/"+k] = true
	}
	protected["origin/HEAD"] = true
	protected["origin"] = true

	var branches []string
	for _, name := range splitLines(out) {
		name = strings.TrimSpace(name)
		if name == "" || protected[name] || strings.Contains(name, "->") {
			continue
		}
		if !strings.HasPrefix(name, "origin/") {
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
