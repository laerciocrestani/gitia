package docker

import (
	"os"
	"strings"
)

// Overview aggregates Docker environment status for the dashboard.
type Overview struct {
	Available     bool
	DaemonRunning bool
	ComposeFile   string
	ProjectName   string
	Containers    []ContainerSummary
	Memory        MemorySummary
	Error         string
}

// LoadOverview collects Docker status from the working directory.
func LoadOverview(workDir string) *Overview {
	ov := &Overview{
		Available: HasDocker(),
	}
	if !ov.Available {
		ov.Error = "docker CLI não encontrado no PATH"
		return ov
	}

	ov.DaemonRunning = DaemonRunning()
	if !ov.DaemonRunning {
		ov.Error = "Docker daemon não está rodando"
		return ov
	}

	if workDir == "" {
		workDir, _ = os.Getwd()
	}
	ov.ComposeFile = FindComposeFile(workDir)
	if ov.ComposeFile == "" {
		return ov
	}

	ov.ProjectName = ProjectName(ov.ComposeFile)
	containers, err := ListComposeContainers(ov.ComposeFile)
	if err != nil {
		ov.Error = err.Error()
		return ov
	}
	ov.Containers = containers
	ov.Memory = AttachContainerStats(ov.ComposeFile, ov.Containers)
	return ov
}

// SummaryLine returns a compact status label for the header.
func (o *Overview) SummaryLine() string {
	if o == nil {
		return "n/a"
	}
	if !o.Available {
		return "n/a"
	}
	if !o.DaemonRunning {
		return "off"
	}
	if o.ComposeFile == "" {
		return "ok"
	}
	running := 0
	for _, c := range o.Containers {
		if strings.EqualFold(c.State, "running") {
			running++
		}
	}
	if running == 0 {
		return "stopped"
	}
	return "ok"
}

// CanUp reports whether compose up is available.
func (o *Overview) CanUp() bool {
	return o != nil && o.Available && o.DaemonRunning && o.ComposeFile != ""
}

// CanDown reports whether compose down is available.
func (o *Overview) CanDown() bool {
	return o != nil && o.CanUp() && HasRunningContainers(o.Containers)
}

// CanLogs reports whether logs can be shown.
func (o *Overview) CanLogs() bool {
	return o != nil && o.CanUp() && len(o.Containers) > 0
}

// CanShell reports whether exec shell is available.
func (o *Overview) CanShell() bool {
	return o != nil && o.CanUp() && FirstRunningService(o.Containers) != ""
}

// DefaultService returns the first running service or first listed service.
func (o *Overview) DefaultService() string {
	if o == nil {
		return ""
	}
	if svc := FirstRunningService(o.Containers); svc != "" {
		return svc
	}
	if len(o.Containers) > 0 {
		return o.Containers[0].Service
	}
	return ""
}
