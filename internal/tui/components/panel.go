package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/mattn/go-runewidth"
	"github.com/laerciocrestani/gitai/internal/tui/theme"
	"github.com/laerciocrestani/gitai/internal/uiprefs"
)

const (
	panelTopGradientWidth    = 40
	panelBottomGradientWidth = 100
)

// RenderPanel renders a titled panel with optional body content.
func RenderPanel(title, body string, width int) string {
	if width < 20 {
		width = 78
	}
	inner := width - 4

	var lines []string
	lines = append(lines, buildPanelTitleLine(title, width))

	if body == "" {
		lines = append(lines, buildBoxLine("", inner, width))
	} else {
		for _, line := range strings.Split(strings.TrimSuffix(body, "\n"), "\n") {
			lines = append(lines, buildBoxLine(line, inner, width))
		}
	}
	lines = append(lines, buildPanelBottom(width))
	return strings.Join(lines, "\n") + "\n"
}

func buildPanelTitleLine(title string, width int) string {
	prefix := stylePanelTitle("╭ ") + stylePanelTitle(title) + stylePanelTitle(" ")
	gradient := buildGradientBlock(panelTopGradientWidth, topGradientDash)
	return alignPrefixAndRightBlock(prefix, gradient, width)
}

func buildPanelBottom(width int) string {
	gradientWidth := panelBottomGradientWidth
	if gradientWidth > width {
		gradientWidth = width
	}
	block := "╰" + buildGradientBlock(max(gradientWidth-1, 0), bottomGradientDash)
	return alignRight(block, width)
}

func buildGradientBlock(length int, dash func(float64) string) string {
	if length <= 0 {
		return ""
	}
	var b strings.Builder
	for i := 0; i < length; i++ {
		progress := float64(i) / float64(max(length-1, 1))
		b.WriteString(dash(progress))
	}
	return b.String()
}

func alignPrefixAndRightBlock(prefix, block string, width int) string {
	blockW := displayWidth(block)
	if blockW > width {
		return alignRight(ansi.Truncate(block, width, ""), width)
	}

	maxPrefix := width - blockW
	if displayWidth(prefix) > maxPrefix {
		prefix = ansi.Truncate(prefix, maxPrefix, "…")
	}
	prefixW := displayWidth(prefix)
	pad := width - prefixW - blockW
	if pad < 0 {
		pad = 0
	}
	return prefix + strings.Repeat(" ", pad) + block
}

func alignRight(block string, width int) string {
	blockW := displayWidth(block)
	if blockW > width {
		return ansi.Truncate(block, width, "")
	}
	pad := width - blockW
	if pad < 0 {
		pad = 0
	}
	return strings.Repeat(" ", pad) + block
}

func buildBoxLine(content string, inner, width int) string {
	w := displayWidth(content)
	if w > inner {
		content = ansi.Truncate(content, inner, "…")
		w = displayWidth(content)
	}
	pad := inner - w
	if pad < 0 {
		pad = 0
	}
	line := "│ " + content + strings.Repeat(" ", pad) + " │"
	return padDisplayWidth(line, width)
}

func displayWidth(s string) int {
	if w := lipgloss.Width(s); w > 0 || !strings.Contains(s, "\x1b") {
		return w
	}
	return runewidth.StringWidth(s)
}

func padDisplayWidth(s string, width int) string {
	w := displayWidth(s)
	if w > width {
		return ansi.Truncate(s, width, "")
	}
	if w < width {
		return s + strings.Repeat(" ", width-w)
	}
	return s
}

func stylePanelTitle(s string) string {
	if !uiprefs.ColorsEnabled() {
		return s
	}
	return theme.S.PanelTitle.Render(s)
}

func topGradientDash(progress float64) string {
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	if !uiprefs.ColorsEnabled() {
		return "─"
	}
	start := colorful.Color{R: 0.34, G: 0.84, B: 0.88}
	end := colorful.Color{R: 0.18, G: 0.18, B: 0.20}
	c := start.BlendLuv(end, progress)
	return lipgloss.NewStyle().Foreground(lipgloss.Color(c.Hex())).Render("─")
}

func bottomGradientDash(progress float64) string {
	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}
	if !uiprefs.ColorsEnabled() {
		return "─"
	}
	start := colorful.Color{R: 0.92, G: 0.92, B: 0.92}
	end := colorful.Color{R: 0.10, G: 0.10, B: 0.10}
	c := start.BlendLuv(end, progress)
	return lipgloss.NewStyle().Foreground(lipgloss.Color(c.Hex())).Render("─")
}

func truncate(s string, max int) string {
	return ansi.Truncate(s, max, "…")
}

// RenderDivider renders a horizontal divider spanning the given width.
func RenderDivider(width int) string {
	if width < 4 {
		width = 78
	}
	line := padDisplayWidth("├"+strings.Repeat("─", width-2)+"┤", width)
	return theme.S.Hint.Render(line) + "\n"
}

// PadLine pads content to the given display width.
func PadLine(left, right string, width int) string {
	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
