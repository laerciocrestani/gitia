package app

import "testing"

func TestLoadUsageReport_defaultPeriod(t *testing.T) {
	snap, err := LoadUsageReport(ReportOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if snap.PeriodLabel != "últimas 24 horas" {
		t.Fatalf("period = %q", snap.PeriodLabel)
	}
	if len(snap.Lines) == 0 {
		t.Fatal("expected lines")
	}
}

func TestLoadUsageReport_all(t *testing.T) {
	snap, err := LoadUsageReport(ReportOptions{All: true})
	if err != nil {
		t.Fatal(err)
	}
	if snap.PeriodLabel != "todo o histórico" {
		t.Fatalf("period = %q", snap.PeriodLabel)
	}
}
