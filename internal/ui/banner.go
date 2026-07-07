package ui

import (
	"fmt"
	"io"
	"strings"
)

var bannerTitle = []string{
	"  ██████╗ ██╗████████╗ █████╗ ██╗",
	"  ██╔════╝ ██║╚══██╔══╝██╔══██╗██║",
	"  ██║  ███╗██║   ██║   ███████║██║",
	"  ██║   ██║██║   ██║   ██╔══██║██║",
	"  ╚██████╔╝██║   ██║   ██║  ██║██║",
	"  ╚═════╝ ╚═╝   ╚═╝   ╚═╝  ╚═╝╚═╝",
}

const bannerMetaIndent = "  "

// BannerContext holds optional status lines shown below the banner art.
type BannerContext struct {
	Repo     string
	Branch   string
	Sync     string
	Provider string
	Model    string
}

func writeBanner(out io.Writer, dryRun bool, ctx *BannerContext, paint func(string, string) string) {
	for i, line := range bannerTitle {
		fmt.Fprintln(out, paint(line, bannerTitleStyle(i)))
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

func bannerTitleStyle(line int) string {
	n := len(bannerTitle)
	if n <= 1 {
		return "\033[38;2;0;255;255m"
	}
	t := float64(line) / float64(n-1)
	g := int(255 * (1 - t))
	b := int(255 * (1 - t))
	return fmt.Sprintf("\033[38;2;0;%d;%dm", g, b)
}
