package dockerpresets

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed kits/*.yaml
var kitsFS embed.FS

// Kit is a packaged preset catalog that can be imported into a project.
type Kit struct {
	ID          string   `yaml:"id" json:"id"`
	Label       string   `yaml:"label" json:"label"`
	Description string   `yaml:"description,omitempty" json:"description,omitempty"`
	Presets     []Preset `yaml:"presets" json:"presets"`
}

// KitInfo is a lightweight kit summary for UI/CLI lists.
type KitInfo struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
	PresetCount int    `json:"presetCount"`
}

// ListKits returns built-in kits sorted by id.
func ListKits() ([]KitInfo, error) {
	entries, err := fs.ReadDir(kitsFS, "kits")
	if err != nil {
		return nil, err
	}
	var out []KitInfo
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}
		kit, err := loadKitFile(path.Join("kits", e.Name()))
		if err != nil {
			return nil, err
		}
		out = append(out, KitInfo{
			ID:          kit.ID,
			Label:       kit.Label,
			Description: kit.Description,
			PresetCount: len(kit.Presets),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// LoadKit loads a built-in kit by id (filename stem or yaml id).
func LoadKit(kitID string) (*Kit, error) {
	kitID = strings.TrimSpace(kitID)
	if kitID == "" {
		return nil, fmt.Errorf("informe o id do kit (ex: laravel)")
	}
	// Prefer filename match.
	name := kitID
	if !strings.HasSuffix(name, ".yaml") {
		name += ".yaml"
	}
	kit, err := loadKitFile(path.Join("kits", name))
	if err == nil {
		return kit, nil
	}
	// Fallback: scan by yaml id field.
	infos, listErr := ListKits()
	if listErr != nil {
		return nil, err
	}
	for _, info := range infos {
		if info.ID == kitID {
			return loadKitFile(path.Join("kits", info.ID+".yaml"))
		}
	}
	return nil, fmt.Errorf("kit %q não encontrado — use: ob docker kit list", kitID)
}

func loadKitFile(rel string) (*Kit, error) {
	data, err := kitsFS.ReadFile(rel)
	if err != nil {
		return nil, fmt.Errorf("kit %s: %w", rel, err)
	}
	var kit Kit
	if err := yaml.Unmarshal(data, &kit); err != nil {
		return nil, fmt.Errorf("parse kit %s: %w", rel, err)
	}
	kit.ID = strings.TrimSpace(kit.ID)
	kit.Label = strings.TrimSpace(kit.Label)
	kit.Description = strings.TrimSpace(kit.Description)
	if kit.ID == "" {
		base := path.Base(rel)
		kit.ID = strings.TrimSuffix(base, path.Ext(base))
	}
	if kit.Label == "" {
		kit.Label = kit.ID
	}
	kit.Presets = normalizePresets(kit.Presets)
	for i := range kit.Presets {
		if kit.Presets[i].Kit == "" {
			kit.Presets[i].Kit = kit.ID
		}
	}
	return &kit, nil
}
