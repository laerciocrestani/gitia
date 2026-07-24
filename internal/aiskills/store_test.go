package aiskills

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListSaveEnableReset(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	// DataDir uses UserHomeDir → ~/.config/openbench
	cfg := filepath.Join(tmp, ".config", "openbench", "skills")
	if err := os.MkdirAll(cfg, 0o755); err != nil {
		t.Fatal(err)
	}

	list, err := List()
	if err != nil {
		t.Fatal(err)
	}
	var docker *Skill
	for i := range list {
		if list[i].ID == "docker-debug" {
			docker = &list[i]
			break
		}
	}
	if docker == nil {
		t.Fatal("builtin docker-debug missing")
	}
	if !docker.Enabled || !docker.Builtin || docker.Customized {
		t.Fatalf("unexpected builtin state: %+v", docker)
	}

	if err := Save("docker-debug", "Debug custom", "desc", "Corpo customizado para teste."); err != nil {
		t.Fatal(err)
	}
	got, err := Get("docker-debug")
	if err != nil {
		t.Fatal(err)
	}
	if !got.Customized || got.Body != "Corpo customizado para teste." {
		t.Fatalf("override failed: %+v", got)
	}

	if err := SetEnabled("docker-debug", false); err != nil {
		t.Fatal(err)
	}
	got, err = Get("docker-debug")
	if err != nil {
		t.Fatal(err)
	}
	if got.Enabled {
		t.Fatal("expected disabled")
	}

	if err := Reset("docker-debug"); err != nil {
		t.Fatal(err)
	}
	got, err = Get("docker-debug")
	if err != nil {
		t.Fatal(err)
	}
	if got.Customized || !got.Enabled {
		t.Fatalf("reset failed: %+v", got)
	}
	if !strings.Contains(got.Body, "Ordem de diagnóstico") {
		t.Fatalf("builtin body not restored")
	}

	if err := Save("minha-skill", "Minha", "d", "corpo da skill custom."); err != nil {
		t.Fatal(err)
	}
	if err := Delete("minha-skill"); err != nil {
		t.Fatal(err)
	}
	if _, err := Get("minha-skill"); err == nil {
		t.Fatal("expected delete")
	}
}
