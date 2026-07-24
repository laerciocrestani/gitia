package aiskills

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/laerciocrestani/openbench/internal/config"
	"gopkg.in/yaml.v3"
)

const (
	skillsSubdir = "skills"
	indexFile    = "index.yaml"
	fileSuffix   = ".skill.md"
)

// Dir returns ~/.config/openbench/skills.
func Dir() (string, error) {
	base, err := config.DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, skillsSubdir), nil
}

// EnsureDir creates the skills directory.
func EnsureDir() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func indexPath() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, indexFile), nil
}

func userSkillPath(id string) (string, error) {
	if err := ValidateID(id); err != nil {
		return "", err
	}
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, id+fileSuffix), nil
}

func loadIndex() (*Index, error) {
	path, err := indexPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Index{Version: IndexVersion}, nil
		}
		return nil, err
	}
	var idx Index
	if err := yaml.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if idx.Version == 0 {
		idx.Version = IndexVersion
	}
	idx.Disabled = normalizeIDs(idx.Disabled)
	return &idx, nil
}

func saveIndex(idx *Index) error {
	if idx == nil {
		idx = &Index{Version: IndexVersion}
	}
	if idx.Version == 0 {
		idx.Version = IndexVersion
	}
	idx.Disabled = normalizeIDs(idx.Disabled)
	if _, err := EnsureDir(); err != nil {
		return err
	}
	path, err := indexPath()
	if err != nil {
		return err
	}
	data, err := yaml.Marshal(idx)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func normalizeIDs(ids []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	sort.Strings(out)
	return out
}

func isDisabled(idx *Index, id string) bool {
	if idx == nil {
		return false
	}
	for _, d := range idx.Disabled {
		if d == id {
			return true
		}
	}
	return false
}

// List returns merged builtin + user skills (sorted by id).
func List() ([]Skill, error) {
	idx, err := loadIndex()
	if err != nil {
		return nil, err
	}
	builtins, err := loadBuiltinSkills()
	if err != nil {
		return nil, err
	}
	byID := make(map[string]Skill, len(builtins)+4)
	for _, s := range builtins {
		s.Enabled = !isDisabled(idx, s.ID)
		byID[s.ID] = s
	}

	dir, err := Dir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), fileSuffix) {
			continue
		}
		id := strings.TrimSuffix(e.Name(), fileSuffix)
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, err
		}
		s, err := ParseSkillMarkdown(string(data))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		if s.ID == "" {
			s.ID = id
		}
		if s.ID != id {
			// Prefer filename as canonical id on disk.
			s.ID = id
		}
		base, isBuiltin := byID[s.ID]
		s.Builtin = isBuiltin
		s.Customized = true
		s.Enabled = !isDisabled(idx, s.ID)
		if isBuiltin {
			// Keep builtin flag; content is user override.
			_ = base
		}
		byID[s.ID] = *s
	}

	out := make([]Skill, 0, len(byID))
	for _, s := range byID {
		out = append(out, s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// Get returns one skill by id.
func Get(id string) (*Skill, error) {
	id = strings.TrimSpace(id)
	if err := ValidateID(id); err != nil {
		return nil, err
	}
	all, err := List()
	if err != nil {
		return nil, err
	}
	for i := range all {
		if all[i].ID == id {
			s := all[i]
			return &s, nil
		}
	}
	return nil, fmt.Errorf("skill %q não encontrada", id)
}

// ListEnabled returns only enabled skills (for prompt injection).
func ListEnabled() ([]Skill, error) {
	all, err := List()
	if err != nil {
		return nil, err
	}
	var out []Skill
	for _, s := range all {
		if s.Enabled {
			out = append(out, s)
		}
	}
	return out, nil
}

// Save writes or updates a user skill file (creates custom or overrides builtin).
func Save(id, name, description, body string) error {
	id = strings.TrimSpace(id)
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)
	body = strings.TrimSpace(body)
	if err := ValidateSkill(id, name, body); err != nil {
		return err
	}
	if _, err := EnsureDir(); err != nil {
		return err
	}
	s := Skill{
		ID:          id,
		Name:        name,
		Description: description,
		Body:        body,
	}
	data, err := EncodeSkillMarkdown(s)
	if err != nil {
		return err
	}
	path, err := userSkillPath(id)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// SetEnabled toggles a skill via index.yaml.
func SetEnabled(id string, enabled bool) error {
	id = strings.TrimSpace(id)
	if err := ValidateID(id); err != nil {
		return err
	}
	if _, err := Get(id); err != nil {
		return err
	}
	idx, err := loadIndex()
	if err != nil {
		return err
	}
	disabled := make([]string, 0, len(idx.Disabled))
	for _, d := range idx.Disabled {
		if d != id {
			disabled = append(disabled, d)
		}
	}
	if !enabled {
		disabled = append(disabled, id)
	}
	idx.Disabled = disabled
	return saveIndex(idx)
}

// Reset removes a user override for a builtin skill (restores embedded body).
func Reset(id string) error {
	id = strings.TrimSpace(id)
	if err := ValidateID(id); err != nil {
		return err
	}
	if _, err := builtinByID(id); err != nil {
		return fmt.Errorf("só skills builtin podem ser restauradas: %w", err)
	}
	path, err := userSkillPath(id)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	// Re-enable on reset.
	return SetEnabled(id, true)
}

// Delete removes a custom (non-builtin) skill.
func Delete(id string) error {
	id = strings.TrimSpace(id)
	if err := ValidateID(id); err != nil {
		return err
	}
	if _, err := builtinByID(id); err == nil {
		return fmt.Errorf("não é possível apagar skill builtin %q — desative ou restaure o padrão", id)
	}
	path, err := userSkillPath(id)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("skill %q não encontrada", id)
		}
		return err
	}
	idx, err := loadIndex()
	if err != nil {
		return err
	}
	filtered := idx.Disabled[:0]
	for _, d := range idx.Disabled {
		if d != id {
			filtered = append(filtered, d)
		}
	}
	idx.Disabled = filtered
	return saveIndex(idx)
}

// SkillsDirString returns the skills directory path for UI display.
func SkillsDirString() (string, error) {
	return Dir()
}
