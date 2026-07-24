package redact

import (
	"regexp"
	"strings"
)

const Placeholder = "***REDACTED***"

var (
	reGitHubPAT = regexp.MustCompile(`(?:ghp_|gho_|ghu_|ghs_|ghr_|github_pat_)[A-Za-z0-9_]{20,}`)
	reAKIA      = regexp.MustCompile(`\bAKIA[0-9A-Z]{16}\b`)
	reSlack     = regexp.MustCompile(`xox[baprs]-[0-9A-Za-z-]{10,}`)
	reBearer    = regexp.MustCompile(`(?i)\bBearer\s+[A-Za-z0-9\-._~+/]+=*`)
	reAuthHdr   = regexp.MustCompile(`(?i)(Authorization:\s*)(\S+)`)
	rePEM       = regexp.MustCompile(`(?s)-----BEGIN [A-Z0-9 ]*PRIVATE KEY-----.*?-----END [A-Z0-9 ]*PRIVATE KEY-----`)
	reKV        = regexp.MustCompile(`(?i)\b([A-Za-z0-9_.-]*(?:secret|token|password|api[_-]?key|credential)[A-Za-z0-9_.-]*)\s*([:=])\s*([^\s"']+)`)
	reKVQuoted  = regexp.MustCompile(`(?i)\b([A-Za-z0-9_.-]*(?:secret|token|password|api[_-]?key|credential)[A-Za-z0-9_.-]*)\s*([:=])\s*(["'])([^"']*)["']`)
)

// Secrets redacts common secret patterns from text (ADR-008).
func Secrets(s string) string {
	if s == "" {
		return s
	}
	out := s
	out = rePEM.ReplaceAllString(out, Placeholder)
	out = reGitHubPAT.ReplaceAllString(out, Placeholder)
	out = reAKIA.ReplaceAllString(out, Placeholder)
	out = reSlack.ReplaceAllString(out, Placeholder)
	out = reBearer.ReplaceAllString(out, "Bearer "+Placeholder)
	out = reAuthHdr.ReplaceAllString(out, "${1}"+Placeholder)
	out = reKVQuoted.ReplaceAllString(out, "${1}${2}${3}"+Placeholder+"${3}")
	out = reKV.ReplaceAllString(out, "${1}${2}"+Placeholder)
	return out
}

// Useful reports whether redacted text still has non-trivial content for AI/UI.
func Useful(redacted string) bool {
	t := strings.TrimSpace(redacted)
	if t == "" {
		return false
	}
	stripped := strings.ReplaceAll(t, Placeholder, "")
	stripped = strings.TrimSpace(stripped)
	return len(stripped) >= 8
}
