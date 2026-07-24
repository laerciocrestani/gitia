package gha

import (
	"strings"
	"testing"

	"github.com/laerciocrestani/openbench/internal/redact"
)

func TestFailureWindowAroundError(t *testing.T) {
	var b strings.Builder
	for i := 0; i < 100; i++ {
		b.WriteString("line ok\n")
	}
	b.WriteString("##[error] Process completed with exit code 1.\n")
	b.WriteString("ghp_abcdefghijklmnopqrstuvwxyz0123456789\n")
	for i := 0; i < 100; i++ {
		b.WriteString("line after\n")
	}
	win := FailureWindow(b.String(), 5)
	if !strings.Contains(win, "##[error]") {
		t.Fatalf("missing error: %q", win)
	}
	if strings.Contains(win, "ghp_") {
		t.Fatalf("token leaked: %q", win)
	}
	if !strings.Contains(win, redact.Placeholder) {
		t.Fatalf("expected redaction: %q", win)
	}
	if strings.Count(win, "\n") > 20 {
		t.Fatalf("window too large: %d lines", strings.Count(win, "\n"))
	}
}

func TestTruncateUTF8(t *testing.T) {
	s := strings.Repeat("a", 100) + "文字"
	got := truncateUTF8(s, 50)
	if len(got) > 50 {
		t.Fatalf("len=%d", len(got))
	}
	if !strings.Contains(got, "truncado") {
		t.Fatalf("got %q", got)
	}
}
