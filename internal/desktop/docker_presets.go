package desktop

import (
	"fmt"
	"strings"

	"github.com/laerciocrestani/openbench/internal/app"
	dockerpkg "github.com/laerciocrestani/openbench/internal/docker"
	"github.com/laerciocrestani/openbench/internal/dockerpresets"
)

// DockerPresetView is a project preset for the UI.
type DockerPresetView struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Command     string `json:"command"`
	Kit         string `json:"kit,omitempty"`
	Interactive bool   `json:"interactive"`
}

// DockerKitView is a built-in kit summary.
type DockerKitView struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description"`
	PresetCount int    `json:"presetCount"`
}

// DockerExecResult is the summary shown after a one-shot preset/exec.
// When Interactive is true, the UI should open a docker shell with Args instead of a modal.
type DockerExecResult struct {
	PresetID    string   `json:"presetId,omitempty"`
	Label       string   `json:"label,omitempty"`
	Service     string   `json:"service"`
	Command     string   `json:"command"`
	Args        []string `json:"args"`
	Output      string   `json:"output"`
	ExitCode    int      `json:"exitCode"`
	OK          bool     `json:"ok"`
	Summary     string   `json:"summary"`
	Interactive bool     `json:"interactive"`
}

// DockerImportResult reports kit import outcome.
type DockerImportResult struct {
	KitID   string             `json:"kitId"`
	Added   int                `json:"added"`
	Message string             `json:"message"`
	Presets []DockerPresetView `json:"presets"`
}

// ListDockerPresets returns presets configured for the project.
func ListDockerPresets(projectPath string) ([]DockerPresetView, error) {
	if strings.TrimSpace(projectPath) == "" {
		return nil, fmt.Errorf("no project open")
	}
	f, err := dockerpresets.LoadProject(projectPath)
	if err != nil {
		return nil, err
	}
	out := make([]DockerPresetView, 0, len(f.Presets))
	for _, p := range f.Presets {
		out = append(out, mapPreset(p))
	}
	return out, nil
}

// ListDockerKits returns built-in importable kits.
func ListDockerKits() ([]DockerKitView, error) {
	kits, err := dockerpresets.ListKits()
	if err != nil {
		return nil, err
	}
	out := make([]DockerKitView, 0, len(kits))
	for _, k := range kits {
		out = append(out, DockerKitView{
			ID:          k.ID,
			Label:       k.Label,
			Description: k.Description,
			PresetCount: k.PresetCount,
		})
	}
	return out, nil
}

// ImportDockerKit merges a kit into the project presets file.
func ImportDockerKit(projectPath, kitID string) (*DockerImportResult, error) {
	if strings.TrimSpace(projectPath) == "" {
		return nil, fmt.Errorf("no project open")
	}
	added, err := dockerpresets.ImportKit(projectPath, kitID)
	if err != nil {
		return nil, err
	}
	presets, err := ListDockerPresets(projectPath)
	if err != nil {
		return nil, err
	}
	msg := fmt.Sprintf("Kit %q: %d preset(s) adicionados", kitID, added)
	if added == 0 {
		msg = fmt.Sprintf("Kit %q: nenhum preset novo (já importados)", kitID)
	}
	return &DockerImportResult{
		KitID:   strings.TrimSpace(kitID),
		Added:   added,
		Message: msg,
		Presets: presets,
	}, nil
}

// RunDockerPreset executes a project preset on the chosen service.
// Interactive presets return Interactive=true so the UI opens a docker shell with Args.
func RunDockerPreset(projectPath, service, presetID string) (*DockerExecResult, error) {
	if strings.TrimSpace(projectPath) == "" {
		return nil, fmt.Errorf("no project open")
	}
	service = strings.TrimSpace(service)
	if service == "" {
		return nil, fmt.Errorf("selecione um container/serviço")
	}
	preset, err := dockerpresets.FindPreset(projectPath, presetID)
	if err != nil {
		return nil, err
	}
	args := dockerpresets.ParseCommand(preset.Command)
	if len(args) == 0 {
		return nil, fmt.Errorf("preset %q sem comando", preset.ID)
	}
	if preset.Interactive {
		return &DockerExecResult{
			PresetID:    preset.ID,
			Label:       preset.Label,
			Service:     service,
			Command:     preset.Command,
			Args:        args,
			OK:          true,
			Summary:     fmt.Sprintf("%s — abrir shell em %s", preset.Label, service),
			Interactive: true,
		}, nil
	}

	compose, err := resolveProjectCompose(projectPath)
	if err != nil {
		return nil, err
	}
	raw, err := dockerpkg.ExecOutput(dockerpkg.ExecOptions{
		ComposeFile: compose,
		Service:     service,
		Command:     args,
	})
	if err != nil {
		return nil, err
	}
	ok := raw.ExitCode == 0
	summary := fmt.Sprintf("%s em %s — exit %d", preset.Label, service, raw.ExitCode)
	if ok {
		summary = fmt.Sprintf("%s em %s — ok", preset.Label, service)
	}
	return &DockerExecResult{
		PresetID: preset.ID,
		Label:    preset.Label,
		Service:  service,
		Command:  preset.Command,
		Args:     args,
		Output:   raw.Output,
		ExitCode: raw.ExitCode,
		OK:       ok,
		Summary:  summary,
	}, nil
}

// RunDockerExecCommand runs an arbitrary one-shot command on a service.
func RunDockerExecCommand(projectPath, service, command string) (*DockerExecResult, error) {
	if strings.TrimSpace(projectPath) == "" {
		return nil, fmt.Errorf("no project open")
	}
	service = strings.TrimSpace(service)
	if service == "" {
		return nil, fmt.Errorf("selecione um container/serviço")
	}
	args := app.ParseExecCommand(command)
	if len(args) == 0 {
		return nil, fmt.Errorf("informe um comando")
	}
	compose, err := resolveProjectCompose(projectPath)
	if err != nil {
		return nil, err
	}
	raw, err := dockerpkg.ExecOutput(dockerpkg.ExecOptions{
		ComposeFile: compose,
		Service:     service,
		Command:     args,
	})
	if err != nil {
		return nil, err
	}
	ok := raw.ExitCode == 0
	summary := fmt.Sprintf("exec em %s — exit %d", service, raw.ExitCode)
	if ok {
		summary = fmt.Sprintf("exec em %s — ok", service)
	}
	return &DockerExecResult{
		Service:  service,
		Command:  strings.TrimSpace(command),
		Args:     args,
		Output:   raw.Output,
		ExitCode: raw.ExitCode,
		OK:       ok,
		Summary:  summary,
	}, nil
}

// ResolveDockerShellCommand builds argv for an interactive docker shell or interactive preset.
func ResolveDockerShellCommand(projectPath, service, presetID string) (composeFile string, argv []string, err error) {
	if strings.TrimSpace(projectPath) == "" {
		return "", nil, fmt.Errorf("no project open")
	}
	service = strings.TrimSpace(service)
	if service == "" {
		return "", nil, fmt.Errorf("selecione um container/serviço")
	}
	compose, err := resolveProjectCompose(projectPath)
	if err != nil {
		return "", nil, err
	}
	if strings.TrimSpace(presetID) == "" {
		return compose, []string{"sh"}, nil
	}
	preset, err := dockerpresets.FindPreset(projectPath, presetID)
	if err != nil {
		return "", nil, err
	}
	args := dockerpresets.ParseCommand(preset.Command)
	if len(args) == 0 {
		return "", nil, fmt.Errorf("preset %q sem comando", preset.ID)
	}
	return compose, args, nil
}

func resolveProjectCompose(projectPath string) (string, error) {
	ov := dockerpkg.LoadOverview(projectPath)
	if !ov.Available {
		return "", fmt.Errorf("docker CLI não encontrado")
	}
	if !ov.DaemonRunning {
		return "", fmt.Errorf("Docker daemon não está rodando")
	}
	if ov.ComposeFile == "" {
		return "", fmt.Errorf("compose file não encontrado no projeto")
	}
	return ov.ComposeFile, nil
}

func mapPreset(p dockerpresets.Preset) DockerPresetView {
	return DockerPresetView{
		ID:          p.ID,
		Label:       p.Label,
		Description: p.Description,
		Command:     p.Command,
		Kit:         p.Kit,
		Interactive: p.Interactive,
	}
}
