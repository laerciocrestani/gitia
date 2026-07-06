package tui

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

// Run inicia a TUI fullscreen. Respeita GITAI_NO_UI e terminais não interativos.
func Run() error {
	if os.Getenv("GITAI_NO_UI") != "" || os.Getenv("CI") != "" {
		return fmt.Errorf("TUI desabilitada (GITAI_NO_UI ou CI)")
	}
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return fmt.Errorf("stdout não é um terminal interativo")
	}

	m := newApp()
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
