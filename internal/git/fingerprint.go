package git

import (
	"fmt"
	"hash/fnv"
)

// StatusFingerprint returns a cheap digest of HEAD + working-tree status.
// Used by desktop/TUI watchers to decide when a full dashboard reload is needed.
func (r *Repo) StatusFingerprint() (string, error) {
	head, err := r.run("rev-parse", "HEAD")
	if err != nil {
		// Empty repo (no commits yet): fall back to a stable placeholder.
		head = "UNBORN"
	}
	status, err := r.runRaw("status", "--porcelain")
	if err != nil {
		return "", err
	}
	h := fnv.New64a()
	_, _ = h.Write([]byte(head))
	_, _ = h.Write([]byte{0})
	_, _ = h.Write([]byte(status))
	return fmt.Sprintf("%016x", h.Sum64()), nil
}

// StatusFingerprintAt opens path and returns StatusFingerprint.
func StatusFingerprintAt(path string) (string, error) {
	repo, err := Open(path)
	if err != nil {
		return "", err
	}
	if err := repo.IsRepo(); err != nil {
		return "", err
	}
	return repo.StatusFingerprint()
}
