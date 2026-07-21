package docker

import (
	"slices"
	"strings"
	"testing"
)

func TestUpArgs(t *testing.T) {
	args := upArgs(UpOptions{
		ForceRecreate: true,
		NoDeps:        true,
		Build:         true,
		Services:      []string{"app", "db"},
	})
	for _, want := range []string{"up", "-d", "--build", "--force-recreate", "--no-deps", "app", "db"} {
		if !slices.Contains(args, want) {
			t.Fatalf("upArgs missing %q: %v", want, args)
		}
	}
}

func TestStopArgs(t *testing.T) {
	args := serviceArgs("stop", []string{"app"})
	if args[0] != "stop" || args[1] != "app" {
		t.Fatalf("stopArgs = %v", args)
	}
}

func TestExecArgs(t *testing.T) {
	args := execArgs("app", true, []string{"php", "artisan", "migrate"})
	want := []string{"exec", "-it", "app", "php", "artisan", "migrate"}
	if strings.Join(args, " ") != strings.Join(want, " ") {
		t.Fatalf("execArgs = %v want %v", args, want)
	}

	args = execArgs("app", false, []string{"php", "artisan", "migrate"})
	want = []string{"exec", "-T", "app", "php", "artisan", "migrate"}
	if strings.Join(args, " ") != strings.Join(want, " ") {
		t.Fatalf("execArgs non-interactive = %v want %v", args, want)
	}
}

func TestBuildExecCommand(t *testing.T) {
	cmd, err := BuildExecCommand("/proj/docker-compose.yml", "app", true, "php", "artisan", "migrate")
	if err != nil {
		t.Fatal(err)
	}
	if cmd.Dir != "/proj" {
		t.Fatalf("Dir = %q", cmd.Dir)
	}
	got := strings.Join(cmd.Args, " ")
	for _, part := range []string{"compose", "-f", "docker-compose.yml", "exec", "-it", "app", "php", "artisan", "migrate"} {
		if !strings.Contains(got, part) {
			t.Fatalf("Args missing %q: %s", part, got)
		}
	}
}
