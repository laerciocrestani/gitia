package aiskills

import (
	"strings"
	"testing"
)

func TestParseAndEncodeRoundTrip(t *testing.T) {
	raw := `---
id: docker-debug
name: Debug Docker
description: teste
---

Corpo da skill.
`
	s, err := ParseSkillMarkdown(raw)
	if err != nil {
		t.Fatal(err)
	}
	if s.ID != "docker-debug" || s.Name != "Debug Docker" {
		t.Fatalf("unexpected skill: %+v", s)
	}
	if !strings.Contains(s.Body, "Corpo da skill") {
		t.Fatalf("body=%q", s.Body)
	}
	out, err := EncodeSkillMarkdown(*s)
	if err != nil {
		t.Fatal(err)
	}
	s2, err := ParseSkillMarkdown(string(out))
	if err != nil {
		t.Fatal(err)
	}
	if s2.ID != s.ID || s2.Body != s.Body {
		t.Fatalf("round-trip mismatch: %+v vs %+v", s, s2)
	}
}

func TestLoadBuiltinDockerDebug(t *testing.T) {
	all, err := loadBuiltinSkills()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) == 0 {
		t.Fatal("expected builtin skills")
	}
	found := false
	for _, s := range all {
		if s.ID == "docker-debug" {
			found = true
			if !s.Builtin {
				t.Fatal("expected builtin flag")
			}
			if !strings.Contains(s.Body, "Ordem de diagnóstico") {
				t.Fatalf("unexpected body: %s", s.Body[:min(80, len(s.Body))])
			}
		}
	}
	if !found {
		t.Fatal("docker-debug missing")
	}
}

func TestFormatPromptSection(t *testing.T) {
	section := FormatPromptSection([]Skill{
		{ID: "a", Name: "A", Body: "corpo", Enabled: true},
		{ID: "b", Name: "B", Body: "off", Enabled: false},
	})
	if !strings.Contains(section, "## Skills ativas") {
		t.Fatalf("missing header: %s", section)
	}
	if !strings.Contains(section, "`a`") || strings.Contains(section, "`b`") {
		t.Fatalf("enabled filter failed: %s", section)
	}
}

func TestValidateID(t *testing.T) {
	if err := ValidateID("docker-debug"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateID("Bad_ID"); err == nil {
		t.Fatal("expected error")
	}
}
