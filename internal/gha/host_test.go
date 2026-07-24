package gha

import "testing"

func TestHostFromRemote(t *testing.T) {
	cases := map[string]string{
		"git@github.com:acme/app.git":                 "github.com",
		"https://github.com/acme/app.git":             "github.com",
		"git@github.example.com:org/repo.git":         "github.example.com",
		"https://github.example.com/org/repo":         "github.example.com",
		"": "github.com",
	}
	for in, want := range cases {
		if got := hostFromRemote(in); got != want {
			t.Fatalf("%q: got %q want %q", in, got, want)
		}
	}
}

func TestIsEnterpriseHost(t *testing.T) {
	if IsEnterpriseHost("github.com") {
		t.Fatal("github.com should not be enterprise")
	}
	if !IsEnterpriseHost("github.mycompany.com") {
		t.Fatal("expected enterprise")
	}
}
