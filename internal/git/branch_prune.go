package git

import (
	"fmt"
	"strconv"
	"strings"
)

// BranchPruneIssue descreve divergência entre uma branch local e seu upstream.
type BranchPruneIssue struct {
	Name          string
	Upstream      string
	LocalAhead    int
	RemoteAhead   int
	LocalCommits  []string
	RemoteCommits []string
}

const maxPruneCommitPreview = 8

// BranchUpstream retorna o tracking branch configurado (ex.: origin/feature/x).
func (r *Repo) BranchUpstream(name string) (string, error) {
	return r.run("rev-parse", "--abbrev-ref", "--symbolic-full-name", name+"@{upstream}")
}

// BranchAheadBehind compara branch com upstream.
// ahead = commits em branch que não estão no upstream; behind = o inverso.
func (r *Repo) BranchAheadBehind(branch, upstream string) (ahead, behind int, err error) {
	if _, err := r.run("rev-parse", "--verify", upstream); err != nil {
		return 0, 0, fmt.Errorf("upstream %q não encontrado", upstream)
	}

	out, err := r.run("rev-list", "--left-right", "--count", fmt.Sprintf("%s...%s", branch, upstream))
	if err != nil {
		return 0, 0, err
	}
	parts := strings.Fields(out)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("formato inesperado: %q", out)
	}
	ahead, _ = strconv.Atoi(parts[0])
	behind, _ = strconv.Atoi(parts[1])
	return ahead, behind, nil
}

// LocalBranchPruneIssue detecta commits locais/remotos fora de sincronia antes do prune.
// Retorna nil quando não há upstream ou quando branch e upstream estão alinhados.
func (r *Repo) LocalBranchPruneIssue(name string) (*BranchPruneIssue, error) {
	upstream, err := r.BranchUpstream(name)
	if err != nil {
		return nil, nil
	}

	ahead, behind, err := r.BranchAheadBehind(name, upstream)
	if err != nil {
		return nil, err
	}
	if ahead == 0 && behind == 0 {
		return nil, nil
	}

	issue := &BranchPruneIssue{
		Name:        name,
		Upstream:    upstream,
		LocalAhead:  ahead,
		RemoteAhead: behind,
	}

	if ahead > 0 {
		commits, err := r.logOnelineRange(upstream, name, maxPruneCommitPreview)
		if err != nil {
			return nil, err
		}
		issue.LocalCommits = commits
	}
	if behind > 0 {
		commits, err := r.logOnelineRange(name, upstream, maxPruneCommitPreview)
		if err != nil {
			return nil, err
		}
		issue.RemoteCommits = commits
	}

	return issue, nil
}

func (r *Repo) logOnelineRange(base, head string, limit int) ([]string, error) {
	args := []string{"log", fmt.Sprintf("%s..%s", base, head), "--oneline", "--no-decorate"}
	if limit > 0 {
		args = append(args, "-n", strconv.Itoa(limit))
	}
	out, err := r.run(args...)
	if err != nil {
		return nil, err
	}
	return splitLines(out), nil
}

// DeleteLocalBranchForce remove a branch local com git branch -D.
func (r *Repo) DeleteLocalBranchForce(name string) error {
	_, err := r.run("branch", "-D", name)
	return err
}
