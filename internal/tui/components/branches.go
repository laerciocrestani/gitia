package components

import (
	"fmt"
	"strings"

	gitpkg "github.com/laerciocrestani/gitai/internal/git"
	"github.com/laerciocrestani/gitai/internal/tui/theme"
)

// RenderBranchesPanel wraps the branch picker list.
func RenderBranchesPanel(cursor, total int, base, body string, width int) string {
	title := "Branches"
	if total > 0 {
		title += fmt.Sprintf("  %d/%d", cursor+1, total)
	}
	if base != "" {
		title += " · base " + base
	}
	if strings.TrimSpace(body) == "" {
		body = theme.S.Hint.Render("  (no local branches)")
	}
	return RenderPanel(title, body, width)
}

// RenderBranchDetail renders contextual information for the selected branch.
func RenderBranchDetail(detail *gitpkg.BranchDetail, summary *gitpkg.BranchSummary, branchName, base string, width, tick int) string {
	title := "Context"
	if branchName != "" {
		title += " · " + branchName
	}

	if detail == nil {
		msg := "Loading"
		if branchName != "" {
			msg = "Loading context for " + branchName
		}
		return RenderPanel(title, RenderSpinnerLine(msg, tick), width)
	}

	var lines []string
	if base == "" {
		base = "main"
	}

	if summary != nil {
		lines = append(lines, "  "+RenderBranchStageBadge(*summary))
		if stageHint := branchStageDetailHint(*summary, base); stageHint != "" {
			lines = append(lines, theme.S.Hint.Render("  "+stageHint))
		}
	}

	headLine := detail.HeadHash
	if headLine == "" {
		headLine = "—"
	}
	lines = append(lines, theme.S.Hint.Render("  HEAD "+headLine))

	if detail.Info.Upstream != "" {
		sync := detail.Info.Upstream
		if detail.Info.Ahead > 0 || detail.Info.Behind > 0 {
			sync += fmt.Sprintf("  ↑%d ↓%d", detail.Info.Ahead, detail.Info.Behind)
		}
		lines = append(lines, theme.S.Hint.Render("  "+sync))
	}

	if detail.CommitsAheadOfBase > 0 {
		lines = append(lines, theme.S.Hint.Render(fmt.Sprintf(
			"  %d commit(s) ahead of %s",
			detail.CommitsAheadOfBase, base,
		)))
	} else if detail.Info.Name == base || strings.TrimSuffix(detail.Info.Name, "/") == base {
		lines = append(lines, theme.S.Hint.Render("  on base branch"))
	} else {
		lines = append(lines, theme.S.Hint.Render(fmt.Sprintf("  aligned with %s", base)))
	}

	if detail.FilesChanged > 0 {
		lines = append(lines, theme.S.Success.Render(fmt.Sprintf("  +%d", detail.Insertions))+
			theme.S.Hint.Render(" · ")+
			theme.S.Error.Render(fmt.Sprintf("-%d", detail.Deletions))+
			theme.S.Hint.Render(fmt.Sprintf("  vs %d file(s) vs %s", detail.FilesChanged, base)))
	}

	if len(detail.RecentCommits) > 0 {
		lines = append(lines, "")
		lines = append(lines, theme.S.Hint.Render("  Recent commits:"))
		for _, c := range detail.RecentCommits {
			lines = append(lines, theme.S.Hint.Render("  ● "+c))
		}
	}

	panelBody := strings.Join(lines, "\n")
	return RenderPanel(title, panelBody, width)
}

// RenderBranchListLine renders a single branch entry for the picker list.
func RenderBranchListLine(summary gitpkg.BranchSummary, selected bool) string {
	info := summary.BranchInfo
	prefix := "  "
	if selected {
		prefix = theme.S.Current.Render("> ")
	}

	name := info.Name
	if info.Current {
		name = theme.S.Current.Render("* " + info.Name)
	} else if selected {
		name = theme.S.Current.Render(info.Name)
	} else {
		name = theme.S.Hint.Render(info.Name)
	}

	line := prefix + name
	if summary.Stage != gitpkg.StageBase {
		line += "  " + RenderBranchStageBadge(summary)
	}
	if info.Upstream != "" {
		line += theme.S.Hint.Render("  → " + info.Upstream)
	}
	if info.Ahead > 0 || info.Behind > 0 {
		line += theme.S.Warn.Render(fmt.Sprintf("  ↑%d ↓%d", info.Ahead, info.Behind))
	}
	return line
}

// RenderBranchStageBadge renders the compact lifecycle badge for a branch.
func RenderBranchStageBadge(summary gitpkg.BranchSummary) string {
	label := branchStageLabel(summary)
	switch summary.Stage {
	case gitpkg.StageWIP, gitpkg.StageSync:
		return theme.S.Warn.Render(label)
	case gitpkg.StagePush, gitpkg.StageReady:
		return theme.S.Info.Render(label)
	case gitpkg.StagePROpen, gitpkg.StageOK:
		return theme.S.Success.Render(label)
	case gitpkg.StagePRDraft, gitpkg.StageMerged, gitpkg.StageStale:
		return theme.S.Hint.Render(label)
	case gitpkg.StageBase:
		return theme.S.Hint.Render(label)
	default:
		return theme.S.Hint.Render(label)
	}
}

func branchStageLabel(summary gitpkg.BranchSummary) string {
	switch summary.Stage {
	case gitpkg.StageBase:
		return "base"
	case gitpkg.StageWIP:
		return "wip"
	case gitpkg.StageMerged:
		return "merged"
	case gitpkg.StagePROpen:
		if summary.PRNumber > 0 {
			return fmt.Sprintf("PR #%d", summary.PRNumber)
		}
		return "PR open"
	case gitpkg.StagePRDraft:
		if summary.PRNumber > 0 {
			return fmt.Sprintf("draft #%d", summary.PRNumber)
		}
		return "draft"
	case gitpkg.StagePush:
		return "push"
	case gitpkg.StageSync:
		return "sync"
	case gitpkg.StageReady:
		return "ready"
	case gitpkg.StageOK:
		return "ok"
	case gitpkg.StageStale:
		return "stale"
	default:
		return string(summary.Stage)
	}
}

func branchStageDetailHint(summary gitpkg.BranchSummary, base string) string {
	switch summary.Stage {
	case gitpkg.StageWIP:
		return "uncommitted changes on current branch"
	case gitpkg.StagePush:
		return "local commits not pushed to remote"
	case gitpkg.StageSync:
		return "behind remote — run sync"
	case gitpkg.StagePROpen:
		return "pull request open on GitHub"
	case gitpkg.StagePRDraft:
		return "draft pull request open on GitHub"
	case gitpkg.StageMerged:
		return "merged into " + base
	case gitpkg.StageReady:
		return fmt.Sprintf("%d commit(s) ahead of %s — ready for PR", summary.CommitsAheadOfBase, base)
	case gitpkg.StageOK:
		return "aligned with " + base
	case gitpkg.StageStale:
		return "local branch without upstream"
	case gitpkg.StageBase:
		return "base branch"
	default:
		return ""
	}
}

// RenderBranchListLineNumbered renders a branch row with its position in the list.
func RenderBranchListLineNumbered(index int, summary gitpkg.BranchSummary, selected bool) string {
	num := fmt.Sprintf("%2d", index+1)
	line := RenderBranchListLine(summary, selected)
	return "  " + theme.S.Hint.Render(num) + strings.TrimPrefix(line, "  ")
}
