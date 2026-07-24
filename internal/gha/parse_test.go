package gha

import (
	"testing"
	"time"
)

func TestParseRunListJSON(t *testing.T) {
	const raw = `[
	  {
	    "databaseId": 42,
	    "name": "CI",
	    "displayTitle": "fix: something",
	    "event": "push",
	    "status": "completed",
	    "conclusion": "failure",
	    "headBranch": "feature/ci",
	    "headSha": "abc123def",
	    "url": "https://github.com/acme/repo/actions/runs/42",
	    "workflowDatabaseId": 7,
	    "workflowName": "CI",
	    "createdAt": "2026-07-24T12:00:00Z",
	    "updatedAt": "2026-07-24T12:05:00Z",
	    "startedAt": "2026-07-24T12:00:30Z"
	  }
	]`
	runs, err := parseRunListJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(runs) != 1 {
		t.Fatalf("len=%d", len(runs))
	}
	r := runs[0]
	if r.ID != 42 || r.Conclusion != "failure" || r.HeadBranch != "feature/ci" {
		t.Fatalf("unexpected run: %+v", r)
	}
	if r.Name != "fix: something" {
		t.Fatalf("name=%q", r.Name)
	}
	if !r.StartedAt.Equal(time.Date(2026, 7, 24, 12, 0, 30, 0, time.UTC)) {
		t.Fatalf("startedAt=%v", r.StartedAt)
	}
}

func TestParseRunViewJSONJobs(t *testing.T) {
	const raw = `{
	  "databaseId": 42,
	  "name": "CI",
	  "displayTitle": "CI",
	  "event": "push",
	  "status": "completed",
	  "conclusion": "failure",
	  "headBranch": "main",
	  "headSha": "deadbeef",
	  "url": "https://example/runs/42",
	  "workflowDatabaseId": 1,
	  "workflowName": "CI",
	  "createdAt": "2026-07-24T12:00:00Z",
	  "updatedAt": "2026-07-24T12:02:00Z",
	  "startedAt": "2026-07-24T12:00:00Z",
	  "jobs": [
	    {
	      "databaseId": 99,
	      "name": "test",
	      "status": "completed",
	      "conclusion": "failure",
	      "steps": [
	        {"name": "Checkout", "status": "completed", "conclusion": "success", "number": 1},
	        {"name": "Test", "status": "completed", "conclusion": "failure", "number": 2}
	      ]
	    }
	  ]
	}`
	detail, err := parseRunViewJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if detail == nil || detail.Run.ID != 42 {
		t.Fatalf("detail=%+v", detail)
	}
	if len(detail.Jobs) != 1 || detail.Jobs[0].ID != 99 {
		t.Fatalf("jobs=%+v", detail.Jobs)
	}
	if len(detail.Jobs[0].Steps) != 2 || detail.Jobs[0].Steps[1].Conclusion != "failure" {
		t.Fatalf("steps=%+v", detail.Jobs[0].Steps)
	}
}

func TestApplyListFilterFailedAndSHA(t *testing.T) {
	runs := []WorkflowRun{
		{ID: 1, Conclusion: "success", HeadSHA: "aaa111"},
		{ID: 2, Conclusion: "failure", HeadSHA: "bbb222"},
		{ID: 3, Conclusion: "cancelled", HeadSHA: "bbb222ff"},
	}
	got := applyListFilter(runs, ListFilter{FailedOnly: true, HeadSHA: "bbb222"})
	if len(got) != 2 {
		t.Fatalf("got=%+v", got)
	}
	if got[0].ID != 2 || got[1].ID != 3 {
		t.Fatalf("ids=%d,%d", got[0].ID, got[1].ID)
	}
}

func TestFailed(t *testing.T) {
	if !Failed("completed", "FAILURE") {
		t.Fatal("expected failure")
	}
	if Failed("completed", "success") {
		t.Fatal("success should not be failed")
	}
}

func TestParseOwnerFromRemote(t *testing.T) {
	cases := map[string]string{
		"git@github.com:laerciocrestani/openbench.git": "laerciocrestani",
		"https://github.com/acme/app.git":              "acme",
		"https://github.example.com/org/repo":          "org",
	}
	for in, want := range cases {
		if got := parseOwnerFromRemote(in); got != want {
			t.Fatalf("%s: got %q want %q", in, got, want)
		}
	}
}

func TestWindowMinutes(t *testing.T) {
	start := time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC)
	runs := []WorkflowRun{{
		StartedAt: start,
		UpdatedAt: start.Add(90 * time.Second),
	}}
	if got := windowMinutes(runs); got != 1.5 {
		t.Fatalf("got=%v", got)
	}
}
