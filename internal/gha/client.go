package gha

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Client talks to GitHub Actions via the gh CLI in a repo workdir.
type Client struct {
	dir string
}

// Open returns a client bound to dir (must be inside a git repo with origin).
func Open(dir string) (*Client, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	if _, err := exec.LookPath("gh"); err != nil {
		return nil, &Error{Kind: ErrGhNotInstalled, Message: ErrGhNotInstalled.Error(), Cause: err}
	}
	return &Client{dir: abs}, nil
}

// New opens a client for the process working directory.
func New() (*Client, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	return Open(dir)
}

func (c *Client) Dir() string { return c.dir }

func (c *Client) run(args ...string) (string, error) {
	cmd := exec.Command("gh", args...)
	cmd.Dir = c.dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", classifyGhErr(stderr.String(), fmt.Errorf("gh %s: %w", strings.Join(args, " "), err))
	}
	return strings.TrimSpace(stdout.String()), nil
}

// runAllowExit runs gh and returns stdout even when exit code != 0 (some gh
// commands exit non-zero for failed runs). Caller must validate JSON.
func (c *Client) runAllowExit(args ...string) (string, string, error) {
	cmd := exec.Command("gh", args...)
	cmd.Dir = c.dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return strings.TrimSpace(stdout.String()), strings.TrimSpace(stderr.String()), err
}
