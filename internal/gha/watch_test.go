package gha

import "testing"

func TestAllTerminalAndSummarize(t *testing.T) {
	runs := []WorkflowRun{
		{Status: "completed", Conclusion: "success"},
		{Status: "completed", Conclusion: "failure"},
	}
	if !allTerminal(runs) {
		t.Fatal("expected terminal")
	}
	msg := summarizeWatch(runs)
	if msg == "" {
		t.Fatal("empty summary")
	}
	pending := []WorkflowRun{{Status: "in_progress", Conclusion: ""}}
	if allTerminal(pending) {
		t.Fatal("in_progress should not be terminal")
	}
}
