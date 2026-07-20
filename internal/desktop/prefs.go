package desktop

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	prefsFileName = "desktop.yaml"
	MaxPinned     = 8
	MaxRecent     = 5
)

// PinnedProject is a project tab entry.
type PinnedProject struct {
	Path  string `yaml:"path" json:"path"`
	Alias string `yaml:"alias,omitempty" json:"alias,omitempty"`
}

// Prefs stores desktop UI preferences (separate from AI config.yaml).
type Prefs struct {
	ValidateCommit bool            `yaml:"validate_commit" json:"validateCommit"`
	ValidatePR     bool            `yaml:"validate_pr" json:"validatePR"`
	LastProject    string          `yaml:"last_project,omitempty" json:"lastProject,omitempty"`
	Pinned         []PinnedProject `yaml:"pinned,omitempty" json:"pinned,omitempty"`
	Recent         []string        `yaml:"recent,omitempty" json:"recent,omitempty"`
}

// PrefsPath returns ~/.config/openbench/desktop.yaml.
func PrefsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "openbench", prefsFileName), nil
}

// LoadPrefs reads desktop prefs; missing file yields defaults.
func LoadPrefs() (Prefs, error) {
	path, err := PrefsPath()
	if err != nil {
		return Prefs{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Prefs{}, nil
		}
		return Prefs{}, err
	}
	var p Prefs
	if err := yaml.Unmarshal(data, &p); err != nil {
		return Prefs{}, err
	}
	return p, nil
}

// SavePrefs writes desktop prefs atomically enough for v1.
func SavePrefs(p Prefs) error {
	path, err := PrefsPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(p)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// PinProject adds path to pinned (max MaxPinned), most-recent first.
func PinProject(path, alias string) (Prefs, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return Prefs{}, err
	}
	prefs, err := LoadPrefs()
	if err != nil {
		return Prefs{}, err
	}

	existingAlias := alias
	found := false
	kept := make([]PinnedProject, 0, len(prefs.Pinned))
	for _, p := range prefs.Pinned {
		if samePath(p.Path, abs) {
			found = true
			if alias == "" {
				existingAlias = p.Alias
			}
			continue
		}
		kept = append(kept, p)
	}
	if !found && len(prefs.Pinned) >= MaxPinned {
		return prefs, ErrTooManyPinned
	}

	prefs.Pinned = append([]PinnedProject{{Path: abs, Alias: existingAlias}}, kept...)
	prefs.LastProject = abs
	prefs.Recent = pushRecent(prefs.Recent, abs)
	return prefs, SavePrefs(prefs)
}

// UnpinProject removes path from pinned.
func UnpinProject(path string) (Prefs, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return Prefs{}, err
	}
	prefs, err := LoadPrefs()
	if err != nil {
		return Prefs{}, err
	}
	out := prefs.Pinned[:0]
	for _, p := range prefs.Pinned {
		if !samePath(p.Path, abs) {
			out = append(out, p)
		}
	}
	prefs.Pinned = out
	if samePath(prefs.LastProject, abs) {
		prefs.LastProject = ""
		if len(prefs.Pinned) > 0 {
			prefs.LastProject = prefs.Pinned[0].Path
		}
	}
	return prefs, SavePrefs(prefs)
}

// RememberProject updates last + recent and bumps a pinned project to the front.
func RememberProject(path string) (Prefs, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return Prefs{}, err
	}
	prefs, err := LoadPrefs()
	if err != nil {
		return Prefs{}, err
	}
	prefs.LastProject = abs
	prefs.Recent = pushRecent(prefs.Recent, abs)
	prefs.Pinned = bumpPinnedFront(prefs.Pinned, abs)
	return prefs, SavePrefs(prefs)
}

// bumpPinnedFront moves path to index 0 when it is already pinned.
func bumpPinnedFront(pinned []PinnedProject, path string) []PinnedProject {
	var hit *PinnedProject
	kept := make([]PinnedProject, 0, len(pinned))
	for i := range pinned {
		if samePath(pinned[i].Path, path) {
			cp := pinned[i]
			hit = &cp
			continue
		}
		kept = append(kept, pinned[i])
	}
	if hit == nil {
		return pinned
	}
	return append([]PinnedProject{*hit}, kept...)
}

func pushRecent(recent []string, path string) []string {
	out := make([]string, 0, MaxRecent)
	out = append(out, path)
	for _, r := range recent {
		if samePath(r, path) {
			continue
		}
		out = append(out, r)
		if len(out) >= MaxRecent {
			break
		}
	}
	return out
}

func samePath(a, b string) bool {
	aa, errA := filepath.Abs(a)
	bb, errB := filepath.Abs(b)
	if errA != nil || errB != nil {
		return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
	}
	return aa == bb
}

// SamePath reports whether two filesystem paths refer to the same location.
func SamePath(a, b string) bool {
	return samePath(a, b)
}

// ErrTooManyPinned is returned when MaxPinned is exceeded.
var ErrTooManyPinned = errTooManyPinned{}

type errTooManyPinned struct{}

func (errTooManyPinned) Error() string {
	return "limite de projetos pinned atingido (máx. 8)"
}
