package app

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/laerciocrestani/openbench/internal/gha"
	"github.com/laerciocrestani/openbench/internal/ui"
)

// CIStatusOptions controls `ob ci status`.
type CIStatusOptions struct {
	FailedOnly  bool
	Limit       int
	Branch      string
	AllBranches bool
}

// RunCIStatus prints workflow runs for the current repo.
func RunCIStatus(opts CIStatusOptions) error {
	sess := ui.New("ci status", false)
	sess.Header()

	client, err := gha.New()
	if err != nil {
		return err
	}

	f := gha.ListFilter{
		FailedOnly: opts.FailedOnly,
		Limit:      opts.Limit,
		Branch:     strings.TrimSpace(opts.Branch),
	}
	if opts.AllBranches {
		f.Branch = ""
	}

	var snap *gha.StatusSnapshot
	if err := sess.Step("Listando workflow runs", func() error {
		var err error
		snap, err = client.LoadStatus(f)
		return err
	}); err != nil {
		return err
	}

	printActionsUsage(snap.Usage)
	if snap.Branch != "" {
		sess.Detail(fmt.Sprintf("branch %s · HEAD %s", snap.Branch, shortSHA(snap.HeadSHA)))
	}
	if len(snap.Runs) == 0 {
		sess.Success("nenhum run encontrado")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tSTATUS\tEVENT\tWORKFLOW\tTITLE\tSHA")
	for _, r := range snap.Runs {
		status := r.Conclusion
		if status == "" {
			status = r.Status
		}
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n",
			r.ID, status, r.Event, r.WorkflowName, truncateCI(r.Name, 40), shortSHA(r.HeadSHA))
	}
	_ = w.Flush()
	sess.Success(fmt.Sprintf("%d run(s)", len(snap.Runs)))
	return nil
}

// RunCIView prints one run with jobs/steps.
func RunCIView(runID int64) error {
	sess := ui.New("ci view", false)
	sess.Header()

	client, err := gha.New()
	if err != nil {
		return err
	}

	var detail *gha.RunDetail
	if err := sess.Step("Carregando run "+strconv.FormatInt(runID, 10), func() error {
		var err error
		detail, err = client.ViewRun(runID)
		return err
	}); err != nil {
		return err
	}

	usage := client.UsageForRun(detail.Run, client.RemoteOwner())
	printActionsUsage(usage)

	r := detail.Run
	status := r.Conclusion
	if status == "" {
		status = r.Status
	}
	sess.Detail(fmt.Sprintf("%s · %s · %s · %s", status, r.Event, r.WorkflowName, r.HeadBranch))
	if r.URL != "" {
		sess.Detail(r.URL)
	}

	if len(detail.Jobs) == 0 {
		sess.Success("run sem jobs")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "JOB\tSTATUS\tSTEPS")
	for _, j := range detail.Jobs {
		js := j.Conclusion
		if js == "" {
			js = j.Status
		}
		failedSteps := 0
		for _, s := range j.Steps {
			if gha.Failed(s.Status, s.Conclusion) {
				failedSteps++
			}
		}
		fmt.Fprintf(w, "%s (%d)\t%s\t%d total", j.Name, j.ID, js, len(j.Steps))
		if failedSteps > 0 {
			fmt.Fprintf(w, " · %d fail", failedSteps)
		}
		fmt.Fprintln(w)
		for _, s := range j.Steps {
			ss := s.Conclusion
			if ss == "" {
				ss = s.Status
			}
			mark := " "
			if gha.Failed(s.Status, s.Conclusion) {
				mark = "✗"
			} else if strings.EqualFold(ss, "success") {
				mark = "✓"
			}
			fmt.Fprintf(w, "  %s %d. %s\t%s\t\n", mark, s.Number, s.Name, ss)
		}
	}
	_ = w.Flush()
	sess.Success("ok")
	return nil
}

// RunCIUsage prints Actions usage for the current repo window.
func RunCIUsage() error {
	sess := ui.New("ci usage", false)
	sess.Header()

	client, err := gha.New()
	if err != nil {
		return err
	}

	var snap *gha.StatusSnapshot
	if err := sess.Step("Calculando uso de Actions", func() error {
		var err error
		snap, err = client.LoadStatus(gha.ListFilter{Limit: 20})
		return err
	}); err != nil {
		return err
	}

	printActionsUsage(snap.Usage)
	sess.Success("ok")
	return nil
}

// CILogsOptions controls `ob ci logs`.
type CILogsOptions struct {
	RunID      int64
	JobID      int64
	FailedOnly bool
}

// RunCILogs prints a redacted workflow log (on demand).
func RunCILogs(opts CILogsOptions) error {
	sess := ui.New("ci logs", false)
	sess.Header()

	client, err := gha.New()
	if err != nil {
		return err
	}

	var payload *gha.LogPayload
	label := "Baixando log"
	if opts.FailedOnly {
		label += " (falhos)"
	}
	if err := sess.Step(label, func() error {
		var err error
		payload, err = client.FetchLog(gha.LogOptions{
			RunID:      opts.RunID,
			JobID:      opts.JobID,
			FailedOnly: opts.FailedOnly,
			MaxBytes:   -1, // CLI: full redacted log
		})
		return err
	}); err != nil {
		return err
	}

	if payload.Message != "" {
		sess.Detail(payload.Message)
	}
	sess.Detail(fmt.Sprintf("bytes=%d raw=%d useful=%v truncated=%v",
		payload.Bytes, payload.RawBytes, payload.Useful, payload.Truncated))
	fmt.Fprintln(os.Stdout)
	fmt.Fprint(os.Stdout, payload.RedactedText)
	if !strings.HasSuffix(payload.RedactedText, "\n") {
		fmt.Fprintln(os.Stdout)
	}
	sess.Success("log redigido")
	return nil
}

// CIRerunOptions controls `ob ci rerun`.
type CIRerunOptions struct {
	RunID      int64
	JobID      int64
	FailedOnly bool
	Yes        bool
}

// RunCIRerun previews cost and re-runs after confirmation.
func RunCIRerun(opts CIRerunOptions) error {
	sess := ui.New("ci rerun", false)
	sess.Header()

	client, err := gha.New()
	if err != nil {
		return err
	}

	var prev *gha.RerunPreview
	if err := sess.Step("Preparando re-run", func() error {
		var err error
		prev, err = client.PreviewRerun(opts.RunID, gha.RerunOptions{
			JobID:      opts.JobID,
			FailedOnly: opts.FailedOnly,
		})
		return err
	}); err != nil {
		return err
	}

	printActionsUsage(prev.Usage)
	sess.Detail(prev.CostWarning)
	sess.Detail(fmt.Sprintf("run %d · %s · %s", prev.RunID, prev.RunName, prev.HeadBranch))

	if err := confirmCIAction(prev.CostWarning, opts.Yes); err != nil {
		return err
	}

	if err := sess.Step("Re-executando", func() error {
		return client.ConfirmRerun(opts.RunID, gha.RerunOptions{
			JobID:      opts.JobID,
			FailedOnly: opts.FailedOnly,
		})
	}); err != nil {
		return err
	}
	sess.Success("re-run disparado")
	return nil
}

// CIDispatchOptions controls `ob ci dispatch`.
type CIDispatchOptions struct {
	Workflow string
	Ref      string
	Fields   map[string]string
	Yes      bool
}

// RunCIDispatch previews cost and dispatches after confirmation.
func RunCIDispatch(opts CIDispatchOptions) error {
	sess := ui.New("ci dispatch", false)
	sess.Header()

	client, err := gha.New()
	if err != nil {
		return err
	}

	var prev *gha.DispatchPreview
	if err := sess.Step("Preparando dispatch", func() error {
		var err error
		prev, err = client.PreviewDispatch(opts.Workflow, opts.Ref, opts.Fields)
		return err
	}); err != nil {
		return err
	}

	printActionsUsage(prev.Usage)
	sess.Detail(prev.CostWarning)
	if !prev.CanDispatch {
		sess.Detail("aviso: workflow_dispatch não detectado no YAML")
	}

	if err := confirmCIAction(prev.CostWarning, opts.Yes); err != nil {
		return err
	}

	if err := sess.Step("Disparando workflow", func() error {
		return client.ConfirmDispatch(opts.Workflow, opts.Ref, opts.Fields)
	}); err != nil {
		return err
	}
	sess.Success("workflow_dispatch enviado")
	return nil
}

// RunCIWorkflows lists workflows.
func RunCIWorkflows() error {
	sess := ui.New("ci workflows", false)
	sess.Header()

	client, err := gha.New()
	if err != nil {
		return err
	}

	var list []gha.Workflow
	if err := sess.Step("Listando workflows", func() error {
		var err error
		list, err = client.ListWorkflows()
		return err
	}); err != nil {
		return err
	}

	if len(list) == 0 {
		sess.Success("nenhum workflow")
		return nil
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tSTATE\tNAME\tPATH")
	for _, wf := range list {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", wf.ID, wf.State, wf.Name, wf.Path)
	}
	_ = w.Flush()
	sess.Success(fmt.Sprintf("%d workflow(s)", len(list)))
	return nil
}

func confirmCIAction(warning string, yes bool) error {
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

func printActionsUsage(u gha.ActionsUsage) {
	fmt.Fprintf(os.Stdout, "Actions usage [%s]: %s\n", u.State, u.Message)
	if u.RunMinutes != nil {
		fmt.Fprintf(os.Stdout, "  run: ~%.1f min\n", *u.RunMinutes)
	}
	if u.WindowMinutes != nil {
		fmt.Fprintf(os.Stdout, "  janela listada: ~%.1f min\n", *u.WindowMinutes)
	}
	if u.OrgUsedMinutes != nil {
		fmt.Fprintf(os.Stdout, "  org usado: %.0f", *u.OrgUsedMinutes)
		if u.OrgIncludedMinutes != nil {
			fmt.Fprintf(os.Stdout, " / %.0f", *u.OrgIncludedMinutes)
		}
		if u.OrgRemainingMinutes != nil {
			fmt.Fprintf(os.Stdout, " (restante %.0f)", *u.OrgRemainingMinutes)
		}
		fmt.Fprintln(os.Stdout)
	}
}

func shortSHA(sha string) string {
	sha = strings.TrimSpace(sha)
	if len(sha) > 7 {
		return sha[:7]
	}
	return sha
}

func truncateCI(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	if n <= 1 {
		return s[:n]
	}
	return s[:n-1] + "…"
}
