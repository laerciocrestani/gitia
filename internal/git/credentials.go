package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// EnsureGitHubCredentials configures git to use GitHub CLI for HTTPS auth when gh is logged in.
func EnsureGitHubCredentials() error {
	if _, err := exec.LookPath("gh"); err != nil {
		return nil
	}

	status := exec.Command("gh", "auth", "status")
	status.Stdout = nil
	status.Stderr = nil
	if err := status.Run(); err != nil {
		return nil
	}

	setup := exec.Command("gh", "auth", "setup-git")
	setup.Stdout = nil
	setup.Stderr = nil
	if err := setup.Run(); err != nil {
		return fmt.Errorf("gh auth setup-git: %w (rode: gh auth login)", err)
	}
	return nil
}

func wrapGitAuthError(args []string, stderr string, err error) error {
	if err == nil {
		return nil
	}
	msg := strings.TrimSpace(stderr)
	lower := strings.ToLower(msg)
	if strings.Contains(lower, "authentication failed") ||
		strings.Contains(lower, "could not read username") ||
		strings.Contains(lower, "invalid username or password") ||
		strings.Contains(lower, "terminal prompts disabled") ||
		strings.Contains(msg, "Username for 'https://github.com'") {
		return fmt.Errorf(
			"git %s: autenticação GitHub falhou — rode: gh auth setup-git (ou gh auth login)\n  %s",
			strings.Join(args, " "),
			msg,
		)
	}
	return fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
}
