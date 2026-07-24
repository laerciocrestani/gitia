package docker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectComposeFile(t *testing.T) {
	dir := t.TempDir()
	if got := DetectComposeFile(dir); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}

	path := filepath.Join(dir, "compose.yaml")
	if err := os.WriteFile(path, []byte("services: {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := DetectComposeFile(dir); got != path {
		t.Fatalf("got %q want %q", got, path)
	}
}

func TestOverviewSummaryLine(t *testing.T) {
	ov := &Overview{Available: false}
	if ov.SummaryLine() != "n/a" {
		t.Fatalf("expected n/a")
	}

	ov = &Overview{Available: true, DaemonRunning: false}
	if ov.SummaryLine() != "off" {
		t.Fatalf("expected off")
	}

	ov = &Overview{
		Available:     true,
		DaemonRunning: true,
		ComposeFile:   "/tmp/compose.yaml",
		Containers: []ContainerSummary{
			{Service: "app", State: "running"},
		},
	}
	if ov.SummaryLine() != "ok" {
		t.Fatalf("expected ok, got %q", ov.SummaryLine())
	}

	ov = &Overview{
		Available:     true,
		DaemonRunning: true,
		ComposeFile:   "/tmp/compose.yaml",
	}
	if ov.SummaryLine() != "stopped" {
		t.Fatalf("expected stopped, got %q", ov.SummaryLine())
	}
}

func TestLoadOverview_detectsComposeWhenDaemonDown(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "compose.yaml")
	if err := os.WriteFile(path, []byte("services: {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Even without a running daemon (or docker binary in CI), ComposeFile must
	// still be populated so Doctor/UI can warn that the repo has Docker stopped.
	ov := LoadOverview(dir)
	if ov.ComposeFile != path {
		t.Fatalf("ComposeFile=%q want %q", ov.ComposeFile, path)
	}
}

func TestHasRunningContainers(t *testing.T) {
	if HasRunningContainers([]ContainerSummary{{State: "exited"}}) {
		t.Fatal("expected false")
	}
	if !HasRunningContainers([]ContainerSummary{{State: "running"}}) {
		t.Fatal("expected true")
	}
}
