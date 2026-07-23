package components_test

import (
	"strings"
	"testing"

	"github.com/laerciocrestani/openbench/internal/app"
	"github.com/laerciocrestani/openbench/internal/tui/components"
)

func TestHygieneModeCatalog(t *testing.T) {
	t.Parallel()
	modes := components.HygieneModeCatalog()
	if len(modes) != 2 {
		t.Fatalf("expected 2 modes, got %d", len(modes))
	}
	if modes[0].Mode != app.HygieneModeFull {
		t.Fatalf("first mode should be full")
	}
	if modes[1].Flag != "--local" {
		t.Fatalf("second flag = %q", modes[1].Flag)
	}
}

func TestRenderHygieneOptionsPanel(t *testing.T) {
	modes := components.HygieneModeCatalog()
	out := components.RenderHygieneOptionsPanel(0, modes, "main", false, 90)
	for _, want := range []string{"Hygiene · Options", "Full hygiene", "Base: main", "git fetch origin --prune", "--full", "--local"} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
}
