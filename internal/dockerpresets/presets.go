package dockerpresets

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// FileVersion is the schema version written to project files.
	FileVersion = 1
	// RelativePath is the presets file inside a project root.
	RelativePath = ".openbench/docker-presets.yaml"
)

// Preset is one runnable docker exec command template.
type Preset struct {
	ID          string `yaml:"id" json:"id"`
	Label       string `yaml:"label" json:"label"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Command     string `yaml:"command" json:"command"`
	Kit         string `yaml:"kit,omitempty" json:"kit,omitempty"`
	Interactive bool   `yaml:"interactive,omitempty" json:"interactive,omitempty"`
}

// File is the on-disk project presets document.
type File struct {
	Version int      `yaml:"version" json:"version"`
	Presets []Preset `yaml:"presets" json:"presets"`
}

// ProjectPath returns the absolute presets path for a project root.
func ProjectPath(projectRoot string) string {
	return filepath.Join(projectRoot, RelativePath)
}

// LoadProject reads presets for a project. Missing file → empty list (not an error).
func LoadProject(projectRoot string) (*File, error) {
	if strings.TrimSpace(projectRoot) == "" {
		return nil, fmt.Errorf("project root vazio")
	}
	path := ProjectPath(projectRoot)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &File{Version: FileVersion, Presets: nil}, nil
		}
		return nil, fmt.Errorf("ler %s: %w", path, err)
	}
	var f File
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if f.Version == 0 {
		f.Version = FileVersion
	}
	f.Presets = normalizePresets(f.Presets)
	return &f, nil
}

// SaveProject writes presets for a project (creates .openbench/ if needed).
func SaveProject(projectRoot string, f *File) error {
	if strings.TrimSpace(projectRoot) == "" {
		return fmt.Errorf("project root vazio")
	}
	if f == nil {
		f = &File{}
	}
	if f.Version == 0 {
		f.Version = FileVersion
	}
	f.Presets = normalizePresets(f.Presets)
	path := ProjectPath(projectRoot)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(f)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// FindPreset returns a preset by id from the project file.
func FindPreset(projectRoot, id string) (Preset, error) {
	f, err := LoadProject(projectRoot)
	if err != nil {
		return Preset{}, err
	}
	id = strings.TrimSpace(id)
	for _, p := range f.Presets {
		if p.ID == id {
			return p, nil
		}
	}
	return Preset{}, fmt.Errorf("preset %q não encontrado — importe um kit ou adicione em %s", id, RelativePath)
}

// ImportKit merges a built-in kit into the project file.
// Existing preset ids are kept (not overwritten). Returns how many were added.
func ImportKit(projectRoot, kitID string) (added int, err error) {
	kit, err := LoadKit(kitID)
	if err != nil {
		return 0, err
	}
	f, err := LoadProject(projectRoot)
	if err != nil {
		return 0, err
	}
	have := make(map[string]struct{}, len(f.Presets))
	for _, p := range f.Presets {
		have[p.ID] = struct{}{}
	}
	for _, p := range kit.Presets {
		p = normalizePreset(p)
		if p.ID == "" {
			continue
		}
		if _, ok := have[p.ID]; ok {
			continue
		}
		if p.Kit == "" {
			p.Kit = kit.ID
		}
		f.Presets = append(f.Presets, p)
		have[p.ID] = struct{}{}
		added++
	}
	if err := SaveProject(projectRoot, f); err != nil {
		return 0, err
	}
	return added, nil
}

// ParseCommand splits a preset command string into argv.
func ParseCommand(command string) []string {
	return strings.Fields(strings.TrimSpace(command))
}

func normalizePresets(in []Preset) []Preset {
	out := make([]Preset, 0, len(in))
	seen := make(map[string]struct{}, len(in))
	for _, p := range in {
		p = normalizePreset(p)
		if p.ID == "" || p.Command == "" {
			continue
		}
		if _, ok := seen[p.ID]; ok {
			continue
		}
		seen[p.ID] = struct{}{}
		out = append(out, p)
	}
	return out
}

func normalizePreset(p Preset) Preset {
	p.ID = strings.TrimSpace(p.ID)
	p.Label = strings.TrimSpace(p.Label)
	p.Description = strings.TrimSpace(p.Description)
	p.Command = strings.TrimSpace(p.Command)
	p.Kit = strings.TrimSpace(p.Kit)
	if p.Label == "" {
		p.Label = p.ID
	}
	return p
}
