package gha

import (
	"fmt"
	"strconv"
	"strings"
)

// ListRuns returns recent workflow runs for the repository.
func (c *Client) ListRuns(f ListFilter) ([]WorkflowRun, error) {
	limit := clampLimit(f.Limit)
	args := []string{
		"run", "list",
		"--limit", strconv.Itoa(limit),
		"--json", "databaseId,name,displayTitle,event,status,conclusion,headBranch,headSha,url,createdAt,updatedAt,startedAt,workflowDatabaseId,workflowName",
	}
	if branch := strings.TrimSpace(f.Branch); branch != "" {
		args = append(args, "--branch", branch)
	}

	out, err := c.run(args...)
	if err != nil {
		return nil, err
	}
	runs, err := parseRunListJSON(out)
	if err != nil {
		return nil, fmt.Errorf("parse run list: %w", err)
	}
	return applyListFilter(runs, f), nil
}
