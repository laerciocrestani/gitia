package gha

import "time"

// ListFilter controls which workflow runs are returned.
type ListFilter struct {
	Branch     string // empty = all branches
	FailedOnly bool
	Limit      int // default 20, max 50
	HeadSHA    string
}

// WorkflowRun is a GitHub Actions workflow run summary.
type WorkflowRun struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	DisplayTitle string    `json:"displayTitle"`
	Event        string    `json:"event"`
	Status       string    `json:"status"`
	Conclusion   string    `json:"conclusion"`
	HeadBranch   string    `json:"headBranch"`
	HeadSHA      string    `json:"headSha"`
	URL          string    `json:"url"`
	WorkflowID   int64     `json:"workflowId"`
	WorkflowName string    `json:"workflowName"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	StartedAt    time.Time `json:"startedAt,omitempty"`
}

// Job is one job inside a workflow run.
type Job struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	Steps      []Step `json:"steps,omitempty"`
}

// Step is one step inside a job.
type Step struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	Number     int    `json:"number"`
}

// RunDetail is a workflow run with jobs.
type RunDetail struct {
	Run  WorkflowRun `json:"run"`
	Jobs []Job       `json:"jobs"`
}

// UsageState values for ActionsUsage.State.
const (
	UsageStateRun         = "run"
	UsageStateRepoWindow  = "repo_window"
	UsageStateOrg         = "org"
	UsageStateUnavailable = "unavailable"
)

// ActionsUsage is always present in CI views (ADR-008).
type ActionsUsage struct {
	State               string   `json:"state"`
	RunMinutes          *float64 `json:"runMinutes,omitempty"`
	WindowMinutes       *float64 `json:"windowMinutes,omitempty"`
	OrgUsedMinutes      *float64 `json:"orgUsedMinutes,omitempty"`
	OrgIncludedMinutes  *float64 `json:"orgIncludedMinutes,omitempty"`
	OrgRemainingMinutes *float64 `json:"orgRemainingMinutes,omitempty"`
	Message             string   `json:"message"`
}

// StatusSnapshot is the Observe-slice payload for CLI/desktop.
type StatusSnapshot struct {
	Branch  string        `json:"branch"`
	HeadSHA string        `json:"headSha,omitempty"`
	Runs    []WorkflowRun `json:"runs"`
	Usage   ActionsUsage  `json:"usage"`
	Filter  ListFilter    `json:"filter"`
}

// Failed returns true when the run/job/step looks failed for UI filters.
func Failed(status, conclusion string) bool {
	c := normalize(conclusion)
	switch c {
	case "failure", "cancelled", "timed_out", "startup_failure":
		return true
	}
	return false
}

func normalize(s string) string {
	b := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b = append(b, c)
	}
	return string(b)
}
