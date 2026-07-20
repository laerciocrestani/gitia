package desktop

import (
	"fmt"
	"strings"

	"github.com/laerciocrestani/openbench/internal/app"
)

// Sync mode identifiers exposed to the frontend.
const (
	SyncModeStandard    = "standard"
	SyncModePruneRemote = "prune_remote"
	SyncModePruneFull   = "prune_full"
)

// SyncModeView describes one sync preset for the UI picker.
type SyncModeView struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
}

// SyncResult is returned after a successful sync.
type SyncResult struct {
	Mode      string     `json:"mode"`
	Message   string     `json:"message"`
	Logs      []string   `json:"logs"`
	Dashboard *Dashboard `json:"dashboard"`
}

// SyncModes returns the catalog of sync presets for the desktop dialog.
func SyncModes() []SyncModeView {
	return []SyncModeView{
		{
			ID:          SyncModeStandard,
			Label:       "Só sincronizar base",
			Summary:     "Fetch + pull da base",
			Description: "Atualiza refs remotas (fetch --prune) e faz fast-forward da branch base com origin. Não remove branches.",
		},
		{
			ID:          SyncModePruneRemote,
			Label:       "Sync + limpar no GitHub",
			Summary:     "Remove branches merged/absorbed no remoto",
			Description: "Sync da base, depois apaga no GitHub branches já merged/absorbed. Mantém branches locais.",
		},
		{
			ID:          SyncModePruneFull,
			Label:       "Sync + limpar local e GitHub",
			Summary:     "Remove branches merged local e remoto",
			Description: "Sync da base, apaga no GitHub e depois remove branches locais merged/absorbed/gone.",
		},
	}
}

// RunSync executes sync for projectPath with the given mode and optional base.
func RunSync(projectPath, mode, base string) (*SyncResult, error) {
	if strings.TrimSpace(projectPath) == "" {
		return nil, fmt.Errorf("no project open")
	}
	mode = strings.TrimSpace(mode)
	if mode == "" {
		mode = SyncModeStandard
	}

	prune, pruneRemote, err := syncFlags(mode)
	if err != nil {
		return nil, err
	}

	base = strings.TrimSpace(base)
	prog := &syncProgress{}
	if err := app.RunSync(app.SyncOptions{
		Prune:       prune,
		PruneRemote: pruneRemote,
		Base:        base,
		WorkDir:     projectPath,
		Progress:    prog,
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
		Mode:      mode,
		Message:   msg,
		Logs:      prog.logs,
		Dashboard: dash,
	}, nil
}

func syncFlags(mode string) (prune, pruneRemote bool, err error) {
	switch mode {
	case SyncModeStandard:
		return false, false, nil
	case SyncModePruneRemote:
		return false, true, nil
	case SyncModePruneFull:
		return true, false, nil
	default:
		return false, false, fmt.Errorf("modo de sync inválido: %s", mode)
	}
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

func (p *syncProgress) Info(msg string)  { p.append(msg) }
func (p *syncProgress) Warn(msg string)  { p.append(msg) }
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
