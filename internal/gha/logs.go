package gha

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/laerciocrestani/openbench/internal/redact"
)

// DefaultMaxLogBytes caps redacted log returned to UI/CLI preview.
const DefaultMaxLogBytes = 256 * 1024

// LogOptions controls on-demand log fetch.
type LogOptions struct {
	RunID      int64
	JobID      int64 // optional; 0 = all jobs in the run
	FailedOnly bool  // use --log-failed
	MaxBytes   int   // 0 = DefaultMaxLogBytes; negative = no truncate
}

// LogPayload is always redacted before leaving the domain (ADR-008).
type LogPayload struct {
	RunID        int64  `json:"runId"`
	JobID        int64  `json:"jobId,omitempty"`
	FailedOnly   bool   `json:"failedOnly"`
	RedactedText string `json:"redactedText"`
	Bytes        int    `json:"bytes"`
	RawBytes     int    `json:"rawBytes"`
	Truncated    bool   `json:"truncated"`
	Useful       bool   `json:"useful"`
	Message      string `json:"message,omitempty"`
}

// FetchLog downloads workflow logs via gh and returns a redacted payload.
func (c *Client) FetchLog(opts LogOptions) (*LogPayload, error) {
	if opts.RunID <= 0 {
		return nil, fmt.Errorf("run id inválido")
	}
	args := []string{"run", "view", strconv.FormatInt(opts.RunID, 10)}
	if opts.FailedOnly {
		args = append(args, "--log-failed")
	} else {
		args = append(args, "--log")
	}
	if opts.JobID > 0 {
		args = append(args, "--job", strconv.FormatInt(opts.JobID, 10))
	}

	out, stderr, err := c.runAllowExit(args...)
	if out == "" {
		if err != nil {
			return nil, classifyGhErr(stderr, err)
		}
		return &LogPayload{
			RunID:      opts.RunID,
			JobID:      opts.JobID,
			FailedOnly: opts.FailedOnly,
			Useful:     false,
			Message:    "log vazio",
		}, nil
	}

	rawBytes := len(out)
	redacted := redact.Secrets(out)
	payload := &LogPayload{
		RunID:        opts.RunID,
		JobID:        opts.JobID,
		FailedOnly:   opts.FailedOnly,
		RawBytes:     rawBytes,
		RedactedText: redacted,
		Bytes:        len(redacted),
		Useful:       redact.Useful(redacted),
	}
	if !payload.Useful {
		payload.Message = "log só com material sensível / vazio após redação"
	}

	max := opts.MaxBytes
	if max == 0 {
		max = DefaultMaxLogBytes
	}
	if max > 0 && len(payload.RedactedText) > max {
		payload.RedactedText = truncateUTF8(payload.RedactedText, max)
		payload.Truncated = true
		payload.Bytes = len(payload.RedactedText)
		if payload.Message == "" {
			payload.Message = fmt.Sprintf("truncado em %d bytes (sob demanda)", max)
		}
	}
	return payload, nil
}

// FailureWindow extracts a redacted window around failed steps for AI (slice E).
// ContextLines is lines before/after each failed marker (default 40).
func FailureWindow(redactedLog string, contextLines int) string {
	if contextLines <= 0 {
		contextLines = 40
	}
	lines := strings.Split(redactedLog, "\n")
	if len(lines) == 0 {
		return ""
	}
	keep := make([]bool, len(lines))
	for i, line := range lines {
		low := strings.ToLower(line)
		if strings.Contains(low, "##[error]") ||
			strings.Contains(low, "error:") ||
			strings.Contains(low, "failed") && (strings.Contains(low, "step") || strings.Contains(low, "process completed")) {
			start := i - contextLines
			if start < 0 {
				start = 0
			}
			end := i + contextLines
			if end >= len(lines) {
				end = len(lines) - 1
			}
			for j := start; j <= end; j++ {
				keep[j] = true
			}
		}
	}
	var b strings.Builder
	for i, line := range lines {
		if keep[i] {
			b.WriteString(line)
			b.WriteByte('\n')
		}
	}
	out := strings.TrimSpace(b.String())
	if out == "" {
		// fallback: last N lines of redacted log
		n := contextLines * 2
		if n > len(lines) {
			n = len(lines)
		}
		out = strings.TrimSpace(strings.Join(lines[len(lines)-n:], "\n"))
	}
	return redact.Secrets(out)
}

func truncateUTF8(s string, maxBytes int) string {
	if maxBytes <= 0 || len(s) <= maxBytes {
		return s
	}
	// cut on rune boundary
	cut := maxBytes
	for cut > 0 && !utf8.ValidString(s[:cut]) {
		cut--
	}
	const note = "\n\n… [log truncado]"
	if cut > len(note) {
		return s[:cut-len(note)] + note
	}
	return s[:cut]
}
