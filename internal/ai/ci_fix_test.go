package ai

import (
	"strings"
	"testing"
)

func TestParseCIFixSuggestion(t *testing.T) {
	raw := `{
	  "summary": "teste falhou por nil",
	  "commit_message": "fix: evita nil pointer no handler",
	  "files": [
	    {"path": "internal/foo.go", "content": "package foo\n"},
	    {"path": "../evil", "content": "nope"},
	    {"path": "/abs", "content": "nope"}
	  ],
	  "notes": ["ok"]
	}`
	sug, err := parseCIFixSuggestion(raw)
	if err != nil {
		t.Fatal(err)
	}
	if sug.CommitMessage == "" || sug.Summary == "" {
		t.Fatalf("incomplete: %+v", sug)
	}
	if len(sug.Files) != 1 || sug.Files[0].Path != "internal/foo.go" {
		t.Fatalf("files=%+v", sug.Files)
	}
}

func TestParseCIFixSuggestionRejectsHugeFile(t *testing.T) {
	big := strings.Repeat("a", maxCIFixFileRunes+10)
	raw := `{"summary":"x","commit_message":"fix: x","files":[{"path":"a.go","content":"` + big + `"}]}`
	if _, err := parseCIFixSuggestion(raw); err == nil {
		t.Fatal("expected size error")
	}
}
