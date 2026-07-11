package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/laerciocrestani/openbench/internal/app"
	dockerpkg "github.com/laerciocrestani/openbench/internal/docker"
	"github.com/laerciocrestani/openbench/internal/tui/components"
)

type environmentMode int

const (
	environmentModeList environmentMode = iota
	environmentModeExec
)

type environmentModel struct {
	snap         *app.WorkspaceSnapshot
	containers   []dockerpkg.ContainerSummary
	cursor       int
	listViewport viewport.Model
	ready        bool
	mode         environmentMode
	execInput    textinput.Model
	execReady    bool
}

func newEnvironmentModel() environmentModel {
	ti := textinput.New()
	ti.Placeholder = "php artisan migrate"
	ti.CharLimit = 512
	return environmentModel{execInput: ti}
}

func (m *environmentModel) Load(snap *app.WorkspaceSnapshot) {
	m.snap = snap
	m.containers = nil
	m.cursor = 0
	m.mode = environmentModeList
	m.execInput.SetValue("")
	m.execInput.Blur()
	if snap != nil && snap.Docker != nil {
		m.containers = append(m.containers, snap.Docker.Containers...)
	}
	m.refreshListContent()
}

func (m *environmentModel) SetSize(width, height int) {
	listRows := height/2 - 4
	if listRows < 6 {
		listRows = 6
	}
	if len(m.containers) > 0 && listRows > len(m.containers) {
		listRows = len(m.containers)
	}
	if !m.ready {
		m.listViewport = viewport.New(width, listRows)
		m.ready = true
	} else {
		m.listViewport.Width = width
		m.listViewport.Height = listRows
	}
	m.refreshListContent()
}

func (m *environmentModel) refreshListContent() {
	if !m.ready {
		return
	}
	m.listViewport.SetContent(m.listContent())
	m.syncScroll()
}

func (m *environmentModel) listContent() string {
	if len(m.containers) == 0 {
		return ""
	}
	lines := make([]string, len(m.containers))
	for i, c := range m.containers {
		lines[i] = components.RenderServiceListLine(c, i == m.cursor)
	}
	return strings.Join(lines, "\n")
}

func (m *environmentModel) syncScroll() {
	if !m.ready || len(m.containers) == 0 {
		return
	}
	if m.cursor >= m.listViewport.YOffset+m.listViewport.Height {
		m.listViewport.SetYOffset(m.cursor - m.listViewport.Height + 1)
	} else if m.cursor < m.listViewport.YOffset {
		m.listViewport.SetYOffset(m.cursor)
	}
}

func (m *environmentModel) selectedService() string {
	if m.cursor < 0 || m.cursor >= len(m.containers) {
		return ""
	}
	return m.containers[m.cursor].Service
}

func (m *environmentModel) selectedContainer() (dockerpkg.ContainerSummary, bool) {
	if m.cursor < 0 || m.cursor >= len(m.containers) {
		return dockerpkg.ContainerSummary{}, false
	}
	return m.containers[m.cursor], true
}

func (m *environmentModel) moveCursor(delta int) {
	if len(m.containers) == 0 {
		return
	}
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(m.containers) {
		m.cursor = len(m.containers) - 1
	}
	m.refreshListContent()
}

func (m *environmentModel) startExecPrompt() {
	m.mode = environmentModeExec
	m.execInput.SetValue("")
	m.execInput.Focus()
	m.execReady = true
}

func (m *environmentModel) cancelExecPrompt() {
	m.mode = environmentModeList
	m.execInput.Blur()
}

func (m *environmentModel) Update(msg tea.Msg) (environmentModel, tea.Cmd) {
	if m.mode != environmentModeExec {
		return *m, nil
	}
	var cmd tea.Cmd
	m.execInput, cmd = m.execInput.Update(msg)
	return *m, cmd
}

func (m environmentModel) View(width int, tick int) string {
	var b strings.Builder
	b.WriteString(styleSection.Render("Environment · Docker"))
	b.WriteString("\n")

	if m.snap == nil || m.snap.Docker == nil {
		b.WriteString(styleError.Render("  Docker indisponível"))
		return b.String()
	}

	ov := m.snap.Docker
	if ov.ComposeFile != "" {
		b.WriteString(styleHint.Render(fmt.Sprintf("  %s · %s", filepath.Base(ov.ComposeFile), ov.ProjectName)))
		b.WriteString("\n")
		if ov.Memory.LimitBytes > 0 {
			mem := fmt.Sprintf("  RAM %s / %s", dockerpkg.FormatBytes(ov.Memory.UsedBytes), dockerpkg.FormatBytes(ov.Memory.LimitBytes))
			if pct := ov.Memory.Percent(); pct > 0 {
				mem += fmt.Sprintf(" (%.0f%%)", pct)
			}
			b.WriteString(styleHint.Render(mem))
		}
		b.WriteString("\n\n")
	}

	if m.mode == environmentModeExec {
		svc := m.selectedService()
		b.WriteString(styleHint.Render(fmt.Sprintf("  Exec em %s — digite o comando (ex: php artisan migrate)", svc)))
		b.WriteString("\n\n")
		b.WriteString("  ")
		b.WriteString(m.execInput.View())
		b.WriteString("\n")
		return b.String()
	}

	if len(m.containers) == 0 {
		b.WriteString(styleHint.Render("  Nenhum serviço encontrado"))
		return b.String()
	}

	if m.ready {
		b.WriteString(m.listViewport.View())
	} else {
		b.WriteString(m.listContent())
	}
	b.WriteString("\n\n")

	if c, ok := m.selectedContainer(); ok {
		b.WriteString(components.RenderServiceDetail(c, width))
	}

	return b.String()
}

func environmentHelpLine(snap *app.WorkspaceSnapshot, mode environmentMode, service string) string {
	if mode == environmentModeExec {
		return styleKey.Render("enter") + " run  " +
			styleKey.Render("esc") + " cancel"
	}
	parts := []string{
		styleKey.Render("↑↓") + " select",
		styleKey.Render("U") + " up svc",
		styleKey.Render("D") + " stop svc",
		styleKey.Render("R") + " recreate",
		styleKey.Render("E") + " shell",
		styleKey.Render("x") + " exec",
		styleKey.Render("L") + " logs",
	}
	if app.CanDockerProjectUp(snap) {
		parts = append(parts, styleKey.Render("Shift+U")+" up all")
	}
	if app.CanDockerProjectDown(snap) {
		parts = append(parts, styleKey.Render("Shift+D")+" down all")
	}
	parts = append(parts, styleKey.Render("r")+" refresh", styleKey.Render("esc")+" back")
	_ = service
	return strings.Join(parts, "  ")
}
