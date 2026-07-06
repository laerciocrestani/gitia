package app

import (
	"fmt"
	"sort"
	"time"

	"github.com/laerciocrestani/gitai/internal/pricing"
	"github.com/laerciocrestani/gitai/internal/usage"
)

// UsageReportSnapshot agrega dados formatados para a TUI de report.
type UsageReportSnapshot struct {
	PeriodLabel string
	Lines       []string
	Empty       bool
}

// LoadUsageReport monta linhas de texto para exibição na TUI.
func LoadUsageReport(opts ReportOptions) (*UsageReportSnapshot, error) {
	period, err := usage.ResolvePeriod(usage.PeriodOptions{
		Hour:  opts.Hour,
		Hours: opts.Hours,
		Days:  opts.Days,
		Month: opts.Month,
		All:   opts.All,
	}, time.Now())
	if err != nil {
		return nil, err
	}

	report, err := usage.BuildReport(period)
	if err != nil {
		return nil, err
	}

	snap := &UsageReportSnapshot{PeriodLabel: period.Label}
	snap.Lines = formatReportLines(report, period, opts)

	ledgerPath, _ := usage.LedgerPath()
	snap.Lines = append([]string{
		fmt.Sprintf("Período: %s", period.Label),
		fmt.Sprintf("Ledger: %s", ledgerPath),
		"",
	}, snap.Lines...)

	if report.Summary.TotalEntries == 0 {
		snap.Empty = true
		snap.Lines = append(snap.Lines, "Nenhum uso registrado neste período.")
	}

	return snap, nil
}

func formatReportLines(report *usage.Report, period usage.Period, opts ReportOptions) []string {
	var lines []string

	store, _ := pricing.Load()
	if store != nil && !store.UpdatedAt.IsZero() {
		lines = append(lines, fmt.Sprintf("Preços atualizados: %s",
			store.UpdatedAt.Format("2006-01-02 15:04 UTC")))
	} else {
		lines = append(lines, "Preços não encontrados — execute: gitai pricing update")
	}
	lines = append(lines, "")

	if !opts.All {
		lines = append(lines,
			fmt.Sprintf("De:  %s", period.Since.Local().Format("2006-01-02 15:04")),
			fmt.Sprintf("Até: %s", period.Until.Local().Format("2006-01-02 15:04")),
			"",
		)
	}

	if report.Summary.TotalEntries == 0 {
		return lines
	}

	lines = append(lines,
		"Resumo",
		fmt.Sprintf("  Chamadas:       %d", report.Summary.TotalEntries),
		fmt.Sprintf("  Tokens entrada: %s", usage.FormatTokens(report.Summary.TotalInput)),
		fmt.Sprintf("  Tokens saída:   %s", usage.FormatTokens(report.Summary.TotalOutput)),
		fmt.Sprintf("  Tokens total:   %s", usage.FormatTokens(report.Summary.TotalInput+report.Summary.TotalOutput)),
	)
	if report.Summary.HasCost {
		lines = append(lines, fmt.Sprintf("  Custo total:    $%.6f USD", report.Summary.TotalCost))
	}
	lines = append(lines, "")

	if len(report.ByModel) > 0 {
		lines = append(lines, "Por modelo")
		for _, model := range sortedModelKeys(report.ByModel) {
			mu := report.ByModel[model]
			line := fmt.Sprintf("  %s — %d chamada(s) · %s in · %s out",
				model, mu.Calls,
				usage.FormatTokens(mu.InputTokens),
				usage.FormatTokens(mu.OutputTokens))
			if mu.HasCost {
				line += fmt.Sprintf(" · $%.6f USD", mu.CostUSD)
			}
			lines = append(lines, line)
		}
		lines = append(lines, "")
	}

	if len(report.ByProject) > 0 {
		lines = append(lines, "Por projeto")
		for _, project := range sortedProjectKeys(report.ByProject) {
			pu := report.ByProject[project]
			line := fmt.Sprintf("  %s — %d chamada(s) · %s in · %s out",
				project, pu.Calls,
				usage.FormatTokens(pu.InputTokens),
				usage.FormatTokens(pu.OutputTokens))
			if pu.HasCost {
				line += fmt.Sprintf(" · $%.6f USD", pu.CostUSD)
			}
			lines = append(lines, line)
		}
		lines = append(lines, "")
	}

	if len(report.Entries) > 0 {
		lines = append(lines, "Detalhes")
		entries := report.Entries
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Timestamp.After(entries[j].Timestamp)
		})
		for _, e := range entries {
			line := fmt.Sprintf("  %s · %s · %s · %s · %s in · %s out",
				e.Timestamp.Local().Format("2006-01-02 15:04"),
				e.Command,
				e.Project,
				e.Model,
				usage.FormatTokens(e.InputTokens),
				usage.FormatTokens(e.OutputTokens),
			)
			if e.CostUSD != nil {
				line += fmt.Sprintf(" · $%.6f USD", *e.CostUSD)
			}
			lines = append(lines, line)
		}
	}

	return lines
}
