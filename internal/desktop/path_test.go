package desktop

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckOnboarding_ghNotBlockingForNeedsOnboarding(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", "/nonexistent/bin") // force LookPath("gh") to fail
	dir := filepath.Join(home, ".config", "openbench")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := SaveAIConfig("openrouter", "sk-test-key-12345678", "deepseek/deepseek-chat"); err != nil {
		t.Fatal(err)
	}

	st, err := CheckOnboarding("")
	if err != nil {
		t.Fatal(err)
	}
	if !st.APIKeyOK {
		t.Fatalf("expected api key ok: %+v", st)
	}
	if st.NeedsOnboarding {
		t.Fatalf("gh missing should not set NeedsOnboarding: %+v", st)
	}
	if st.GhInstalled {
		t.Fatal("expected gh not installed")
	}
	found := false
	for _, iss := range st.Issues {
		if iss.ID == "gh_install" {
			found = true
			if iss.Blocking {
				t.Fatal("gh_install must not be blocking")
			}
		}
	}
	if !found {
		t.Fatalf("expected gh_install issue: %+v", st.Issues)
	}
}

func TestAugmentUserPath_idempotent(t *testing.T) {
	AugmentUserPath()
	AugmentUserPath() // must not panic
}
