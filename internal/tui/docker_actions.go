package tui

import (
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/laerciocrestani/openbench/internal/app"
	dockerpkg "github.com/laerciocrestani/openbench/internal/docker"
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

type appRunError string

func (e appRunError) Error() string { return string(e) }
