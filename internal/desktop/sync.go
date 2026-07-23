package desktop

import (
	"fmt"
	"strings"

	"github.com/laerciocrestani/openbench/internal/app"
)

// SyncResult is returned after a successful sync.
type SyncResult struct {
	Message   string     `json:"message"`
	Logs      []string   `json:"logs"`
	Dashboard *Dashboard `json:"dashboard"`
}

// HygieneResult is returned after a successful hygiene run.
type HygieneResult struct {
	Mode      string     `json:"mode"`
	Message   string     `json:"message"`
	Logs      []string   `json:"logs"`
	Dashboard *Dashboard `json:"dashboard"`
}

// HygieneActionView describes one hygiene action button for the UI.
type HygieneActionView struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
}

// HygieneActions returns the catalog of hygiene actions for the desktop dialog.
func HygieneActions() []HygieneActionView {
	return []HygieneActionView{
		{
			ID:          app.HygieneModeFull,
			Label:       "Limpar local + remoto",
			Summary:     "Remove branches merged/absorbed no local e no GitHub",
			Description: "Fetch refs, apaga no GitHub e depois remove branches locais merged/absorbed/gone.",
		},
		{
			ID:          app.HygieneModeLocal,
			Label:       "Apagar só local",
			Summary:     "Mantém branches no GitHub",
			Description: "Fetch refs e remove apenas branches locais merged/absorbed/gone. Não apaga no remoto.",
		},
	}
}

// RunSync fetches origin and fast-forwards the local base (no prune).
func RunSync(projectPath, base string) (*SyncResult, error) {
	if strings.TrimSpace(projectPath) == "" {
		return nil, fmt.Errorf("no project open")
	}

	base = strings.TrimSpace(base)
	prog := &syncProgress{}
	if err := app.RunSync(app.SyncOptions{
		Base:     base,
		WorkDir:  projectPath,
		Progress: prog,
	}); err != nil {
		return nil, err
	}

	dash, err := LoadDashboard(projectPath)
	if err != nil {
		return nil, err
	}

	msg := prog.success
	if msg == "" {
		msg = "Synced"
	}
	return &SyncResult{
		Message:   msg,
		Logs:      prog.logs,
		Dashboard: dash,
	}, nil
}

// RunHygiene executes branch cleanup for the given mode (full|local).
func RunHygiene(projectPath, mode, base string) (*HygieneResult, error) {
	if strings.TrimSpace(projectPath) == "" {
		return nil, fmt.Errorf("no project open")
	}
	mode = strings.TrimSpace(mode)
	if mode == "" {
		mode = app.HygieneModeFull
	}

	base = strings.TrimSpace(base)
	prog := &syncProgress{}
	if err := app.RunHygiene(app.HygieneOptions{
		Mode:     mode,
		Base:     base,
		WorkDir:  projectPath,
		Progress: prog,
	}); err != nil {
		return nil, err
	}

	dash, err := LoadDashboard(projectPath)
	if err != nil {
		return nil, err
	}

	msg := prog.success
	if msg == "" {
		msg = "Hygiene complete"
	}
	return &HygieneResult{
		Mode:      mode,
		Message:   msg,
		Logs:      prog.logs,
		Dashboard: dash,
	}, nil
}

// syncProgress collects step/success messages for the desktop UI.
type syncProgress struct {
	logs    []string
	success string
}

func (p *syncProgress) Step(label string, fn func() error) error {
	p.logs = append(p.logs, label)
	return fn()
}

func (p *syncProgress) StepQuiet(fn func() error) error { return fn() }

func (p *syncProgress) Detail(msg string) {
	if strings.TrimSpace(msg) != "" {
		p.logs = append(p.logs, "  "+msg)
	}
}

func (p *syncProgress) Info(msg string) { p.append(msg) }
func (p *syncProgress) Warn(msg string) { p.append(msg) }
func (p *syncProgress) Success(msg string) {
	p.success = msg
	p.append(msg)
}

func (p *syncProgress) append(msg string) {
	if strings.TrimSpace(msg) == "" {
		return
	}
	p.logs = append(p.logs, msg)
}
