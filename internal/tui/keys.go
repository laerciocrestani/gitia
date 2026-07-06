package tui

import tea "github.com/charmbracelet/bubbletea"

type keyMsg int

const (
	keyRefresh keyMsg = iota
	keyQuit
)

func parseGlobalKey(msg tea.KeyMsg) (keyMsg, bool) {
	switch msg.String() {
	case "q", "ctrl+c":
		return keyQuit, true
	case "r":
		return keyRefresh, true
	}
	return 0, false
}

func helpLine() string {
	return styleKey.Render("r") + " refresh  " +
		styleKey.Render("q") + " quit"
}
