package ui

import "testing"

func TestNormalizeVersion(t *testing.T) {
	tests := map[string]string{
		"v1.0.0":  "v1.0.0",
		"1.0.0":   "v1.0.0",
		"(devel)": "dev",
		"":        "dev",
		"abc1234": "abc1234",
	}
	for in, want := range tests {
		if got := normalizeVersion(in); got != want {
			t.Errorf("normalizeVersion(%q) = %q, want %q", in, got, want)
		}
	}
}
