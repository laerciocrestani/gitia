package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/laerciocrestani/openbench/internal/app"
	"github.com/laerciocrestani/openbench/internal/tui/components"
)

type dockerLogsModel struct {
	viewport viewport.Model
	content  string
	ready    bool
	err      error
	service  string
}

func newDockerLogsModel() dockerLogsModel {
	return dockerLogsModel{}
}

func loadDockerLogsCmd(snap *app.WorkspaceSnapshot) tea.Cmd {
	return loadDockerLogsForServiceCmd(snap, app.DockerDefaultService(snap))
}

func loadDockerLogsForServiceCmd(snap *app.WorkspaceSnapshot, service string) tea.Cmd {
	return func() tea.Msg {
		content, err := app.RunDockerLogsOutput(app.DockerOptions{
			Service: service,
			Tail:    200,
		})
		return dockerLogsLoadedMsg{
			content: content,
			err:     err,
			service: service,
		}
	}
}

type dockerLogsLoadedMsg struct {
	content string
	err     error
	service string
}

type dockerActionMsg struct {
	action string
	err    error
}

func (m *dockerLogsModel) SetSize(width, height int) {
	headerRows := 4
	footerRows := 2
	vh := height - headerRows - footerRows
	if vh < 3 {
		vh = 3
	}
	if !m.ready {
		m.viewport = viewport.New(width, vh)
		m.viewport.SetContent(m.content)
		m.ready = true
	} else {
		m.viewport.Width = width
		m.viewport.Height = vh
	}
}

func (m *dockerLogsModel) Load(msg dockerLogsLoadedMsg) {
	m.content = msg.content
	m.err = msg.err
	m.service = msg.service
	m.ready = false
}

func (m dockerLogsModel) Update(msg tea.Msg) (dockerLogsModel, tea.Cmd) {
	if m.err != nil {
		return m, nil
	}
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m dockerLogsModel) View(tick int) string {
	if m.err != nil {
		return styleError.Render("  ✗ " + m.err.Error())
	}

	var b strings.Builder
	title := "Docker Logs"
	if m.service != "" {
		title += " · " + m.service
	}
	b.WriteString(styleSection.Render(title))
	b.WriteString("\n")
	b.WriteString(styleHint.Render("  Últimas linhas do serviço"))
	b.WriteString("\n\n")
	if !m.ready {
		b.WriteString(components.RenderSpinnerLine("Loading logs", tick))
		return b.String()
	}
	b.WriteString(m.viewport.View())
	return b.String()
}

func dockerLogsHelpLine() string {
	return styleKey.Render("r") + " refresh  " +
		styleKey.Render("esc") + " back"
}
