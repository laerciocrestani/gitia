package desktop

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveAIConfig_roundTrip(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
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
	if st.Provider != "openrouter" {
		t.Fatalf("provider=%q", st.Provider)
	}
}
