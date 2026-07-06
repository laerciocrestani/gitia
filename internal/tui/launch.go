package tui

import (
	"os"

	"github.com/laerciocrestani/gitai/internal/uiprefs"
)

const minWidth = 80
const minHeight = 24

// ShouldLaunch indica se o comando padrão deve abrir a TUI.
func ShouldLaunch() bool {
	if !uiprefs.InteractiveUIEnabled() {
		return false
	}
	return isTerminal(os.Stdout) && isTerminal(os.Stdin)
}

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func terminalTooSmall(width, height int) bool {
	return width < minWidth || height < minHeight
}
