package desktop

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/laerciocrestani/openbench/internal/ui"
)

// CIFixCLIOptions controls the CLI `ob ci fix` flow.
type CIFixCLIOptions struct {
	RunID int64
	JobID int64
	Yes   bool
	Push  bool
}

// RunCIFixCLI previews an AI fix and applies after confirmation (CLI entry).
func RunCIFixCLI(ctx context.Context, opts CIFixCLIOptions) error {
	sess := ui.New("ci fix", false)
	sess.Header()

	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	var prev *CIFixPreviewView
	if err := sess.Step("Analisando log + IA", func() error {
		var err error
		prev, err = PreviewCIFix(ctx, dir, opts.RunID, opts.JobID)
		return err
	}); err != nil {
		return err
	}

	if prev.Message != "" {
		sess.Detail(prev.Message)
	}
	sess.Detail(prev.Summary)
	if prev.DefaultBranchWarning != "" {
		sess.Detail(prev.DefaultBranchWarning)
	}
	for _, f := range prev.Files {
		sess.Detail(fmt.Sprintf("  %s (%d bytes)", f.Path, f.Bytes))
	}
	if prev.CommitMessage != "" {
		sess.Detail("commit: " + prev.CommitMessage)
	}
	if len(prev.Files) == 0 {
		return fmt.Errorf("nenhuma correção aplicável")
	}

	warn := "Aplicar correção da IA no workspace"
	if opts.Push {
		warn += ", criar commit e push (dispara nova CI)"
	} else {
		warn += " e criar commit"
	}
	if prev.DefaultBranchWarning != "" {
		warn += "\n" + prev.DefaultBranchWarning
	}
	if err := confirmCIFixCLI(warn, opts.Yes); err != nil {
		return err
	}

	var out *CIFixOutcome
	if err := sess.Step("Aplicando correção", func() error {
		var err error
		out, err = ConfirmCIFix(ctx, dir, prev.CommitMessage, opts.Push, func(upd CIWatchUpdate) {
			if upd.Message != "" {
				sess.Detail(upd.Message)
			}
		})
		return err
	}); err != nil {
		return err
	}
	if out != nil {
		sess.Detail(out.Message)
	}
	sess.Success("correção aplicada")
	return nil
}

func confirmCIFixCLI(warning string, yes bool) error {
	if yes {
		return nil
	}
	fmt.Fprintf(os.Stderr, "\n%s\nConfirmar? [y/N] ", warning)
	var line string
	if _, err := fmt.Scanln(&line); err != nil {
		return fmt.Errorf("cancelado")
	}
	line = strings.TrimSpace(strings.ToLower(line))
	if line == "y" || line == "yes" {
		return nil
	}
	return fmt.Errorf("cancelado")
}
