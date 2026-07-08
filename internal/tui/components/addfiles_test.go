package components_test

import (
	"strings"
	"testing"

	gitpkg "github.com/laerciocrestani/gitai/internal/git"
	"github.com/laerciocrestani/gitai/internal/tui/components"
)

func TestSelectMarker(t *testing.T) {
	if !strings.Contains(components.SelectMarker(true), "●") {
		t.Fatalf("selected marker: %q", components.SelectMarker(true))
	}
	if !strings.Contains(components.SelectMarker(false), "( )") {
		t.Fatalf("unselected marker: %q", components.SelectMarker(false))
	}
}

func TestRenderAddTodosLine(t *testing.T) {
	out := components.RenderAddTodosLine(true, true)
	if !strings.Contains(out, "Todos") {
		t.Fatalf("missing Todos: %q", out)
	}
	if !strings.Contains(out, "●") {
		t.Fatalf("missing bullet: %q", out)
	}
}

func TestRenderAddFileLineUsesBullets(t *testing.T) {
	out := components.RenderAddFileLine(true, false, gitpkg.FileChange{
		Path:   "internal/app/stage.go",
		Status: "untracked",
	})
	if strings.Contains(out, "[x]") || strings.Contains(out, "[ ]") {
		t.Fatalf("still using brackets: %q", out)
	}
	if !strings.Contains(out, "●") {
		t.Fatalf("missing bullet: %q", out)
	}
}
