package gha

import (
	"os/exec"
	"strings"

	gitpkg "github.com/laerciocrestani/openbench/internal/git"
)

// LoadStatus lists runs for the current branch (unless filter overrides) and usage.
func (c *Client) LoadStatus(f ListFilter) (*StatusSnapshot, error) {
	repo, err := gitpkg.Open(c.dir)
	if err != nil {
		return nil, err
	}
	branch := strings.TrimSpace(f.Branch)
	if branch == "" {
		b, bErr := repo.CurrentBranch()
		if bErr == nil && b != "" && b != "HEAD" {
			branch = b
			f.Branch = branch
		}
	}
	headSHA := strings.TrimSpace(f.HeadSHA)
	if headSHA == "" {
		headSHA = gitHeadSHA(c.dir)
	}

	runs, err := c.ListRuns(f)
	if err != nil {
		return nil, err
	}

	owner := c.RemoteOwner()
	usage := c.UsageForRuns(runs, owner)

	return &StatusSnapshot{
		Branch:  branch,
		HeadSHA: headSHA,
		Runs:    runs,
		Usage:   usage,
		Filter:  f,
	}, nil
}

// RemoteOwner returns the GitHub owner/org from origin, if resolvable.
func (c *Client) RemoteOwner() string {
	repo, err := gitpkg.Open(c.dir)
	if err != nil {
		return ""
	}
	url, err := repo.RemoteOriginURL()
	if err != nil {
		return ""
	}
	return parseOwnerFromRemote(url)
}

func gitHeadSHA(dir string) string {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func parseOwnerFromRemote(url string) string {
	url = strings.TrimSpace(url)
	url = strings.TrimSuffix(url, ".git")
	// SSH: git@host:owner/repo (avoid matching https://)
	if !strings.Contains(url, "://") {
		if i := strings.Index(url, ":"); i >= 0 {
			rest := url[i+1:]
			parts := strings.Split(rest, "/")
			if len(parts) >= 1 && parts[0] != "" {
				return parts[0]
			}
		}
	}
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	parts := strings.Split(url, "/")
	if len(parts) >= 2 {
		return parts[1]
	}
	return ""
}
