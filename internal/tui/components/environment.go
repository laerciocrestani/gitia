package components

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/laerciocrestani/openbench/internal/app"
	dockerpkg "github.com/laerciocrestani/openbench/internal/docker"
	"github.com/laerciocrestani/openbench/internal/tui/theme"
)

// RenderEnvironmentPanel renders Docker/Compose status at the top of the dashboard.
func RenderEnvironmentPanel(snap *app.WorkspaceSnapshot, width int) string {
	if snap == nil || snap.Docker == nil {
		return ""
	}
	ov := snap.Docker
	var lines []string

	daemon := "missing"
	if ov.Available {
		if ov.DaemonRunning {
			daemon = theme.S.Success.Render("running")
		} else {
			daemon = theme.S.Error.Render("stopped")
		}
	} else {
		daemon = theme.S.Disabled.Render("n/a")
	}
	lines = append(lines, fmt.Sprintf("Docker  %s", daemon))

	if ov.ComposeFile != "" {
		composeName := filepath.Base(ov.ComposeFile)
		lines = append(lines, fmt.Sprintf("Compose  %s  (%s)", composeName, ov.ProjectName))
		if ov.Memory.LimitBytes > 0 {
			memLine := fmt.Sprintf("Memory   %s / %s", dockerpkg.FormatBytes(ov.Memory.UsedBytes), dockerpkg.FormatBytes(ov.Memory.LimitBytes))
			if pct := ov.Memory.Percent(); pct > 0 {
				memLine += fmt.Sprintf("  (%.0f%%)", pct)
			}
			lines = append(lines, memLine)
		}
	} else if ov.Available && ov.DaemonRunning {
		lines = append(lines, theme.S.Hint.Render("Compose  not found"))
	}

	if ov.Error != "" && len(ov.Containers) == 0 {
		lines = append(lines, theme.S.Warn.Render(ov.Error))
	}

	if len(ov.Containers) == 0 {
		note := app.FormatDockerNote(ov)
		if note != "" {
			lines = append(lines, theme.S.Hint.Render(note))
		}
	} else {
		maxRows := 6
		for i, c := range ov.Containers {
			if i >= maxRows {
				lines = append(lines, theme.S.Hint.Render(fmt.Sprintf("… +%d more", len(ov.Containers)-maxRows)))
				break
			}
			lines = append(lines, formatContainerLine(c))
		}
	}

	body := strings.Join(lines, "\n")
	if len(ov.Containers) > 0 && ov.ComposeFile != "" {
		body += "\n" + theme.S.Hint.Render("  i — service details")
	}
	return RenderPanel("Environment", body, width)
}

func formatContainerLine(c dockerpkg.ContainerSummary) string {
	stateStyle := theme.S.Hint
	switch strings.ToLower(c.State) {
	case "running":
		stateStyle = theme.S.Success
	case "exited", "dead":
		stateStyle = theme.S.Error
	}
	line := fmt.Sprintf("%-10s %s", c.Service, stateStyle.Render(c.State))
	if c.Ports != "" {
		line += "  " + theme.S.Info.Render(c.Ports)
	}
	if c.Health != "" {
		line += "  " + theme.S.Hint.Render("("+c.Health+")")
	}
	if c.MemUsageBytes > 0 {
		mem := dockerpkg.FormatBytes(c.MemUsageBytes)
		if c.MemPercent != "" {
			mem += " (" + c.MemPercent + ")"
		}
		line += "  " + theme.S.Info.Render(mem)
	}
	return line
}

// RenderServiceListLine renders one selectable service row for the environment screen.
func RenderServiceListLine(c dockerpkg.ContainerSummary, selected bool) string {
	line := formatContainerLine(c)
	if selected {
		return theme.S.Current.Render("▸ " + line)
	}
	return "  " + line
}

// RenderServiceDetail renders detail text for the selected service.
func RenderServiceDetail(c dockerpkg.ContainerSummary, width int) string {
	var lines []string
	lines = append(lines, theme.S.Title.Render("Service detail"))
	lines = append(lines, fmt.Sprintf("  name    %s", c.Service))
	if c.Name != "" {
		lines = append(lines, fmt.Sprintf("  container %s", c.Name))
	}
	lines = append(lines, fmt.Sprintf("  state   %s", c.State))
	if c.Ports != "" {
		lines = append(lines, fmt.Sprintf("  ports   %s", c.Ports))
	}
	if c.Health != "" {
		lines = append(lines, fmt.Sprintf("  health  %s", c.Health))
	}
	if c.MemUsageBytes > 0 {
		mem := dockerpkg.FormatBytes(c.MemUsageBytes)
		if c.MemLimitBytes > 0 {
			mem += " / " + dockerpkg.FormatBytes(c.MemLimitBytes)
		}
		if c.MemPercent != "" {
			mem += " (" + c.MemPercent + ")"
		}
		lines = append(lines, fmt.Sprintf("  memory  %s", mem))
	}
	body := strings.Join(lines, "\n")
	if width > 0 {
		return RenderPanel("", body, width)
	}
	return body
}
