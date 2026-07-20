package desktop

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPrefsRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	// Force config under temp home
	_ = os.MkdirAll(filepath.Join(dir, ".config", "openbench"), 0o755)

	if err := SavePrefs(Prefs{ValidateCommit: true, LastProject: "/tmp/demo"}); err != nil {
		t.Fatal(err)
	}
	got, err := LoadPrefs()
	if err != nil {
		t.Fatal(err)
	}
	if !got.ValidateCommit || got.LastProject != "/tmp/demo" {
		t.Fatalf("got %+v", got)
	}
}
