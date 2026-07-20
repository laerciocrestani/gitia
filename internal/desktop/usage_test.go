package desktop

import (
	"testing"
	"time"

	"github.com/laerciocrestani/openbench/internal/usage"
)

func TestNormalizeUsagePeriod(t *testing.T) {
	cases := map[string]string{
		"":      "24h",
		"24h":   "24h",
		"7d":    "7d",
		"30d":   "30d",
		"month": "month",
		"all":   "all",
		"DAY":   "24h",
	}
	for in, want := range cases {
		if got := normalizeUsagePeriod(in); got != want {
			t.Fatalf("%q: got %q want %q", in, got, want)
		}
	}
}

func TestBuildUsageSeriesHourly(t *testing.T) {
	loc := time.Local
	now := time.Date(2026, 7, 20, 15, 30, 0, 0, loc)
	entries := []usage.Entry{
		{Timestamp: now.Add(-2 * time.Hour), InputTokens: 100, OutputTokens: 10},
		{Timestamp: now.Add(-2 * time.Hour).Add(10 * time.Minute), InputTokens: 50, OutputTokens: 5},
		{Timestamp: now, InputTokens: 20, OutputTokens: 2},
	}
	period := usage.Period{
		Since: now.Add(-3 * time.Hour),
		Until: now,
	}
	series := buildUsageSeries(entries, period, true)
	if len(series) != 4 { // hours 12,13,14,15
		t.Fatalf("len=%d want 4: %+v", len(series), series)
	}
	var found bool
	for _, pt := range series {
		if pt.Date == seriesKey(now.Add(-2*time.Hour), true) {
			found = true
			if pt.Input != 150 || pt.Output != 15 {
				t.Fatalf("bucket: %+v", pt)
			}
		}
	}
	if !found {
		t.Fatal("missing aggregated hour bucket")
	}
}

func TestBuildUsageSeriesDaily(t *testing.T) {
	loc := time.Local
	day := time.Date(2026, 7, 18, 0, 0, 0, 0, loc)
	entries := []usage.Entry{
		{Timestamp: day.Add(3 * time.Hour), InputTokens: 10, OutputTokens: 1},
		{Timestamp: day.Add(26 * time.Hour), InputTokens: 20, OutputTokens: 2},
	}
	period := usage.Period{
		Since: day,
		Until: day.Add(48 * time.Hour),
	}
	series := buildUsageSeries(entries, period, false)
	if len(series) != 3 {
		t.Fatalf("len=%d want 3", len(series))
	}
	if series[0].Input != 10 || series[1].Input != 20 || series[2].Input != 0 {
		t.Fatalf("series=%+v", series)
	}
}

func TestLoadUsageReportAll(t *testing.T) {
	view, err := LoadUsageReport("all")
	if err != nil {
		t.Fatal(err)
	}
	if view.PeriodKey != "all" {
		t.Fatalf("periodKey=%q", view.PeriodKey)
	}
	if view.Series == nil {
		t.Fatal("series nil")
	}
	if view.Granularity != "day" {
		t.Fatalf("granularity=%q", view.Granularity)
	}
}
