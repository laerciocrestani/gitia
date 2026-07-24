package redact

import (
	"strings"
	"testing"
)

func TestSecretsGitHubPAT(t *testing.T) {
	in := "token=ghp_abcdefghijklmnopqrstuvwxyz0123456789"
	out := Secrets(in)
	if out != "token="+Placeholder {
		t.Fatalf("got %q", out)
	}
}

func TestSecretsBearerAndAuth(t *testing.T) {
	in := "Authorization: Bearer abc.def.ghi\nBearer xyzsecretvalue"
	out := Secrets(in)
	if !strings.Contains(out, Placeholder) || strings.Contains(out, "abc.def.ghi") || strings.Contains(out, "xyzsecretvalue") {
		t.Fatalf("got %q", out)
	}
}

func TestSecretsPEM(t *testing.T) {
	in := "before\n-----BEGIN RSA PRIVATE KEY-----\nMIIE\n-----END RSA PRIVATE KEY-----\nafter"
	out := Secrets(in)
	if strings.Contains(out, "MIIE") || !strings.Contains(out, Placeholder) {
		t.Fatalf("got %q", out)
	}
}

func TestSecretsKV(t *testing.T) {
	in := `API_KEY=supersecret123 password: "hunter2"`
	out := Secrets(in)
	if strings.Contains(out, "supersecret123") || strings.Contains(out, "hunter2") {
		t.Fatalf("got %q", out)
	}
}

func TestUseful(t *testing.T) {
	if Useful(Placeholder + "\n" + Placeholder) {
		t.Fatal("only redacted should not be useful")
	}
	if !Useful("error: test failed at line 10") {
		t.Fatal("expected useful")
	}
}
