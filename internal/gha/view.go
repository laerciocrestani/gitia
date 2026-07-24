package gha

import (
	"fmt"
	"strconv"
)

// ViewRun returns a workflow run with jobs/steps.
func (c *Client) ViewRun(runID int64) (*RunDetail, error) {
	if runID <= 0 {
		return nil, fmt.Errorf("run id inválido")
	}
	args := []string{
		"run", "view", strconv.FormatInt(runID, 10),
		"--json", "databaseId,name,displayTitle,event,status,conclusion,headBranch,headSha,url,createdAt,updatedAt,startedAt,workflowDatabaseId,workflowName,jobs",
	}
	out, stderr, err := c.runAllowExit(args...)
	if out == "" {
		if err != nil {
			return nil, classifyGhErr(stderr, err)
		}
		return nil, &Error{Kind: ErrNotFound, Message: "run sem dados"}
	}
	detail, parseErr := parseRunViewJSON(out)
	if parseErr != nil {
		return nil, fmt.Errorf("parse run view: %w", parseErr)
	}
	if detail == nil {
		return nil, &Error{Kind: ErrNotFound, Message: "run não encontrado"}
	}
	return detail, nil
}
