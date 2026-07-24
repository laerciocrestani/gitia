package desktop

import (
	"strings"
	"testing"
)

func TestFormatDockerSnapshotNoCompose(t *testing.T) {
	dir := t.TempDir()
	out := formatDockerSnapshot(dir)
	if !strings.Contains(out, "## Snapshot Docker") {
		t.Fatalf("missing header: %s", out)
	}
}

func TestBuildProjectChatSystemPromptIncludesSkillsSection(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	out := BuildProjectChatSystemPrompt(tmp)
	if !strings.Contains(out, "docker-debug") && !strings.Contains(out, "Skills ativas") {
		// Builtin skill should load even with empty HOME config dir.
		t.Fatalf("expected skills or docker-debug in prompt:\n%s", out)
	}
	if !strings.Contains(out, "## Snapshot Docker") {
		t.Fatalf("missing docker snapshot:\n%s", out)
	}
}
