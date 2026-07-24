package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/laerciocrestani/openbench/internal/desktop"
)

const maxTerminalSessions = 8

type termEntry struct {
	sess     *desktop.TerminalSession
	kind     string // host | docker
	service  string
	presetID string
}

func newTerminalSessionID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", os.Getpid())
	}
	return hex.EncodeToString(b[:])
}

func (s *AppService) ensureTermsLocked() {
	if s.terms == nil {
		s.terms = make(map[string]*termEntry)
	}
}

func (s *AppService) termCountLocked() int {
	return len(s.terms)
}

func (s *AppService) makeTermEmit(id string) func(event, data string) {
	return func(event, data string) {
		s.mu.RLock()
		app := s.app
		s.mu.RUnlock()
		if app == nil {
			return
		}

		var payload []byte
		switch event {
		case "terminal:data":
			payload, _ = json.Marshal(struct {
				ID   string `json:"id"`
				Data string `json:"data"`
			}{ID: id, Data: data})
		case "terminal:exit":
			// Natural shell exit (e.g. user typed `exit`): free the slot.
			// Intentional Close() does not emit exit (see TerminalSession.waitLoop).
			s.mu.Lock()
			delete(s.terms, id)
			s.mu.Unlock()
			payload, _ = json.Marshal(struct {
				ID string `json:"id"`
			}{ID: id})
		default:
			return
		}
		app.Event.Emit(event, string(payload))
	}
}

func (s *AppService) resolveHostCwd() (string, error) {
	cwd := strings.TrimSpace(s.currentPath())
	if cwd != "" {
		return cwd, nil
	}
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return "", fmt.Errorf("não foi possível resolver o diretório home do usuário")
	}
	return home, nil
}

// TerminalStart opens a new interactive host shell and returns its session id.
func (s *AppService) TerminalStart(cols, rows uint16) (string, error) {
	cwd, err := s.resolveHostCwd()
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	s.ensureTermsLocked()
	if s.termCountLocked() >= maxTerminalSessions {
		s.mu.Unlock()
		return "", fmt.Errorf("limite de %d terminais atingido", maxTerminalSessions)
	}
	id := newTerminalSessionID()
	s.terms[id] = &termEntry{kind: "host"}
	s.mu.Unlock()

	sess, err := desktop.NewTerminalSession(cwd, cols, rows, s.makeTermEmit(id))
	if err != nil {
		s.mu.Lock()
		delete(s.terms, id)
		s.mu.Unlock()
		return "", err
	}

	s.mu.Lock()
	if e, ok := s.terms[id]; ok {
		e.sess = sess
	} else {
		sess.Close()
		s.mu.Unlock()
		return "", fmt.Errorf("sessão cancelada")
	}
	s.mu.Unlock()
	return id, nil
}

// DockerShellStart opens an interactive shell inside a compose service (PTY).
// When presetID is set and interactive, runs that preset command instead of sh.
func (s *AppService) DockerShellStart(service string, cols, rows uint16, presetID string) (string, error) {
	cwd := s.currentPath()
	if cwd == "" {
		return "", fmt.Errorf("abra um projeto para usar o terminal")
	}
	compose, argv, err := desktop.ResolveDockerShellCommand(cwd, service, presetID)
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	s.ensureTermsLocked()
	if s.termCountLocked() >= maxTerminalSessions {
		s.mu.Unlock()
		return "", fmt.Errorf("limite de %d terminais atingido", maxTerminalSessions)
	}
	id := newTerminalSessionID()
	s.terms[id] = &termEntry{
		kind:     "docker",
		service:  strings.TrimSpace(service),
		presetID: strings.TrimSpace(presetID),
	}
	s.mu.Unlock()

	emit := s.makeTermEmit(id)
	sess, err := desktop.NewDockerShellSession(cwd, compose, service, argv, cols, rows, emit)
	if err != nil && (len(argv) == 1 && argv[0] == "sh") {
		sess, err = desktop.NewDockerShellSession(cwd, compose, service, []string{"bash"}, cols, rows, emit)
	}
	if err != nil {
		s.mu.Lock()
		delete(s.terms, id)
		s.mu.Unlock()
		return "", err
	}

	s.mu.Lock()
	if e, ok := s.terms[id]; ok {
		e.sess = sess
	} else {
		sess.Close()
		s.mu.Unlock()
		return "", fmt.Errorf("sessão cancelada")
	}
	s.mu.Unlock()
	return id, nil
}

// TerminalWrite sends keystrokes / paste to the given PTY session.
func (s *AppService) TerminalWrite(id, data string) error {
	s.mu.RLock()
	e := s.terms[id]
	s.mu.RUnlock()
	if e == nil || e.sess == nil {
		return fmt.Errorf("terminal não iniciado")
	}
	return e.sess.Write(data)
}

// TerminalResize updates columns/rows for the given PTY session.
func (s *AppService) TerminalResize(id string, cols, rows uint16) error {
	s.mu.RLock()
	e := s.terms[id]
	s.mu.RUnlock()
	if e == nil || e.sess == nil {
		return nil
	}
	return e.sess.Resize(cols, rows)
}

// TerminalStop kills the given shell session.
func (s *AppService) TerminalStop(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.terms[id]
	if !ok {
		return
	}
	if e.sess != nil {
		e.sess.Close()
	}
	delete(s.terms, id)
}

// TerminalStopAll kills every open shell session.
func (s *AppService) TerminalStopAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopTerminalLocked()
}

// TerminalRestart recreates the PTY for an existing session id (same kind/meta).
func (s *AppService) TerminalRestart(id string, cols, rows uint16) (string, error) {
	s.mu.Lock()
	e, ok := s.terms[id]
	if !ok {
		s.mu.Unlock()
		return "", fmt.Errorf("sessão não encontrada")
	}
	kind, service, presetID := e.kind, e.service, e.presetID
	if e.sess != nil {
		e.sess.Close()
		e.sess = nil
	}
	s.mu.Unlock()

	emit := s.makeTermEmit(id)
	var sess *desktop.TerminalSession
	var err error

	if kind == "docker" {
		cwd := s.currentPath()
		if cwd == "" {
			return "", fmt.Errorf("abra um projeto para usar o terminal")
		}
		compose, argv, rerr := desktop.ResolveDockerShellCommand(cwd, service, presetID)
		if rerr != nil {
			return "", rerr
		}
		sess, err = desktop.NewDockerShellSession(cwd, compose, service, argv, cols, rows, emit)
		if err != nil && (len(argv) == 1 && argv[0] == "sh") {
			sess, err = desktop.NewDockerShellSession(cwd, compose, service, []string{"bash"}, cols, rows, emit)
		}
	} else {
		cwd, rerr := s.resolveHostCwd()
		if rerr != nil {
			return "", rerr
		}
		sess, err = desktop.NewTerminalSession(cwd, cols, rows, emit)
	}
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if e2, ok := s.terms[id]; ok {
		e2.sess = sess
	} else {
		sess.Close()
		return "", fmt.Errorf("sessão cancelada")
	}
	return id, nil
}

// TerminalLabel returns the session label (host / docker:…).
func (s *AppService) TerminalLabel(id string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e := s.terms[id]
	if e == nil || e.sess == nil {
		return ""
	}
	return strings.TrimSpace(e.sess.Label())
}

func (s *AppService) stopTerminalLocked() {
	for id, e := range s.terms {
		if e != nil && e.sess != nil {
			e.sess.Close()
		}
		delete(s.terms, id)
	}
}
