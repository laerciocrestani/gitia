package aiskills

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	IndexVersion     = 1
	maxSkillBodyRunes = 40_000
	maxPromptRunes    = 12_000
)

var idPattern = regexp.MustCompile(`^[a-z][a-z0-9-]{1,62}$`)

// Skill is a chat playbook injected into the project assistant.
type Skill struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Body        string `json:"body"`
	Enabled     bool   `json:"enabled"`
	Builtin     bool   `json:"builtin"`
	Customized  bool   `json:"customized"` // user override file exists
}

// Index persists enable/disable flags for known skills.
type Index struct {
	Version  int      `yaml:"version" json:"version"`
	Disabled []string `yaml:"disabled,omitempty" json:"disabled,omitempty"`
}

// ValidateID checks skill id format.
func ValidateID(id string) error {
	id = strings.TrimSpace(id)
	if !idPattern.MatchString(id) {
		return fmt.Errorf("id inválido %q — use a-z, 0-9 e hífen (ex.: docker-debug)", id)
	}
	return nil
}

// ValidateSkill checks editable fields before save.
func ValidateSkill(id, name, body string) error {
	if err := ValidateID(id); err != nil {
		return err
	}
	if strings.TrimSpace(name) == "" {
		return fmt.Errorf("nome é obrigatório")
	}
	if strings.TrimSpace(body) == "" {
		return fmt.Errorf("corpo da skill é obrigatório")
	}
	if utf8.RuneCountInString(body) > maxSkillBodyRunes {
		return fmt.Errorf("corpo da skill excede %d caracteres", maxSkillBodyRunes)
	}
	return nil
}

// FormatPromptSection renders enabled skills for the chat system prompt.
func FormatPromptSection(skills []Skill) string {
	var parts []string
	total := 0
	for _, s := range skills {
		if !s.Enabled {
			continue
		}
		body := strings.TrimSpace(s.Body)
		if body == "" {
			continue
		}
		name := strings.TrimSpace(s.Name)
		if name == "" {
			name = s.ID
		}
		block := fmt.Sprintf("### %s (`%s`)\n%s", name, s.ID, body)
		n := utf8.RuneCountInString(block)
		if total+n > maxPromptRunes {
			remain := maxPromptRunes - total
			if remain < 80 {
				break
			}
			runes := []rune(block)
			block = string(runes[:remain]) + "\n… [skill truncada]"
			parts = append(parts, block)
			break
		}
		parts = append(parts, block)
		total += n + 2
	}
	if len(parts) == 0 {
		return ""
	}
	return "## Skills ativas\n\n" + strings.Join(parts, "\n\n")
}

func normalizeSkill(s *Skill) {
	if s == nil {
		return
	}
	s.ID = strings.TrimSpace(s.ID)
	s.Name = strings.TrimSpace(s.Name)
	s.Description = strings.TrimSpace(s.Description)
	s.Body = strings.TrimSpace(s.Body)
	if s.Name == "" {
		s.Name = s.ID
	}
}
