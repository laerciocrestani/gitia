package components

import (
	"fmt"
	"strings"

	"github.com/laerciocrestani/openbench/internal/app"
	"github.com/laerciocrestani/openbench/internal/tui/theme"
	"github.com/laerciocrestani/openbench/internal/ui"
)

// HygieneModeOption describes one hygiene preset with CLI flag and explanation.
type HygieneModeOption struct {
	Mode        string
	Label       string
	Flag        string
	Summary     string
	Description string
}

// HygieneModeCatalog returns hygiene presets in display order.
func HygieneModeCatalog() []HygieneModeOption {
	return []HygieneModeOption{
		{
			Mode:        app.HygieneModeFull,
			Label:       "Full hygiene",
			Flag:        "--full",
			Summary:     "Clean local and remote merged branches",
			Description: "Phase 1: fetch. Phase 2: find merged/absorbed. Phase 3: push --delete, fetch --prune, then branch -d/-D locally.",
		},
		{
			Mode:        app.HygieneModeLocal,
			Label:       "Local only",
			Flag:        "--local",
			Summary:     "Clean local branches; keep remotes",
			Description: "Phase 1: fetch. Phase 2: find local merged/absorbed/gone. Phase 3: delete local only. Does not delete on GitHub.",
		},
	}
}

// RenderHygieneOptionsPanel renders the hygiene mode picker with a detail table.
func RenderHygieneOptionsPanel(cursor int, modes []HygieneModeOption, base string, dirty bool, width int) string {
	inner := ui.ContentInner(width)
	var lines []string

	if dirty {
		lines = append(lines, theme.S.Hint.Render("  Dirty working tree — OK for hygiene (other branches only)"))
		lines = append(lines, "")
	}

	for i, mode := range modes {
		marker := "  "
		if i == cursor {
			marker = "> "
		}
		flag := theme.S.Key.Render(mode.Flag)
		label := mode.Label + "  " + flag
		if i == cursor {
			lines = append(lines, theme.S.Current.Render(marker+label))
		} else {
			lines = append(lines, theme.S.Hint.Render(marker+label))
		}
	}

	lines = append(lines, "")
	lines = append(lines, theme.S.Hint.Render("  Base: "+base))
	lines = append(lines, "")

	selected := modes[cursor]
	lines = append(lines, renderHygieneDetailTable(selected, base, inner))

	body := strings.Join(lines, "\n")
	return RenderPanel("Hygiene · Options", body, width)
}

func renderHygieneDetailTable(mode HygieneModeOption, base string, inner int) string {
	const colW = 14

	lines := []string{
		theme.S.Hint.Render(fmt.Sprintf("  %-*s %s", colW, "Option", mode.Label)),
		theme.S.Hint.Render(fmt.Sprintf("  %-*s %s", colW, "Flag", mode.Flag)),
		theme.S.Hint.Render(fmt.Sprintf("  %-*s %s", colW, "Summary", truncatePlain(mode.Summary, inner-colW-2))),
		"",
		theme.S.Hint.Render("  What it does"),
		"  " + wrapPlain(mode.Description, inner-2),
		"",
		theme.S.Hint.Render("  Commands"),
	}

	for _, cmd := range hygieneCommandPreview(mode, base) {
		lines = append(lines, theme.S.Hint.Render("  · "+cmd))
	}

	return strings.Join(lines, "\n")
}

func hygieneCommandPreview(mode HygieneModeOption, base string) []string {
	if base == "" {
		base = "main"
	}
	cmds := []string{
		"[1] git fetch origin --prune",
		"[2] git branch --merged " + base + " …",
		"[2] git cherry " + base + " <branch> …",
	}
	if mode.Mode == app.HygieneModeFull {
		cmds = append(cmds, "[3] git push origin --delete <branch> …")
		cmds = append(cmds, "[3] git fetch origin --prune")
	}
	cmds = append(cmds,
		"[3] git branch -d <merged> …",
		"[3] git branch -D <squash/gone> …",
	)
	return cmds
}

// RenderHygieneBaseEditor renders the base branch edit step.
func RenderHygieneBaseEditor(baseField string, width int) string {
	body := theme.S.Hint.Render("  Base branch for prune:\n\n  ") + baseField
	return RenderPanel("Hygiene · Base branch", body, width)
}

// Deprecated aliases for older call sites during migration.
func SyncModeCatalog() []HygieneModeOption { return HygieneModeCatalog() }
func RenderSyncOptionsPanel(cursor int, modes []HygieneModeOption, base string, dirty bool, width int) string {
	return RenderHygieneOptionsPanel(cursor, modes, base, dirty, width)
}
func RenderSyncBaseEditor(baseField string, width int) string {
	return RenderHygieneBaseEditor(baseField, width)
}

func wrapPlain(text string, width int) string {
	if width < 20 {
		return text
	}
	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}
	var lines []string
	var current strings.Builder
	for _, word := range words {
		add := word
		if current.Len() > 0 {
			add = " " + word
		}
		if current.Len()+len(add) > width && current.Len() > 0 {
			lines = append(lines, current.String())
			current.Reset()
			current.WriteString(word)
			continue
		}
		current.WriteString(add)
	}
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}
	return strings.Join(lines, "\n")
}
