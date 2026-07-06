package tui

import (
	"sync"

	"github.com/laerciocrestani/gitai/internal/app"
)

// ActionProgress implementa app.Progress para ações na TUI.
type ActionProgress struct {
	mu     sync.Mutex
	Status string
	Logs   []string
}

func NewActionProgress() *ActionProgress {
	return &ActionProgress{}
}

func (p *ActionProgress) Step(label string, fn func() error) error {
	p.setStatus(label + "…")
	err := fn()
	p.mu.Lock()
	defer p.mu.Unlock()
	if err != nil {
		p.Logs = append(p.Logs, "✗ "+label)
	} else {
		p.Logs = append(p.Logs, "✓ "+label)
	}
	return err
}

func (p *ActionProgress) StepQuiet(fn func() error) error {
	return fn()
}

func (p *ActionProgress) Detail(msg string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Logs = append(p.Logs, "  "+msg)
}

func (p *ActionProgress) Info(msg string) {
	p.Detail(msg)
}

func (p *ActionProgress) Success(msg string) {
	p.setStatus(msg)
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Logs = append(p.Logs, "✓ "+msg)
}

func (p *ActionProgress) setStatus(s string) {
	p.mu.Lock()
	p.Status = s
	p.mu.Unlock()
}

func (p *ActionProgress) Snapshot() (status string, logs []string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	status = p.Status
	logs = append([]string(nil), p.Logs...)
	return status, logs
}

var _ app.Progress = (*ActionProgress)(nil)
