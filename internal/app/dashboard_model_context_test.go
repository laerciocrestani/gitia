package app

import "testing"

func TestModelContextWindow(t *testing.T) {
	cases := []struct {
		model string
		want  string
	}{
		{"", ""},
		{"gemini-2.5-flash", "128k"},
		{"gemini-2.5-pro", "1M"},
		{"openai/gpt-4o", "128k"},
		{"deepseek/deepseek-chat", "128k"},
		{"anthropic/claude-sonnet-4", "200k"},
		{"some-unknown-model-xyz", ""},
	}
	for _, tc := range cases {
		if got := ModelContextWindow(tc.model); got != tc.want {
			t.Fatalf("ModelContextWindow(%q) = %q want %q", tc.model, got, tc.want)
		}
	}
}
