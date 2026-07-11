package tui

import (
	"strings"
	"testing"

	"github.com/laerciocrestani/openbench/internal/app"
	dockerpkg "github.com/laerciocrestani/openbench/internal/docker"
)

func TestEnvironmentView_rendersServices(t *testing.T) {
	m := newEnvironmentModel()
	m.Load(&app.WorkspaceSnapshot{
		Docker: &dockerpkg.Overview{
			Available:     true,
			DaemonRunning: true,
			ComposeFile:   "/proj/docker-compose.yml",
			ProjectName:   "proj",
			Containers: []dockerpkg.ContainerSummary{
				{Service: "app", State: "running", Ports: "8080:80"},
				{Service: "db", State: "running"},
			},
		},
	})
	m.SetSize(100, 40)

	out := m.View(100, 0)
	for _, want := range []string{"Environment", "app", "db", "running", "Service detail"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
}

func TestEnvironmentModel_selectedService(t *testing.T) {
	m := newEnvironmentModel()
	m.Load(&app.WorkspaceSnapshot{
		Docker: &dockerpkg.Overview{
			Containers: []dockerpkg.ContainerSummary{
				{Service: "app", State: "running"},
				{Service: "db", State: "exited"},
			},
		},
	})
	m.cursor = 1
	if got := m.selectedService(); got != "db" {
		t.Fatalf("selectedService = %q", got)
	}
}

func TestEnvironmentHelpLine_execMode(t *testing.T) {
	help := environmentHelpLine(nil, environmentModeExec, "app")
	if !strings.Contains(help, "enter") || !strings.Contains(help, "cancel") {
		t.Fatalf("help = %q", help)
	}
}
