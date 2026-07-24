package gha

import (
	"errors"
	"strings"
	"testing"
)

func TestSummaryFromErrorUsesCacheOffline(t *testing.T) {
	cached := &Summary{State: SummaryPass, Label: "CI 2✓", Pass: 2}
	got := SummaryFromError("/tmp/repo", &Error{Kind: ErrNetwork, Message: "net"}, cached)
	if got == nil {
		t.Fatal("expected summary")
	}
	if !got.FromCache {
		t.Fatal("expected FromCache")
	}
	if got.State != SummaryOffline {
		t.Fatalf("state=%q want offline", got.State)
	}
	if !strings.Contains(got.Label, "off") {
		t.Fatalf("label=%q", got.Label)
	}
}

func TestSummaryFromErrorAuthUnavailable(t *testing.T) {
	got := SummaryFromError("/tmp/repo", &Error{Kind: ErrGhAuth, Message: "auth"}, nil)
	if got.State != SummaryUnavailable {
		t.Fatalf("state=%q", got.State)
	}
	if !errors.Is(&Error{Kind: ErrGhAuth}, ErrGhAuth) {
		t.Fatal("Error.Is should match Kind")
	}
}

func TestSummaryFromErrorCachedAuthKeepsLabel(t *testing.T) {
	cached := &Summary{State: SummaryFail, Label: "CI 1✗", Fail: 1}
	got := SummaryFromError("/tmp/repo", &Error{Kind: ErrGhAuth, Message: "auth"}, cached)
	if !got.FromCache {
		t.Fatal("expected cache")
	}
	if got.Label != "CI 1✗" {
		t.Fatalf("label=%q", got.Label)
	}
}
