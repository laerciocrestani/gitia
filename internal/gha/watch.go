package gha

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// WatchOptions controls post-push CI observation.
type WatchOptions struct {
	Branch        string
	HeadSHA       string
	AppearTimeout time.Duration // wait for runs for this SHA (default 45s)
	SettleTimeout time.Duration // wait for terminal conclusions (default 5m)
	PollInterval  time.Duration // default 5s
	OnUpdate      func(WatchSnapshot)
}

// WatchSnapshot is emitted on each poll.
type WatchSnapshot struct {
	Branch  string        `json:"branch"`
	HeadSHA string        `json:"headSha"`
	Runs    []WorkflowRun `json:"runs"`
	Usage   ActionsUsage  `json:"usage"`
	Phase   string        `json:"phase"` // appearing | watching | done | timeout
	Message string        `json:"message"`
}

// WatchResult is the final post-push observation.
type WatchResult struct {
	Branch   string        `json:"branch"`
	HeadSHA  string        `json:"headSha"`
	Runs     []WorkflowRun `json:"runs"`
	Usage    ActionsUsage  `json:"usage"`
	TimedOut bool          `json:"timedOut"`
	Message  string        `json:"message"`
}

// WatchAfterPush polls for workflow runs of branch/SHA and waits for settlement.
func (c *Client) WatchAfterPush(ctx context.Context, opts WatchOptions) (*WatchResult, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	appear := opts.AppearTimeout
	if appear <= 0 {
		appear = 45 * time.Second
	}
	settle := opts.SettleTimeout
	if settle <= 0 {
		settle = 5 * time.Minute
	}
	interval := opts.PollInterval
	if interval <= 0 {
		interval = 5 * time.Second
	}
	branch := strings.TrimSpace(opts.Branch)
	sha := strings.TrimSpace(opts.HeadSHA)

	emit := func(phase, msg string, runs []WorkflowRun) {
		if opts.OnUpdate == nil {
			return
		}
		opts.OnUpdate(WatchSnapshot{
			Branch:  branch,
			HeadSHA: sha,
			Runs:    runs,
			Usage:   c.UsageForRuns(runs, c.RemoteOwner()),
			Phase:   phase,
			Message: msg,
		})
	}

	// Phase 1: wait until at least one run for SHA (or branch) appears.
	appearDeadline := time.Now().Add(appear)
	var matched []WorkflowRun
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		runs, err := c.ListRuns(ListFilter{Branch: branch, Limit: 20, HeadSHA: sha})
		if err != nil {
			return nil, err
		}
		matched = runs
		if len(matched) == 0 && sha != "" {
			// SHA filter may be too strict if gh list is stale — also try branch-only and filter locally.
			all, listErr := c.ListRuns(ListFilter{Branch: branch, Limit: 20})
			if listErr == nil {
				matched = applyListFilter(all, ListFilter{HeadSHA: sha})
			}
		}
		if len(matched) > 0 {
			emit("watching", fmt.Sprintf("%d run(s) encontrados — acompanhando", len(matched)), matched)
			break
		}
		emit("appearing", "Aguardando Actions aparecerem para este push…", nil)
		if time.Now().After(appearDeadline) {
			usage := c.UsageForRuns(nil, c.RemoteOwner())
			msg := "nenhum run apareceu a tempo (CI pode estar atrasada ou ausente neste repo)"
			emit("timeout", msg, nil)
			return &WatchResult{
				Branch:   branch,
				HeadSHA:  sha,
				Usage:    usage,
				TimedOut: true,
				Message:  msg,
			}, nil
		}
		if err := sleepCtx(ctx, interval); err != nil {
			return nil, err
		}
	}

	// Phase 2: wait until all matched runs are terminal (or timeout).
	settleDeadline := time.Now().Add(settle)
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		runs, err := c.ListRuns(ListFilter{Branch: branch, Limit: 20, HeadSHA: sha})
		if err != nil {
			return nil, err
		}
		if len(runs) == 0 {
			runs = matched
		} else {
			matched = runs
		}
		if allTerminal(matched) {
			usage := c.UsageForRuns(matched, c.RemoteOwner())
			msg := summarizeWatch(matched)
			emit("done", msg, matched)
			return &WatchResult{
				Branch:  branch,
				HeadSHA: sha,
				Runs:    matched,
				Usage:   usage,
				Message: msg,
			}, nil
		}
		pending := countPending(matched)
		emit("watching", fmt.Sprintf("CI em andamento (%d pending)…", pending), matched)
		if time.Now().After(settleDeadline) {
			usage := c.UsageForRuns(matched, c.RemoteOwner())
			msg := "timeout acompanhando CI — runs ainda em andamento"
			emit("timeout", msg, matched)
			return &WatchResult{
				Branch:   branch,
				HeadSHA:  sha,
				Runs:     matched,
				Usage:    usage,
				TimedOut: true,
				Message:  msg,
			}, nil
		}
		if err := sleepCtx(ctx, interval); err != nil {
			return nil, err
		}
	}
}

func allTerminal(runs []WorkflowRun) bool {
	if len(runs) == 0 {
		return false
	}
	for _, r := range runs {
		if !isTerminal(r.Status, r.Conclusion) {
			return false
		}
	}
	return true
}

func isTerminal(status, conclusion string) bool {
	s := normalize(status)
	if s == "completed" || s == "cancelled" || s == "failure" || s == "success" {
		return true
	}
	c := normalize(conclusion)
	return c != "" && c != "null"
}

func countPending(runs []WorkflowRun) int {
	n := 0
	for _, r := range runs {
		if !isTerminal(r.Status, r.Conclusion) {
			n++
		}
	}
	return n
}

func summarizeWatch(runs []WorkflowRun) string {
	pass, fail, other := 0, 0, 0
	for _, r := range runs {
		switch normalize(r.Conclusion) {
		case "success":
			pass++
		case "failure", "cancelled", "timed_out", "startup_failure":
			fail++
		default:
			other++
		}
	}
	switch {
	case fail > 0:
		return fmt.Sprintf("CI: %d falhou · %d ok", fail, pass)
	case pass > 0 && other == 0:
		return fmt.Sprintf("CI: %d ok", pass)
	default:
		return fmt.Sprintf("CI: %d run(s)", len(runs))
	}
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
