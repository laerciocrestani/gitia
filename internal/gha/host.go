package gha

import (
	"os"
	"os/exec"
	"strings"
)

// ResolveHost returns the GitHub host for the repo (github.com or Enterprise).
func ResolveHost(dir string) string {
	if h := strings.TrimSpace(os.Getenv("GH_HOST")); h != "" {
		return h
	}
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "github.com"
	}
	return hostFromRemote(strings.TrimSpace(string(out)))
}

func hostFromRemote(url string) string {
	url = strings.TrimSpace(url)
	if url == "" {
		return "github.com"
	}
	// git@host:owner/repo
	if !strings.Contains(url, "://") {
		if i := strings.Index(url, "@"); i >= 0 {
			rest := url[i+1:]
			if j := strings.Index(rest, ":"); j >= 0 {
				h := rest[:j]
				if h != "" {
					return h
				}
			}
		}
		return "github.com"
	}
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "ssh://git@")
	url = strings.TrimPrefix(url, "ssh://")
	if i := strings.Index(url, "/"); i >= 0 {
		url = url[:i]
	}
	if i := strings.Index(url, ":"); i >= 0 {
		// host:port
		// keep host only if port-like
	}
	if url == "" {
		return "github.com"
	}
	return url
}

// IsEnterpriseHost reports non-github.com hosts.
func IsEnterpriseHost(host string) bool {
	h := strings.ToLower(strings.TrimSpace(host))
	return h != "" && h != "github.com"
}
