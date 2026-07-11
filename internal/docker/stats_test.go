package docker

import "testing"

func TestParseDockerSize(t *testing.T) {
	cases := []struct {
		in   string
		min  uint64
	}{
		{"48.01MiB", 48 * 1024 * 1024},
		{"3.827GiB", 3 * 1024 * 1024 * 1024},
		{"512KiB", 512 * 1024},
	}
	for _, tc := range cases {
		got, err := parseDockerSize(tc.in)
		if err != nil {
			t.Fatalf("%q: %v", tc.in, err)
		}
		if got < tc.min {
			t.Fatalf("%q = %d want >= %d", tc.in, got, tc.min)
		}
	}
}

func TestParseMemUsagePair(t *testing.T) {
	usage, limit, err := parseMemUsagePair("67.32MiB / 3.827GiB")
	if err != nil {
		t.Fatal(err)
	}
	if usage == 0 || limit == 0 {
		t.Fatalf("usage=%d limit=%d", usage, limit)
	}
}

func TestFormatBytes(t *testing.T) {
	if got := FormatBytes(67 * 1024 * 1024); got != "67 MiB" {
		t.Fatalf("got %q", got)
	}
}

func TestMemorySummaryPercent(t *testing.T) {
	s := MemorySummary{UsedBytes: 512 * 1024 * 1024, LimitBytes: 4 * 1024 * 1024 * 1024}
	if s.Percent() <= 0 {
		t.Fatal("expected positive percent")
	}
}
