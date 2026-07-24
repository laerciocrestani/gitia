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
