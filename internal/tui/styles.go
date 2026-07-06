package tui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/laerciocrestani/gitai/internal/uiprefs"
)

var (
	styleTitle     lipgloss.Style
	styleHeader    lipgloss.Style
	styleSection   lipgloss.Style
	styleCurrent   lipgloss.Style
	styleHint      lipgloss.Style
	styleStatusBar lipgloss.Style
	styleError     lipgloss.Style
	styleKey       lipgloss.Style
	styleModified  lipgloss.Style
	styleNew       lipgloss.Style
	styleUntracked lipgloss.Style
	styleYellow    lipgloss.Style
	styleWarn      lipgloss.Style
)

func init() {
	initTheme()
}

func initTheme() {
	if themePlain() {
		plain := lipgloss.NewStyle()
		bold := lipgloss.NewStyle().Bold(true)
		styleTitle = bold
		styleHeader = plain
		styleSection = bold
		styleCurrent = bold
		styleHint = plain
		styleStatusBar = lipgloss.NewStyle().Padding(0, 1)
		styleError = bold
		styleKey = bold
		styleModified = plain
		styleNew = plain
		styleUntracked = plain
		styleYellow = plain
		styleWarn = plain
		return
	}

	styleTitle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86"))
	styleHeader = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	styleSection = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("86")).MarginTop(1)
	styleCurrent = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	styleHint = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	styleStatusBar = lipgloss.NewStyle().
		Foreground(lipgloss.Color("245")).
		Background(lipgloss.Color("236")).
		Padding(0, 1)
	styleError = lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true)
	styleKey = lipgloss.NewStyle().Foreground(lipgloss.Color("213")).Bold(true)
	styleModified = lipgloss.NewStyle().Foreground(lipgloss.Color("213"))
	styleNew = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	styleUntracked = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	styleYellow = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	styleWarn = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
}

func themePlain() bool {
	return !uiprefs.ColorsEnabled()
}
