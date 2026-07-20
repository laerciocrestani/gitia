package main

import (
	"fmt"

	"github.com/laerciocrestani/openbench/internal/desktop"
)

// TerminalStart opens (or restarts) an interactive shell session in the project root.
func (s *AppService) TerminalStart(cols, rows uint16) error {
	cwd := s.currentPath()
	if cwd == "" {
		return fmt.Errorf("abra um projeto para usar o terminal")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.term != nil && s.term.Cwd() == cwd {
		if cols > 0 || rows > 0 {
			_ = s.term.Resize(cols, rows)
		}
		return nil
	}

	if s.term != nil {
		s.term.Close()
		s.term = nil
	}

	app := s.app
	emit := func(event, data string) {
		if app == nil {
			return
		}
		app.Event.Emit(event, data)
	}

	sess, err := desktop.NewTerminalSession(cwd, cols, rows, emit)
	if err != nil {
		return err
	}
	s.term = sess
	return nil
}

// TerminalWrite sends keystrokes / paste to the active PTY.
func (s *AppService) TerminalWrite(data string) error {
	s.mu.RLock()
	term := s.term
	s.mu.RUnlock()
	if term == nil {
		return fmt.Errorf("terminal não iniciado")
	}
	return term.Write(data)
}

// TerminalResize updates columns/rows for the active PTY.
func (s *AppService) TerminalResize(cols, rows uint16) error {
	s.mu.RLock()
	term := s.term
	s.mu.RUnlock()
	if term == nil {
		return nil
	}
	return term.Resize(cols, rows)
}

// TerminalStop kills the active shell session.
func (s *AppService) TerminalStop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.term != nil {
		s.term.Close()
		s.term = nil
	}
}

// TerminalRestart forces a new shell in the current project root.
func (s *AppService) TerminalRestart(cols, rows uint16) error {
	s.mu.Lock()
	if s.term != nil {
		s.term.Close()
		s.term = nil
	}
	s.mu.Unlock()
	return s.TerminalStart(cols, rows)
}

func (s *AppService) stopTerminalLocked() {
	if s.term != nil {
		s.term.Close()
		s.term = nil
	}
}
