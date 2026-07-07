package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/mattn/go-runewidth"
)

var bannerGraph = []string{
	"●──●────────────●",
	"    ╲           │",
	"     ●──●──●────┘",
	"                ●",
}

var bannerTitle = []string{
	"  ┏━┓┳┏┳┓┏━┓┳",
	"  ┃ ┃┃ ┃ ┃ ┃┃",
	"  ┃ ┓┃ ┃ ┣━┫┃",
	"  ┗━┛┻ ┻ ┻ ┻┻",
}

const bannerMetaIndent = "  "
const bannerArtGap = 3

// BannerContext holds optional status lines shown below the banner art.
type BannerContext struct {
	Repo     string
	Branch   string
	Sync     string
	Provider string
	Model    string
}

func writeBanner(out io.Writer, dryRun bool, ctx *BannerContext, paint func(string, string) string) {
	fmt.Fprintln(out)

	titleWidth := maxLineWidth(bannerTitle)
	gap := strings.Repeat(" ", bannerArtGap)
	height := len(bannerTitle)

	for i := 0; i < height; i++ {
		style := bannerLineStyle(i, height)
		left := bannerTitle[i]
		right := bannerGraph[i]
		pad := strings.Repeat(" ", titleWidth-lineWidth(left))
		fmt.Fprintln(out, paint(left, style)+pad+gap+paint(right, style))
	}

	tagline := "AI-powered Git Workflow · " + Version()
	if dryRun {
		tagline += " · dry-run"
	}
	fmt.Fprintf(out, "%s%s\n", bannerMetaIndent, paint(tagline, dim))
	fmt.Fprintln(out)

	if ctx != nil {
		if ctx.Repo != "" && ctx.Branch != "" {
			status := fmt.Sprintf("%s · %s · %s", ctx.Repo, ctx.Branch, ctx.Sync)
			fmt.Fprintf(out, "%s%s\n", bannerMetaIndent, paint(status, dim))
		}
		if ctx.Provider != "" && ctx.Model != "" {
			line := fmt.Sprintf("Provider: %s · Model: %s", ctx.Provider, ctx.Model)
			fmt.Fprintf(out, "%s%s\n", bannerMetaIndent, paint(line, dim))
		}
	}

	fmt.Fprintln(out)
}

func maxLineWidth(lines []string) int {
	width := 0
	for _, line := range lines {
		if w := lineWidth(line); w > width {
			width = w
		}
	}
	return width
}

func lineWidth(s string) int {
	return runewidth.StringWidth(s)
}

// FormatBanner renders the banner as a string for reuse in TUI and other views.
func FormatBanner(dryRun bool, ctx *BannerContext, colorsEnabled bool) string {
	var buf strings.Builder
	paint := func(text, code string) string {
		if !colorsEnabled {
			return text
		}
		return code + text + reset
	}
	writeBanner(&buf, dryRun, ctx, paint)
	return buf.String()
}

func bannerLineStyle(line, total int) string {
	if total <= 1 {
		return "\033[38;2;0;255;255m"
	}
	if line >= total {
		line = total - 1
	}
	t := float64(line) / float64(total-1)
	g := int(255 * (1 - t))
	b := int(255 * (1 - t))
	return fmt.Sprintf("\033[38;2;0;%d;%dm", g, b)
}

func bannerTitleStyle(line int) string {
	return bannerLineStyle(line, len(bannerTitle))
}
