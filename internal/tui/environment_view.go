package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/laerciocrestani/openbench/internal/app"
	dockerpkg "github.com/laerciocrestani/openbench/internal/docker"
	"github.com/laerciocrestani/openbench/internal/dockerpresets"
	"github.com/laerciocrestani/openbench/internal/tui/components"
)

type environmentMode int

const (
	environmentModeList environmentMode = iota
	environmentModeExec
	environmentModePreset
	environmentModePresetResult
)

type environmentModel struct {
	snap          *app.WorkspaceSnapshot
	containers    []dockerpkg.ContainerSummary
	cursor        int
	listViewport  viewport.Model
	ready         bool
	mode          environmentMode
	execInput     textinput.Model
	execReady     bool
	presets       []dockerpresets.Preset
	presetCursor  int
	presetResult  string
	presetSummary string
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
	m.presetResult = ""
	m.presetSummary = ""
	if snap != nil && snap.Docker != nil {
		m.containers = append(m.containers, snap.Docker.Containers...)
	}
	m.reloadPresets()
	m.refreshListContent()
}

func (m *environmentModel) reloadPresets() {
	m.presets = nil
	m.presetCursor = 0
	cwd, err := os.Getwd()
	if err != nil {
		return
	}
	f, err := dockerpresets.LoadProject(cwd)
	if err != nil || f == nil {
		return
	}
	m.presets = append(m.presets, f.Presets...)
}

func (m *environmentModel) SetSize(width, height int) {
	listRows := height/2 - 4
	if listRows < 6 {
		listRows = 6
	}
	n := len(m.containers)
	if m.mode == environmentModePreset {
		n = len(m.presets)
	}
	if n > 0 && listRows > n {
		listRows = n
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
	if m.mode == environmentModePreset {
		if len(m.presets) == 0 {
			return styleHint.Render("  Nenhum preset — pressione I para importar kit laravel")
		}
		lines := make([]string, len(m.presets))
		for i, p := range m.presets {
			mark := "  "
			if i == m.presetCursor {
				mark = "> "
			}
			extra := ""
			if p.Interactive {
				extra = " · interactive"
			}
			lines[i] = mark + styleKey.Render(p.Label) + styleHint.Render("  "+p.Command+extra)
		}
		return strings.Join(lines, "\n")
	}
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
	if !m.ready {
		return
	}
	cur, n := m.cursor, len(m.containers)
	if m.mode == environmentModePreset {
		cur, n = m.presetCursor, len(m.presets)
	}
	if n == 0 {
		return
	}
	if cur >= m.listViewport.YOffset+m.listViewport.Height {
		m.listViewport.SetYOffset(cur - m.listViewport.Height + 1)
	} else if cur < m.listViewport.YOffset {
		m.listViewport.SetYOffset(cur)
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

func (m *environmentModel) selectedPreset() (dockerpresets.Preset, bool) {
	if m.presetCursor < 0 || m.presetCursor >= len(m.presets) {
		return dockerpresets.Preset{}, false
	}
	return m.presets[m.presetCursor], true
}

func (m *environmentModel) moveCursor(delta int) {
	if m.mode == environmentModePreset {
		if len(m.presets) == 0 {
			return
		}
		m.presetCursor += delta
		if m.presetCursor < 0 {
			m.presetCursor = 0
		}
		if m.presetCursor >= len(m.presets) {
			m.presetCursor = len(m.presets) - 1
		}
		m.refreshListContent()
		return
	}
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

func (m *environmentModel) startPresetMode() {
	m.reloadPresets()
	m.mode = environmentModePreset
	m.presetCursor = 0
	m.refreshListContent()
}

func (m *environmentModel) leavePresetMode() {
	m.mode = environmentModeList
	m.presetResult = ""
	m.presetSummary = ""
	m.refreshListContent()
}

func (m *environmentModel) showPresetResult(summary, output string) {
	m.mode = environmentModePresetResult
	m.presetSummary = summary
	m.presetResult = output
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

	if m.mode == environmentModePresetResult {
		b.WriteString(styleHint.Render("  " + m.presetSummary))
		b.WriteString("\n\n")
		out := m.presetResult
		if strings.TrimSpace(out) == "" {
			out = "(sem output)"
		}
		// Keep result readable in the TUI viewport budget.
		lines := strings.Split(out, "\n")
		if len(lines) > 24 {
			lines = append(lines[:24], "…")
		}
		for _, line := range lines {
			b.WriteString("  ")
			b.WriteString(line)
			b.WriteString("\n")
		}
		b.WriteString("\n")
		b.WriteString(styleHint.Render("  esc fecha o resumo"))
		return b.String()
	}

	if m.mode == environmentModePreset {
		svc := m.selectedService()
		b.WriteString(styleHint.Render(fmt.Sprintf("  Presets · serviço alvo: %s", svc)))
		b.WriteString("\n\n")
		if m.ready {
			b.WriteString(m.listViewport.View())
		} else {
			b.WriteString(m.listContent())
		}
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
	switch mode {
	case environmentModeExec:
		return styleKey.Render("enter") + " run  " +
			styleKey.Render("esc") + " cancel"
	case environmentModePreset:
		return styleKey.Render("↑↓") + " select  " +
			styleKey.Render("enter") + " run  " +
			styleKey.Render("I") + " import laravel  " +
			styleKey.Render("esc") + " back"
	case environmentModePresetResult:
		return styleKey.Render("esc") + " fechar resumo"
	}
	parts := []string{
		styleKey.Render("↑↓") + " select",
		styleKey.Render("U") + " up svc",
		styleKey.Render("D") + " stop svc",
		styleKey.Render("R") + " recreate",
		styleKey.Render("E") + " shell",
		styleKey.Render("x") + " exec",
		styleKey.Render("p") + " presets",
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
