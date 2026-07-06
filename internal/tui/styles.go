package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorCyan    = lipgloss.Color("86")
	colorGreen   = lipgloss.Color("42")
	colorYellow  = lipgloss.Color("214")
	colorMagenta = lipgloss.Color("213")
	colorDim     = lipgloss.Color("245")
	colorRed     = lipgloss.Color("203")

	styleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorCyan)

	styleHeader = lipgloss.NewStyle().
			Foreground(colorDim)

	styleSection = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorCyan).
			MarginTop(1)

	styleCurrent = lipgloss.NewStyle().
			Foreground(colorGreen).
			Bold(true)

	styleHint = lipgloss.NewStyle().
			Foreground(colorDim)

	styleStatusBar = lipgloss.NewStyle().
			Foreground(colorDim).
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	styleError = lipgloss.NewStyle().
			Foreground(colorRed).
			Bold(true)

	styleKey = lipgloss.NewStyle().
			Foreground(colorMagenta).
			Bold(true)

	styleModified  = lipgloss.NewStyle().Foreground(colorMagenta)
	styleNew       = lipgloss.NewStyle().Foreground(colorGreen)
	styleUntracked = lipgloss.NewStyle().Foreground(colorYellow)
	styleYellow    = lipgloss.NewStyle().Foreground(colorYellow)
)