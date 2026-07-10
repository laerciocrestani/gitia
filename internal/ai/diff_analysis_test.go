package ai

import (
	"strings"
	"testing"
)

func TestChangeAreasFromStat_piDoBrasil(t *testing.T) {
	stat := ` app/Console/Commands/ReprocessMetaLeadCommand.php | 138 +++++++++++++
 app/Http/Controllers/LeadController.php           | 229 ++++++++++++++++++----
 app/Http/Controllers/PaymentController.php        | 137 +++++++------
 resources/views/customers/auto-payments.blade.php |  29 +--
 resources/views/customers/payments.blade.php      |  25 +--
 5 files changed, 422 insertions(+), 136 deletions(-)`

	areas := ChangeAreasFromStat(stat)
	if len(areas) != 4 {
		t.Fatalf("areas = %d, want 4: %+v", len(areas), areas)
	}

	keys := make([]string, len(areas))
	for i, a := range areas {
		keys[i] = a.Key
	}
	want := []string{
		"app/Console/Commands/ReprocessMetaLeadCommand",
		"app/Http/Controllers/LeadController",
		"app/Http/Controllers/PaymentController",
		"resources/views/customers",
	}
	for i, w := range want {
		if keys[i] != w {
			t.Fatalf("area[%d] = %q, want %q (all: %v)", i, keys[i], w, keys)
		}
	}
}

func TestShouldSuggestSplit(t *testing.T) {
	stat := ` app/Console/Commands/Foo.php | 1 +
 app/Http/Controllers/Bar.php | 1 +
 1 file changed, 1 insertion(+)`
	areas := ChangeAreasFromStat(stat)
	if !ShouldSuggestSplit(areas) {
		t.Fatal("expected split suggestion for multiple areas")
	}
}

func TestFormatSplitSuggestion(t *testing.T) {
	areas := []ChangeArea{
		{Key: "app/Console/Commands/Foo"},
		{Key: "app/Http/Controllers/Bar"},
	}
	msg := FormatSplitSuggestion(areas)
	if !strings.Contains(msg, "git add -p") {
		t.Fatalf("missing split hint: %s", msg)
	}
	if !strings.Contains(msg, "app/Console/Commands/Foo") {
		t.Fatalf("missing area names: %s", msg)
	}
}

func TestBuildPrompt_includesStatAndRules(t *testing.T) {
	prompt := buildPrompt("diff-body", " foo.go | 1 +", "pt-BR")
	for _, want := range []string{
		"git diff --stat",
		"foo.go",
		"TODAS as áreas alteradas",
		"não invente funcionalidades",
		"diff-body",
	} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q:\n%s", want, prompt)
		}
	}
}
