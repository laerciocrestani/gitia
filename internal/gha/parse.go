package gha

import (
	"encoding/json"
	"strings"
	"time"
)

type ghRunListItem struct {
	DatabaseID         int64  `json:"databaseId"`
	Name               string `json:"name"`
	DisplayTitle       string `json:"displayTitle"`
	Event              string `json:"event"`
	Status             string `json:"status"`
	Conclusion         string `json:"conclusion"`
	HeadBranch         string `json:"headBranch"`
	HeadSHA            string `json:"headSha"`
	URL                string `json:"url"`
	WorkflowDatabaseID int64  `json:"workflowDatabaseId"`
	WorkflowName       string `json:"workflowName"`
	CreatedAt          string `json:"createdAt"`
	UpdatedAt          string `json:"updatedAt"`
	StartedAt          string `json:"startedAt"`
}

type ghRunView struct {
	DatabaseID         int64   `json:"databaseId"`
	Name               string  `json:"name"`
	DisplayTitle       string  `json:"displayTitle"`
	Event              string  `json:"event"`
	Status             string  `json:"status"`
	Conclusion         string  `json:"conclusion"`
	HeadBranch         string  `json:"headBranch"`
	HeadSHA            string  `json:"headSha"`
	URL                string  `json:"url"`
	WorkflowDatabaseID int64   `json:"workflowDatabaseId"`
	WorkflowName       string  `json:"workflowName"`
	CreatedAt          string  `json:"createdAt"`
	UpdatedAt          string  `json:"updatedAt"`
	StartedAt          string  `json:"startedAt"`
	Jobs               []ghJob `json:"jobs"`
}

type ghJob struct {
	DatabaseID int64    `json:"databaseId"`
	Name       string   `json:"name"`
	Status     string   `json:"status"`
	Conclusion string   `json:"conclusion"`
	Steps      []ghStep `json:"steps"`
}

type ghStep struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Conclusion string `json:"conclusion"`
	Number     int    `json:"number"`
}

func parseRunListJSON(out string) ([]WorkflowRun, error) {
	out = strings.TrimSpace(out)
	if out == "" || out == "null" {
		return nil, nil
	}
	var raw []ghRunListItem
	if err := json.Unmarshal([]byte(out), &raw); err != nil {
		return nil, err
	}
	runs := make([]WorkflowRun, 0, len(raw))
	for _, item := range raw {
		runs = append(runs, item.toRun())
	}
	return runs, nil
}

func parseRunViewJSON(out string) (*RunDetail, error) {
	out = strings.TrimSpace(out)
	if out == "" || out == "null" {
		return nil, nil
	}
	var raw ghRunView
	if err := json.Unmarshal([]byte(out), &raw); err != nil {
		return nil, err
	}
	detail := &RunDetail{Run: raw.toRun()}
	for _, j := range raw.Jobs {
		job := Job{
			ID:         j.DatabaseID,
			Name:       j.Name,
			Status:     j.Status,
			Conclusion: j.Conclusion,
		}
		for _, s := range j.Steps {
			job.Steps = append(job.Steps, Step{
				Name:       s.Name,
				Status:     s.Status,
				Conclusion: s.Conclusion,
				Number:     s.Number,
			})
		}
		detail.Jobs = append(detail.Jobs, job)
	}
	return detail, nil
}

func (r ghRunListItem) toRun() WorkflowRun {
	return WorkflowRun{
		ID:           r.DatabaseID,
		Name:         firstNonEmpty(r.DisplayTitle, r.Name, r.WorkflowName),
		DisplayTitle: r.DisplayTitle,
		Event:        r.Event,
		Status:       r.Status,
		Conclusion:   r.Conclusion,
		HeadBranch:   r.HeadBranch,
		HeadSHA:      r.HeadSHA,
		URL:          r.URL,
		WorkflowID:   r.WorkflowDatabaseID,
		WorkflowName: r.WorkflowName,
		CreatedAt:    parseTime(r.CreatedAt),
		UpdatedAt:    parseTime(r.UpdatedAt),
		StartedAt:    parseTime(r.StartedAt),
	}
}

func (r ghRunView) toRun() WorkflowRun {
	return WorkflowRun{
		ID:           r.DatabaseID,
		Name:         firstNonEmpty(r.DisplayTitle, r.Name, r.WorkflowName),
		DisplayTitle: r.DisplayTitle,
		Event:        r.Event,
		Status:       r.Status,
		Conclusion:   r.Conclusion,
		HeadBranch:   r.HeadBranch,
		HeadSHA:      r.HeadSHA,
		URL:          r.URL,
		WorkflowID:   r.WorkflowDatabaseID,
		WorkflowName: r.WorkflowName,
		CreatedAt:    parseTime(r.CreatedAt),
		UpdatedAt:    parseTime(r.UpdatedAt),
		StartedAt:    parseTime(r.StartedAt),
	}
}

func parseTime(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t
	}
	return time.Time{}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func applyListFilter(runs []WorkflowRun, f ListFilter) []WorkflowRun {
	if len(runs) == 0 {
		return runs
	}
	out := make([]WorkflowRun, 0, len(runs))
	sha := strings.ToLower(strings.TrimSpace(f.HeadSHA))
	for _, r := range runs {
		if sha != "" && !strings.HasPrefix(strings.ToLower(r.HeadSHA), sha) {
			continue
		}
		if f.FailedOnly && !Failed(r.Status, r.Conclusion) {
			continue
		}
		out = append(out, r)
	}
	return out
}

func clampLimit(n int) int {
	if n <= 0 {
		return 20
	}
	if n > 50 {
		return 50
	}
	return n
}
