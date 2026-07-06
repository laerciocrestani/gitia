package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Run inicia a TUI fullscreen. Respeita GITAI_NO_UI e terminais não interativos.
func Run() error {
	if !ShouldLaunch() {
		return fmt.Errorf("TUI indisponível (não é terminal interativo ou GITAI_NO_UI/CI)")
	}

	initTheme()
	m := newApp()
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
