package components_test

import (
	"strings"
	"testing"

	"github.com/laerciocrestani/gitai/internal/tui/components"
)

func TestBranchTemplateCatalogOrder(t *testing.T) {
	t.Parallel()
	catalog := components.BranchTemplateCatalog()
	if len(catalog) < 10 {
		t.Fatalf("expected catalog, got %d", len(catalog))
	}

	common := catalog[:8]
	for i := 1; i < len(common); i++ {
		if common[i-1].Prefix > common[i].Prefix {
			t.Fatalf("common not sorted: %q > %q", common[i-1].Prefix, common[i].Prefix)
		}
	}

	rest := catalog[8 : len(catalog)-1]
	for i := 1; i < len(rest); i++ {
		if rest[i-1].Prefix > rest[i].Prefix {
			t.Fatalf("rest not sorted: %q > %q", rest[i-1].Prefix, rest[i].Prefix)
		}
	}

	last := catalog[len(catalog)-1]
	if !last.Other {
		t.Fatalf("last item should be Outro")
	}
}

func TestBranchTemplateItemsSeparator(t *testing.T) {
	t.Parallel()
	items := components.BranchTemplateItems()
	foundSep := false
	for _, item := range items {
		if item.Separator {
			foundSep = true
			break
		}
	}
	if !foundSep {
		t.Fatal("expected separator after common templates")
	}
}

func TestNewBranchTemplateNameSeed(t *testing.T) {
	t.Parallel()
	hotfix := components.NewBranchTemplate{Prefix: "hotfix/", Icon: "🚑"}
	if got := hotfix.NameSeed(); got != "hotfix/" {
		t.Fatalf("NameSeed() = %q", got)
	}
	other := components.NewBranchTemplate{Other: true}
	if got := other.NameSeed(); got != "" {
		t.Fatalf("Other NameSeed() = %q", got)
	}
}

func TestNewBranchTemplateListLabel(t *testing.T) {
	t.Parallel()
	feature := components.NewBranchTemplate{
		Prefix:  "feature/",
		Icon:    "✨",
		Example: "feature/user-profile",
	}
	if got := feature.ListLabel(); got != "✨ feature/user-profile" {
		t.Fatalf("ListLabel() = %q", got)
	}
}

func TestRenderNewBranchTemplatePanel(t *testing.T) {
	items := components.BranchTemplateItems()
	selected := components.TemplateAtCursor(items, 0)
	body := components.RenderNewBranchTemplateBody(0, items, selected, 70)
	out := components.RenderNewBranchTemplatePanel(0, components.SelectableTemplateCount(items), body, 80)
	for _, want := range []string{
		"Prefixo", "Uso", "Exemplo", "New Branch · Template",
		"✨ feature/user-profile", "🐛 fix/login-error", "✏️ Outro",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("missing %q in:\n%s", want, out)
		}
	}
}
