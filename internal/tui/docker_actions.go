package tui

import (
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/laerciocrestani/openbench/internal/app"
	dockerpkg "github.com/laerciocrestani/openbench/internal/docker"
	"github.com/laerciocrestani/openbench/internal/dockerpresets"
)

func runDockerUpCmd() tea.Cmd {
	return func() tea.Msg {
		return dockerActionMsg{action: "up", err: app.RunDockerUp(app.DockerOptions{})}
	}
}

func runDockerDownCmd() tea.Cmd {
	return func() tea.Msg {
		return dockerActionMsg{action: "down", err: app.RunDockerDown(app.DockerOptions{})}
	}
}

func runDockerServiceUpCmd(service string) tea.Cmd {
	return func() tea.Msg {
		err := app.RunDockerUp(app.DockerOptions{Services: []string{service}})
		return dockerActionMsg{action: "up:" + service, err: err}
	}
}

func runDockerServiceStopCmd(service string) tea.Cmd {
	return func() tea.Msg {
		err := app.RunDockerStop(app.DockerOptions{Services: []string{service}})
		return dockerActionMsg{action: "stop:" + service, err: err}
	}
}

func runDockerServiceRecreateCmd(service string) tea.Cmd {
	return func() tea.Msg {
		err := app.RunDockerRecreate(app.DockerOptions{Service: service})
		return dockerActionMsg{action: "recreate:" + service, err: err}
	}
}

func runDockerShellExecCmd(snap *app.WorkspaceSnapshot, service string) tea.Cmd {
	compose := app.DockerComposeFile(snap)
	if compose == "" {
		return func() tea.Msg {
			return dockerActionMsg{action: "shell:" + service, err: errComposeMissing()}
		}
	}
	cmd, err := dockerpkg.BuildShellCommand(compose, service)
	if err != nil {
		return func() tea.Msg {
			return dockerActionMsg{action: "shell:" + service, err: err}
		}
	}
	return tea.ExecProcess(cmd, func(runErr error) tea.Msg {
		if runErr != nil {
			if bashCmd, bashErr := dockerpkg.BuildExecCommand(compose, service, true, "bash"); bashErr == nil {
				return execFallbackMsg{primary: runErr, fallback: bashCmd, action: "shell:" + service}
			}
		}
		return dockerActionMsg{action: "shell:" + service, err: runErr}
	})
}

type execFallbackMsg struct {
	primary  error
	fallback *exec.Cmd
	action   string
}

func runDockerExecProcessCmd(snap *app.WorkspaceSnapshot, service string, command []string, interactive bool) tea.Cmd {
	compose := app.DockerComposeFile(snap)
	if compose == "" {
		return func() tea.Msg {
			return dockerActionMsg{action: "exec:" + service, err: errComposeMissing()}
		}
	}
	cmd, err := dockerpkg.BuildExecCommand(compose, service, interactive, command...)
	if err != nil {
		return func() tea.Msg {
			return dockerActionMsg{action: "exec:" + service, err: err}
		}
	}
	return tea.ExecProcess(cmd, func(runErr error) tea.Msg {
		return dockerActionMsg{action: "exec:" + service, err: runErr}
	})
}

func runDockerShellCmd(snap *app.WorkspaceSnapshot) tea.Cmd {
	service := app.DockerDefaultService(snap)
	return runDockerShellExecCmd(snap, service)
}

func errComposeMissing() error {
	return appRunError("compose file não encontrado")
}

type dockerPresetResultMsg struct {
	summary string
	output  string
	err     error
}

func runDockerPresetExecCmd(snap *app.WorkspaceSnapshot, service, presetID string) tea.Cmd {
	return func() tea.Msg {
		cwd, err := os.Getwd()
		if err != nil {
			return dockerPresetResultMsg{err: err}
		}
		preset, err := dockerpresets.FindPreset(cwd, presetID)
		if err != nil {
			return dockerPresetResultMsg{err: err}
		}
		args := dockerpresets.ParseCommand(preset.Command)
		if len(args) == 0 {
			return dockerPresetResultMsg{err: fmt.Errorf("preset sem comando")}
		}
		if preset.Interactive {
			return dockerPresetResultMsg{
				err: fmt.Errorf("preset interativo — use E (shell) e rode: %s", preset.Command),
			}
		}
		compose := app.DockerComposeFile(snap)
		if compose == "" {
			return dockerPresetResultMsg{err: errComposeMissing()}
		}
		res, err := dockerpkg.ExecOutput(dockerpkg.ExecOptions{
			ComposeFile: compose,
			Service:     service,
			Command:     args,
		})
		if err != nil {
			return dockerPresetResultMsg{err: err}
		}
		summary := fmt.Sprintf("%s @ %s — exit %d", preset.Label, service, res.ExitCode)
		if res.ExitCode == 0 {
			summary = fmt.Sprintf("%s @ %s — ok", preset.Label, service)
		}
		return dockerPresetResultMsg{summary: summary, output: res.Output}
	}
}

func runDockerKitImportCmd(kitID string) tea.Cmd {
	return func() tea.Msg {
		cwd, err := os.Getwd()
		if err != nil {
			return dockerActionMsg{action: "kit-import", err: err}
		}
		added, err := dockerpresets.ImportKit(cwd, kitID)
		if err != nil {
			return dockerActionMsg{action: "kit-import", err: err}
		}
		msg := fmt.Sprintf("kit %s: %d adicionados", kitID, added)
		if added == 0 {
			msg = fmt.Sprintf("kit %s: já importado", kitID)
		}
		return dockerActionMsg{action: msg, err: nil}
	}
}

type appRunError string

func (e appRunError) Error() string { return string(e) }
