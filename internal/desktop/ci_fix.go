package desktop

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/laerciocrestani/openbench/internal/ai"
	"github.com/laerciocrestani/openbench/internal/config"
	"github.com/laerciocrestani/openbench/internal/gha"
	gitpkg "github.com/laerciocrestani/openbench/internal/git"
	"github.com/laerciocrestani/openbench/internal/redact"
)

// CIFixFileView is one proposed file write.
type CIFixFileView struct {
	Path    string `json:"path"`
	Bytes   int    `json:"bytes"`
	Preview string `json:"preview"`
}

// CIFixPreviewView is shown before ConfirmCIFix applies changes.
type CIFixPreviewView struct {
	RunID                int64           `json:"runId"`
	JobID                int64           `json:"jobId,omitempty"`
	Summary              string          `json:"summary"`
	CommitMessage        string          `json:"commitMessage"`
	Files                []CIFixFileView `json:"files"`
	Notes                []string        `json:"notes,omitempty"`
	Branch               string          `json:"branch"`
	DefaultBranchWarning string          `json:"defaultBranchWarning,omitempty"`
	LogUseful            bool            `json:"logUseful"`
	Message              string          `json:"message,omitempty"`
}

// CIFixOutcome is returned after applying the fix (and optional push).
type CIFixOutcome struct {
	Applied              int          `json:"applied"`
	CommitMessage        string       `json:"commitMessage"`
	Pushed               bool         `json:"pushed"`
	Branch               string       `json:"branch"`
	DefaultBranchWarning string       `json:"defaultBranchWarning,omitempty"`
	Message              string       `json:"message"`
	CI                   *CIWatchView `json:"ci,omitempty"`
}

// stored pending fix content (not exposed to UI in full beyond preview).
type pendingCIFix struct {
	files   []ai.CIFixFile
	message string
	runID   int64
	jobID   int64
}

var (
	ciFixMu      sync.Mutex
	ciFixPending = map[string]*pendingCIFix{} // keyed by project path
)

// PreviewCIFix fetches failed logs, asks AI for a fix, and returns a confirmable preview.
func PreviewCIFix(ctx context.Context, projectPath string, runID, jobID int64) (*CIFixPreviewView, error) {
	if strings.TrimSpace(projectPath) == "" {
		return nil, fmt.Errorf("no project open")
	}
	if runID <= 0 {
		return nil, fmt.Errorf("run id inválido")
	}
	client, err := gha.Open(projectPath)
	if err != nil {
		return nil, err
	}
	payload, err := client.FetchLog(gha.LogOptions{
		RunID:      runID,
		JobID:      jobID,
		FailedOnly: true,
		MaxBytes:   gha.DefaultMaxLogBytes,
	})
	if err != nil {
		return nil, err
	}
	window := gha.FailureWindow(payload.RedactedText, 50)
	if !redact.Useful(window) {
		return &CIFixPreviewView{
			RunID:     runID,
			JobID:     jobID,
			LogUseful: false,
			Message:   "log só com material sensível / vazio — IA não foi chamada",
			Files:     []CIFixFileView{},
		}, nil
	}

	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	provider, err := ai.New(cfg)
	if err != nil {
		return nil, err
	}
	branch := ""
	if b, err := currentBranch(projectPath); err == nil {
		branch = b
	}
	lang := cfg.Language
	if strings.TrimSpace(lang) == "" {
		lang = "pt-BR"
	}
	sug, err := provider.SuggestCIFix(ctx, window, lang, branch)
	if err != nil {
		return nil, err
	}

	def := gha.ResolveDefaultBranch(projectPath)
	warn := gha.DefaultBranchWarning(branch, def)
	view := &CIFixPreviewView{
		RunID:                runID,
		JobID:                jobID,
		Summary:              sug.Summary,
		CommitMessage:        sug.CommitMessage,
		Notes:                sug.Notes,
		Branch:               branch,
		DefaultBranchWarning: warn,
		LogUseful:            true,
		Files:                []CIFixFileView{},
	}
	if len(sug.Files) == 0 {
		view.Message = "IA não propôs arquivos — revise o log manualmente"
	}
	for _, f := range sug.Files {
		view.Files = append(view.Files, CIFixFileView{
			Path:    f.Path,
			Bytes:   len(f.Content),
			Preview: truncateRunes(f.Content, 800),
		})
	}

	ciFixMu.Lock()
	ciFixPending[projectPath] = &pendingCIFix{
		files:   append([]ai.CIFixFile(nil), sug.Files...),
		message: sug.CommitMessage,
		runID:   runID,
		jobID:   jobID,
	}
	ciFixMu.Unlock()
	return view, nil
}

// ConfirmCIFix applies the pending AI fix, commits, and optionally pushes (+ CI watch).
func ConfirmCIFix(ctx context.Context, projectPath, commitMessage string, push bool, onWatch func(CIWatchUpdate)) (*CIFixOutcome, error) {
	if strings.TrimSpace(projectPath) == "" {
		return nil, fmt.Errorf("no project open")
	}
	ciFixMu.Lock()
	pending := ciFixPending[projectPath]
	ciFixMu.Unlock()
	if pending == nil || len(pending.files) == 0 {
		return nil, fmt.Errorf("nenhuma correção pendente — rode PreviewCIFix antes")
	}
	msg := strings.TrimSpace(commitMessage)
	if msg == "" {
		msg = pending.message
	}
	if msg == "" {
		return nil, fmt.Errorf("mensagem de commit vazia")
	}

	applied := 0
	for _, f := range pending.files {
		if _, err := toolWriteFile(projectPath, f.Path, f.Content); err != nil {
			return nil, fmt.Errorf("aplicar %s: %w", f.Path, err)
		}
		applied++
	}

	branch := ""
	warn := ""
	if b, err := currentBranch(projectPath); err == nil {
		branch = b
		warn = gha.DefaultBranchWarning(b, gha.ResolveDefaultBranch(projectPath))
	}

	out := &CIFixOutcome{
		Applied:              applied,
		CommitMessage:        msg,
		Branch:               branch,
		DefaultBranchWarning: warn,
		Message:              fmt.Sprintf("%d arquivo(s) aplicados", applied),
	}

	if push {
		commitOut, err := ConfirmCommitAndPush(ctx, projectPath, msg, false, nil)
		if err != nil {
			return nil, err
		}
		out.Pushed = true
		out.CommitMessage = commitOut.Message
		out.Message += " · commit+push ok"
		if onWatch != nil {
			out.CI = WatchCIAfterPush(projectPath, branch, "", onWatch)
			if out.CI != nil && out.CI.Message != "" {
				out.Message += " · " + out.CI.Message
			}
		} else {
			// background watch left to AppService
		}
	} else {
		if _, err := ConfirmCommit(ctx, projectPath, msg); err != nil {
			return nil, err
		}
		out.Message += " · commit ok"
	}

	ciFixMu.Lock()
	delete(ciFixPending, projectPath)
	ciFixMu.Unlock()
	return out, nil
}

func currentBranch(projectPath string) (string, error) {
	repo, err := gitpkg.Open(projectPath)
	if err != nil {
		return "", err
	}
	return repo.CurrentBranch()
}
