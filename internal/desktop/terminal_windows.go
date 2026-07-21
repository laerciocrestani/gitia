//go:build windows

package desktop

import "fmt"

// TerminalSession is a stub on Windows (ConPTY TBD).
type TerminalSession struct {
	cwd   string
	label string
}

// NewTerminalSession returns an error on Windows until ConPTY is implemented.
func NewTerminalSession(cwd string, cols, rows uint16, emit func(event string, data string)) (*TerminalSession, error) {
	return nil, fmt.Errorf("terminal embutido ainda não suportado no Windows")
}

// NewDockerShellSession returns an error on Windows until ConPTY is implemented.
func NewDockerShellSession(cwd, composeFile, service string, command []string, cols, rows uint16, emit func(event string, data string)) (*TerminalSession, error) {
	return nil, fmt.Errorf("terminal embutido ainda não suportado no Windows")
}

func (s *TerminalSession) Write(data string) error {
	return fmt.Errorf("terminal fechado")
}

func (s *TerminalSession) Resize(cols, rows uint16) error {
	return fmt.Errorf("terminal fechado")
}

func (s *TerminalSession) Cwd() string {
	if s == nil {
		return ""
	}
	return s.cwd
}

func (s *TerminalSession) Label() string {
	if s == nil {
		return ""
	}
	return s.label
}

func (s *TerminalSession) Close() {}
