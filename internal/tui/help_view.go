package tui

import "strings"

func helpContent() string {
	lines := []string{
		"Dashboard shortcuts",
		"",
		"  a       Stage files (one, many, or git add .)",
		"  c       AI commit (preview → edit → Enter confirms)",
		"  p       Push to origin (preview → Enter confirms)",
		"  P       Create AI Pull Request (preview → Enter confirms)",
		"  d       View diff (working tree or branch)",
		"  b       Switch branch (list + context)",
		"  y       Copy commit hash",
		"  l       View commit log",
		"  s       Sync base with origin (when behind / base behind)",
		"  S       Hygiene — clean merged branches (local / full)",
		"  o       Open PR in browser",
		"  u       AI usage/cost report",
		"  h       Repository health (doctor)",
		"  i       Docker environment (services, up/down, exec)",
		"  U       Docker compose up (when all stopped)",
		"  D       Docker compose down (when running)",
		"  L       Docker logs (default service)",
		"  E       Docker shell (default service)",
		"  r       Refresh dashboard",
		"  ?       This help",
		"  q       Quit",
		"",
		"Auto-refresh (dashboard and diff)",
		"  File changes detected in ~400ms (fsnotify)",
		"  External git add/reset/branch: polling every ui_auto_refresh_seconds",
		"",
		"On diff/report/branches/add/doctor screens",
		"  ↑↓      Scroll / navigate (numbered list on branches)",
		"  esc     Back",
		"",
		"On doctor screen (h)",
		"  e       Enrich with AI explanation",
		"  r       Refresh diagnosis",
		"",
		"On environment screen (i)",
		"  ↑↓      Select service",
		"  U       Up selected service",
		"  Shift+U Up entire compose project",
		"  D       Stop selected service",
		"  Shift+D Compose down (all services)",
		"  R       Force recreate selected service",
		"  E       Interactive shell in selected service",
		"  x       Run custom exec command (e.g. php artisan migrate)",
		"  p       Presets do projeto (kits: I importa laravel)",
		"  L       Logs for selected service",
		"  r       Refresh",
		"",
		"On branches screen",
		"  n       New branch (from → template → name)",
		"  Enter   Checkout selected branch",
		"",
		"On hygiene screen (S)",
		"  ↑↓      Choose mode (full / local only)",
		"  b       Edit base branch",
		"  Enter   Run hygiene",
		"",
		"On add screen",
		"  (●)/( ) All at top — space toggles all",
		"  ●/○       Toggle individual file",
		"  space     Toggle selection (All or file)",
		"  Enter     git add selected; on All → git add .",
		"  .         git add . (all files)",
		"",
		"On commit/push/PR preview",
		"  e       Edit message/title/body",
		"  Enter   Confirm",
		"  esc     Cancel (or back to preview when editing)",
		"",
		"On PR modal",
		"  d       Toggle draft",
		"  tab     Switch title/body (when editing)",
		"",
		"Preferences in config.yaml",
		"  interactive_ui          TUI when running ob (default: true)",
		"  ui_color                Colors in CLI and TUI (default: true)",
		"  ui_auto_refresh_seconds Polling fallback in seconds (default: 5, 0=off)",
		"  ui_watch_files          Watch filesystem (default: true)",
		"  language                AI commit/PR language (pt-BR, en, etc.)",
		"",
		"Environment variables (override config)",
		"  OB_NO_UI=1      Force CLI overview instead of TUI",
		"  NO_COLOR=1      No colors (Unix convention; see no-color.org)",
		"  CI=1            No TUI or colors",
	}

	var b strings.Builder
	b.WriteString(styleSection.Render("Help"))
	b.WriteString("\n\n")
	for _, line := range lines {
		if line == "" {
			b.WriteString("\n")
			continue
		}
		if !strings.HasPrefix(line, "  ") {
			b.WriteString(styleTitle.Render(line))
		} else {
			b.WriteString(styleHint.Render(line))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func helpHelpLine() string {
	return styleKey.Render("esc") + " or " + styleKey.Render("?") + " close"
}
