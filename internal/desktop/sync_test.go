package desktop

import (
	"testing"
)

func TestHygieneActionsCatalog(t *testing.T) {
	actions := HygieneActions()
	if len(actions) != 2 {
		t.Fatalf("actions=%d", len(actions))
	}
	if actions[0].ID != "full" || actions[1].ID != "local" {
		t.Fatalf("unexpected order: %+v", actions)
	}
}

func TestRunSync_emptyPath(t *testing.T) {
	if _, err := RunSync("", "main"); err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestRunHygiene_emptyPath(t *testing.T) {
	if _, err := RunHygiene("", "full", "main"); err == nil {
		t.Fatal("expected error for empty path")
	}
}
