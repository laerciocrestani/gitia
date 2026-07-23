package app

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/laerciocrestani/openbench/internal/ai"
	"github.com/laerciocrestani/openbench/internal/config"
	gitpkg "github.com/laerciocrestani/openbench/internal/git"
	"github.com/laerciocrestani/openbench/internal/ui"
)

// ChangeSummary aggregates working tree change statistics.
type ChangeSummary struct {
	FileCount   int
	Insertions  int
	Deletions   int
	Languages   map[string]int
	DominantDir string
}

// TUINextAction is a keyboard-oriented suggested action for the TUI.
type TUINextAction struct {
	Message string
	Key     string
	Label   string
}

// CanSync reports whether sync can run (requires a clean working tree).
func CanSync(snap *WorkspaceSnapshot) bool {
	if snap == nil || snap.Overview == nil {
		return false
	}
	return !snap.Overview.IsDirty()
}

// CanHygiene reports whether branch cleanup can run (works with a dirty tree).
func CanHygiene(snap *WorkspaceSnapshot) bool {
	return snap != nil && snap.Overview != nil && !snap.Overview.Detached
}

// BuildHeaderContext builds dashboard header data from a workspace snapshot.
func BuildHeaderContext(snap *WorkspaceSnapshot) ui.HeaderContext {
	ctx := ui.HeaderContext{}
	if snap == nil {
		return ctx
	}

	if snap.Overview != nil {
		o := snap.Overview
		ctx.Repo = repoDisplayName(o)
		if o.Detached {
			ctx.Branch = "detached HEAD"
		} else {
			ctx.Branch = o.Branch
		}
		ctx.HeadHash = o.HeadHash
		ctx.HeadFullHash = o.HeadFullHash
		ctx.Status = headerStatusLabel(o)
		ctx.Sync = headerSyncLabel(o)
		ctx.OnBase = !o.Detached && o.Branch == o.BaseBranch
	}

	ctx.AIReady = snap.ConfigErr == nil && snap.Config != nil && snap.Config.APIKey != ""
	if snap.ConfigErr == nil && snap.Config != nil {
		ctx.Provider = string(snap.Config.Provider)
		ctx.Model = snap.Config.Model
	}
	if snap.Docker != nil {
		ctx.Docker = snap.Docker.SummaryLine()
	}
	return ctx
}

// BuildBannerContext is kept for callers that still use the old name.
func BuildBannerContext(snap *WorkspaceSnapshot) ui.HeaderContext {
	return BuildHeaderContext(snap)
}

// CanPush indicates whether push is available for the current snapshot.
func CanPush(snap *WorkspaceSnapshot) bool {
	if snap == nil || snap.Overview == nil || snap.ConfigErr != nil {
		return false
	}
	o := snap.Overview
	return o.IsDirty() || o.Ahead > 0
}

// CanPR indicates whether creating a PR is available.
func CanPR(snap *WorkspaceSnapshot) bool {
	if snap == nil || snap.Overview == nil || snap.ConfigErr != nil || !snap.HasGH {
		return false
	}
	o := snap.Overview
	return o.CommitsAheadOfBase > 0 || o.IsDirty()
}

// BuildChangeSummary computes file and language statistics from the overview.
func BuildChangeSummary(o *gitpkg.Overview) ChangeSummary {
	if o == nil {
		return ChangeSummary{Languages: map[string]int{}}
	}

	summary := ChangeSummary{Languages: map[string]int{}}
	dirCounts := map[string]int{}

	for _, f := range o.FileChanges {
		summary.FileCount++
		summary.Insertions += f.Insertions
		summary.Deletions += f.Deletions

		lang := classifyFile(f.Path)
		summary.Languages[lang]++

		dir := filepath.Dir(f.Path)
		if dir != "." {
			parts := strings.Split(dir, "/")
			if len(parts) > 0 {
				dirCounts[parts[0]]++
			}
		}
	}

	max := 0
	for dir, count := range dirCounts {
		if count > max {
			max = count
			summary.DominantDir = dir
		}
	}
	return summary
}

// BuildTUINextAction picks the primary suggested action for the TUI dashboard.
func BuildTUINextAction(snap *WorkspaceSnapshot) TUINextAction {
	if snap == nil {
		return TUINextAction{Message: "Press [?] for help.", Key: "?", Label: "Help"}
	}

	summary := ChangeSummary{}
	if snap.Overview != nil {
		summary = BuildChangeSummary(snap.Overview)
	}

	for _, step := range snap.NextSteps {
		if step.Plain {
			if step.Command == "working tree clean" {
				return TUINextAction{
					Message: "Working tree is clean. Press [D] to view diff or [?] for help.",
					Key:     "d",
					Label:   "Diff",
				}
			}
			continue
		}

		action := mapStepToAction(step.Command)
		if action.Key == "" {
			continue
		}

		msg := buildActionMessage(action, summary, snap)
		return TUINextAction{Message: msg, Key: action.Key, Label: action.Label}
	}

	return TUINextAction{Message: "Press [?] for help.", Key: "?", Label: "Help"}
}

// EstimateAICost returns a formatted cost estimate for the current working tree.
func EstimateAICost(snap *WorkspaceSnapshot) string {
	if snap == nil || snap.ConfigErr != nil || snap.Config == nil {
		return ""
	}
	diffSize := 0
	if snap.Overview != nil {
		for _, f := range snap.Overview.FileChanges {
			diffSize += len(f.Path) * 40
			diffSize += f.Insertions*20 + f.Deletions*20
		}
	}
	if diffSize == 0 {
		diffSize = 500
	}
	est := ai.EstimateCost(snap.Config, strings.Repeat("x", diffSize), "commit")
	if !est.HasCost {
		return ""
	}
	return fmt.Sprintf("$%.4f", est.CostUSD)
}

// ModelContextWindow returns a human-readable context window for known models.
// Used in UI hints (índice de contexto / painel AI); unknown models return "".
func ModelContextWindow(model string) string {
	m := strings.ToLower(strings.TrimSpace(model))
	switch {
	case m == "":
		return ""
	case strings.Contains(m, "gemini-2.5-pro"), strings.Contains(m, "gemini-3.1-pro"),
		strings.Contains(m, "gemini-1.5-pro"):
		return "1M"
	case strings.HasPrefix(m, "gemini-"):
		return "128k"
	case strings.Contains(m, "gpt-4.1"), strings.Contains(m, "gpt-4o"),
		strings.Contains(m, "o3"), strings.Contains(m, "o4-mini"):
		return "128k"
	case strings.Contains(m, "gpt-4-turbo"), strings.Contains(m, "gpt-4-1106"):
		return "128k"
	case strings.Contains(m, "deepseek"):
		return "128k"
	case strings.Contains(m, "claude-3-5"), strings.Contains(m, "claude-3.5"),
		strings.Contains(m, "claude-4"), strings.Contains(m, "claude-sonnet"),
		strings.Contains(m, "claude-opus"):
		return "200k"
	default:
		return ""
	}
}

// FormatProviderName returns a display name for an AI provider.
func FormatProviderName(p config.Provider) string {
	switch p {
	case config.ProviderGemini:
		return "Gemini"
	case config.ProviderOpenAI:
		return "OpenAI"
	case config.ProviderOpenRouter:
		return "OpenRouter"
	default:
		return string(p)
	}
}

type mappedAction struct {
	Key   string
	Label string
}

func mapStepToAction(command string) mappedAction {
	switch command {
	case "ob commit":
		return mappedAction{Key: "c", Label: "Commit"}
	case "ob push":
		return mappedAction{Key: "p", Label: "Push"}
	case "ob pr":
		return mappedAction{Key: "P", Label: "PR"}
	case "ob sync":
		return mappedAction{Key: "s", Label: "Sync"}
	case "ob config":
		return mappedAction{Key: "?", Label: "Config"}
	case "ob pr view":
		return mappedAction{Key: "o", Label: "Open PR"}
	case "ob docker up":
		return mappedAction{Key: "U", Label: "Docker up"}
	case "ob docker status":
		return mappedAction{Key: "U", Label: "Docker status"}
	default:
		return mappedAction{}
	}
}

func buildActionMessage(action mappedAction, summary ChangeSummary, snap *WorkspaceSnapshot) string {
	key := action.Key
	switch action.Label {
	case "Commit":
		if summary.DominantDir != "" {
			return fmt.Sprintf(
				"The detected changes appear related to %s.\nPress [%s] to generate an AI Commit.",
				summary.DominantDir, key,
			)
		}
		if snap.Overview != nil && snap.Overview.IsDirty() {
			n := len(snap.Overview.FileChanges)
			return fmt.Sprintf(
				"%d file(s) changed. Press [%s] to generate an AI Commit.",
				n, key,
			)
		}
		return fmt.Sprintf("Press [%s] to generate an AI Commit.", key)
	case "Push":
		return fmt.Sprintf("Commits are ready to publish. Press [%s] to Push.", key)
	case "PR":
		return fmt.Sprintf("Branch is ahead of base. Press [%s] to create a Pull Request.", key)
	case "Sync":
		return fmt.Sprintf("Branch is behind remote. Press [%s] to Sync.", key)
	default:
		return fmt.Sprintf("Press [%s] for %s.", key, action.Label)
	}
}

func headerStatusLabel(o *gitpkg.Overview) string {
	if !o.IsDirty() {
		return "✓ Clean"
	}
	n := len(o.FileChanges)
	if n == 1 {
		return "1 file changed"
	}
	return fmt.Sprintf("%d files changed", n)
}

func headerSyncLabel(o *gitpkg.Overview) string {
	if o.IsDirty() {
		return ""
	}
	switch {
	case o.Ahead > 0 && o.Behind > 0:
		return fmt.Sprintf("↑%d ↓%d", o.Ahead, o.Behind)
	case o.Ahead > 0:
		return fmt.Sprintf("↑ %d ahead", o.Ahead)
	case o.Behind > 0:
		return fmt.Sprintf("↓ %d behind", o.Behind)
	default:
		return "✓ in sync"
	}
}

func classifyFile(path string) string {
	base := filepath.Base(path)
	lower := strings.ToLower(base)
	if strings.Contains(lower, "_test.") || strings.HasSuffix(lower, "_test.go") {
		return "Tests"
	}
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go":
		return "Go"
	case ".php":
		return "PHP"
	case ".js", ".mjs", ".cjs":
		return "JS"
	case ".ts", ".tsx":
		return "TS"
	case ".css", ".scss", ".sass":
		return "CSS"
	case ".html", ".htm":
		return "HTML"
	case ".yaml", ".yml":
		return "YAML"
	case ".json":
		return "JSON"
	case ".md":
		return "MD"
	case ".sql":
		return "SQL"
	case ".py":
		return "Python"
	case ".rb":
		return "Ruby"
	case ".java":
		return "Java"
	case ".rs":
		return "Rust"
	default:
		if ext != "" {
			return strings.TrimPrefix(ext, ".")
		}
		return "Other"
	}
}
