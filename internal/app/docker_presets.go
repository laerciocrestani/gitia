package app

import (
	"fmt"
	"os"
	"strings"

	dockerpkg "github.com/laerciocrestani/openbench/internal/docker"
	"github.com/laerciocrestani/openbench/internal/dockerpresets"
	"github.com/laerciocrestani/openbench/internal/ui"
)

// RunDockerPresetList prints project presets.
func RunDockerPresetList(workDir string) error {
	root, err := resolveWorkDir(workDir)
	if err != nil {
		return err
	}
	sess := ui.New("docker preset list", false)
	sess.Header()
	f, err := dockerpresets.LoadProject(root)
	if err != nil {
		return err
	}
	if len(f.Presets) == 0 {
		sess.Info(fmt.Sprintf("Nenhum preset em %s", dockerpresets.RelativePath))
		sess.Info("Importe um kit: ob docker kit import laravel")
		return nil
	}
	for _, p := range f.Presets {
		extra := ""
		if p.Interactive {
			extra = " [interactive]"
		}
		if p.Kit != "" {
			extra += " (" + p.Kit + ")"
		}
		sess.MetaRow(p.ID, fmt.Sprintf("%s — %s%s", p.Label, p.Command, extra))
		if p.Description != "" {
			sess.Detail("  " + p.Description)
		}
	}
	return nil
}

// RunDockerPresetRun executes a preset against a service (one-shot; prints output).
func RunDockerPresetRun(workDir, service, presetID string, dryRun bool) error {
	root, err := resolveWorkDir(workDir)
	if err != nil {
		return err
	}
	service = strings.TrimSpace(service)
	presetID = strings.TrimSpace(presetID)
	if presetID == "" {
		return fmt.Errorf("informe o preset: ob docker preset run <id> --service <svc>")
	}
	preset, err := dockerpresets.FindPreset(root, presetID)
	if err != nil {
		return err
	}
	args := dockerpresets.ParseCommand(preset.Command)
	if len(args) == 0 {
		return fmt.Errorf("preset %q sem comando", preset.ID)
	}
	if service == "" {
		ov := dockerpkg.LoadOverview(root)
		service = ov.DefaultService()
	}
	if service == "" {
		return fmt.Errorf("informe o serviço: ob docker preset run %s --service <svc>", presetID)
	}
	if preset.Interactive {
		return fmt.Errorf("preset %q é interativo — use: ob docker sh %s  (ou exec com %s)", preset.ID, service, preset.Command)
	}
	compose := dockerpkg.FindComposeFile(root)
	if compose == "" {
		return fmt.Errorf("compose file não encontrado no projeto")
	}
	sess := ui.New("docker preset run", false)
	sess.Header()
	sess.MetaRow("Preset", preset.ID)
	sess.MetaRow("Service", service)
	sess.MetaRow("Command", preset.Command)
	if dryRun {
		fmt.Printf("[dry-run] docker compose exec -T %s %s\n", service, strings.Join(args, " "))
		return nil
	}
	res, err := dockerpkg.ExecOutput(dockerpkg.ExecOptions{
		ComposeFile: compose,
		Service:     service,
		Command:     args,
	})
	if err != nil {
		return err
	}
	fmt.Print(res.Output)
	if res.ExitCode != 0 {
		return fmt.Errorf("comando terminou com exit %d", res.ExitCode)
	}
	sess.Success(fmt.Sprintf("%s ok", preset.Label))
	return nil
}

// RunDockerKitList prints built-in kits.
func RunDockerKitList() error {
	sess := ui.New("docker kit list", false)
	sess.Header()
	kits, err := dockerpresets.ListKits()
	if err != nil {
		return err
	}
	for _, k := range kits {
		sess.MetaRow(k.ID, fmt.Sprintf("%s — %d presets", k.Label, k.PresetCount))
		if k.Description != "" {
			sess.Detail("  " + k.Description)
		}
	}
	return nil
}

// RunDockerKitImport merges a kit into the project presets file.
func RunDockerKitImport(workDir, kitID string) error {
	root, err := resolveWorkDir(workDir)
	if err != nil {
		return err
	}
	sess := ui.New("docker kit import", false)
	sess.Header()
	added, err := dockerpresets.ImportKit(root, kitID)
	if err != nil {
		return err
	}
	sess.MetaRow("Arquivo", dockerpresets.ProjectPath(root))
	if added == 0 {
		sess.Info(fmt.Sprintf("Kit %q: nenhum preset novo (já presentes)", kitID))
		return nil
	}
	sess.Success(fmt.Sprintf("Kit %q: %d preset(s) adicionados", kitID, added))
	return nil
}

func resolveWorkDir(workDir string) (string, error) {
	if strings.TrimSpace(workDir) != "" {
		return workDir, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return cwd, nil
}
