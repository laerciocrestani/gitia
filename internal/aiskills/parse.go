package aiskills

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type skillFrontmatter struct {
	ID          string `yaml:"id"`
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

// ParseSkillMarkdown parses YAML frontmatter + markdown body.
func ParseSkillMarkdown(raw string) (*Skill, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("skill vazia")
	}

	var fmRaw, body string
	if strings.HasPrefix(raw, "---") {
		rest := strings.TrimPrefix(raw, "---")
		rest = strings.TrimLeft(rest, "\r\n")
		idx := strings.Index(rest, "\n---")
		if idx < 0 {
			return nil, fmt.Errorf("frontmatter YAML sem fechamento ---")
		}
		fmRaw = rest[:idx]
		body = strings.TrimSpace(rest[idx+len("\n---"):])
		body = strings.TrimLeft(body, "\r\n")
	} else {
		return nil, fmt.Errorf("skill deve começar com frontmatter YAML (---)")
	}

	var fm skillFrontmatter
	if err := yaml.Unmarshal([]byte(fmRaw), &fm); err != nil {
		return nil, fmt.Errorf("parse frontmatter: %w", err)
	}
	s := &Skill{
		ID:          strings.TrimSpace(fm.ID),
		Name:        strings.TrimSpace(fm.Name),
		Description: strings.TrimSpace(fm.Description),
		Body:        strings.TrimSpace(body),
		Enabled:     true,
	}
	normalizeSkill(s)
	if err := ValidateSkill(s.ID, s.Name, s.Body); err != nil {
		return nil, err
	}
	return s, nil
}

// EncodeSkillMarkdown writes a skill file.
func EncodeSkillMarkdown(s Skill) ([]byte, error) {
	normalizeSkill(&s)
	if err := ValidateSkill(s.ID, s.Name, s.Body); err != nil {
		return nil, err
	}
	fm := skillFrontmatter{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
	}
	fmBytes, err := yaml.Marshal(&fm)
	if err != nil {
		return nil, err
	}
	var b strings.Builder
	b.WriteString("---\n")
	b.Write(fmBytes)
	if !strings.HasSuffix(b.String(), "\n") {
		b.WriteByte('\n')
	}
	b.WriteString("---\n\n")
	b.WriteString(strings.TrimSpace(s.Body))
	b.WriteByte('\n')
	return []byte(b.String()), nil
}
