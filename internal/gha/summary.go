package gha

import (
	"errors"
	"fmt"
	"strings"
)

// SummaryState values for CI badge / cache.
const (
	SummaryPass        = "pass"
	SummaryFail        = "fail"
	SummaryPending     = "pending"
	SummaryUnknown     = "unknown"
	SummaryOffline     = "offline"
	SummaryUnavailable = "unavailable"
)

// Summary is a lightweight Actions status for StatusHub / Doctor / cache.
type Summary struct {
	State      string `json:"state"`
	Label      string `json:"label"`
	Pass       int    `json:"pass"`
	Fail       int    `json:"fail"`
	Pending    int    `json:"pending"`
	Host       string `json:"host"`
	Branch     string `json:"branch,omitempty"`
	FromCache  bool   `json:"fromCache,omitempty"`
	Message    string `json:"message,omitempty"`
	Enterprise bool   `json:"enterprise,omitempty"`
}

// LoadSummary lists recent runs for the current branch and builds a badge summary.
func (c *Client) LoadSummary(branch string) (*Summary, error) {
	host := ResolveHost(c.dir)
	sum := &Summary{
		Host:       host,
		Branch:     branch,
		Enterprise: IsEnterpriseHost(host),
		State:      SummaryUnknown,
		Label:      "CI ?",
	}
	f := ListFilter{Branch: strings.TrimSpace(branch), Limit: 8}
	runs, err := c.ListRuns(f)
	if err != nil {
		return nil, err
	}
	if len(runs) == 0 {
		sum.State = SummaryUnknown
		sum.Label = "CI —"
		sum.Message = "nenhum run recente nesta branch"
		return sum, nil
	}
	for _, r := range runs {
		switch {
		case Failed(r.Status, r.Conclusion):
			sum.Fail++
		case normalize(r.Conclusion) == "success":
			sum.Pass++
		case !isTerminal(r.Status, r.Conclusion):
			sum.Pending++
		default:
			// neutral/skipped count as pass-ish for badge
			sum.Pass++
		}
	}
	switch {
	case sum.Fail > 0:
		sum.State = SummaryFail
		sum.Label = fmt.Sprintf("CI %d✗", sum.Fail)
	case sum.Pending > 0:
		sum.State = SummaryPending
		sum.Label = fmt.Sprintf("CI %d…", sum.Pending)
	case sum.Pass > 0:
		sum.State = SummaryPass
		sum.Label = fmt.Sprintf("CI %d✓", sum.Pass)
	default:
		sum.State = SummaryUnknown
		sum.Label = "CI ?"
	}
	if sum.Enterprise {
		sum.Message = "host " + host
	}
	return sum, nil
}

// SummaryFromError maps classified errors to an offline/unavailable summary.
func SummaryFromError(dir string, err error, cached *Summary) *Summary {
	host := ResolveHost(dir)
	sum := &Summary{
		Host:       host,
		Enterprise: IsEnterpriseHost(host),
		State:      SummaryUnavailable,
		Label:      "CI !",
		Message:    err.Error(),
	}
	if cached != nil {
		out := *cached
		out.FromCache = true
		out.Host = host
		out.Enterprise = IsEnterpriseHost(host)
		switch {
		case errorsIs(err, ErrNetwork):
			out.State = SummaryOffline
			out.Label = cached.Label + " (off)"
			out.Message = "offline — último status em cache"
		case errorsIs(err, ErrGhAuth), errorsIs(err, ErrGhNotInstalled):
			out.State = SummaryUnavailable
			out.Message = err.Error()
		default:
			out.Message = err.Error() + " · cache"
		}
		return &out
	}
	if errorsIs(err, ErrNetwork) {
		sum.State = SummaryOffline
		sum.Label = "CI off"
		sum.Message = "sem rede — sem cache de CI"
	}
	return sum
}

func errorsIs(err, target error) bool {
	return errors.Is(err, target)
}
