package desktop

import (
	"fmt"
	"strings"

	"github.com/laerciocrestani/openbench/internal/app"
	dockerpkg "github.com/laerciocrestani/openbench/internal/docker"
)

// DockerActionResult is returned after a docker mutate action.
type DockerActionResult struct {
	Action    string       `json:"action"`
	Message   string       `json:"message"`
	Dashboard *Dashboard   `json:"dashboard"`
	Docker    DockerStatus `json:"docker"`
}

// LoadDockerStatus returns docker status for a project path.
func LoadDockerStatus(projectPath string) (DockerStatus, error) {
	if strings.TrimSpace(projectPath) == "" {
		return DockerStatus{}, fmt.Errorf("no project open")
	}
	ov := dockerpkg.LoadOverview(projectPath)
	return mapDocker(ov, dockerpkg.HasDocker()), nil
}

// DockerUp runs compose up -d for the project.
func DockerUp(projectPath string, build bool) (*DockerActionResult, error) {
	return runDockerAction(projectPath, "up", func(compose string) error {
		return app.RunDockerUp(app.DockerOptions{
			WorkDir:     projectPath,
			ComposeFile: compose,
			Build:       build,
		})
	})
}

// DockerDown runs compose down for the project.
func DockerDown(projectPath string) (*DockerActionResult, error) {
	return runDockerAction(projectPath, "down", func(compose string) error {
		return app.RunDockerDown(app.DockerOptions{
			WorkDir:     projectPath,
			ComposeFile: compose,
		})
	})
}

// DockerStop stops all running compose services (or named ones).
func DockerStop(projectPath string, services []string) (*DockerActionResult, error) {
	return runDockerAction(projectPath, "stop", func(compose string) error {
		svcs := services
		if len(svcs) == 0 {
			ov := dockerpkg.LoadOverview(projectPath)
			for _, c := range ov.Containers {
				if strings.EqualFold(c.State, "running") {
					svcs = append(svcs, c.Service)
				}
			}
		}
		if len(svcs) == 0 {
			return fmt.Errorf("nenhum serviço running para stop")
		}
		return app.RunDockerStop(app.DockerOptions{
			WorkDir:     projectPath,
			ComposeFile: compose,
			Services:    svcs,
		})
	})
}

// DockerStart starts compose services (defaults to all listed containers).
func DockerStart(projectPath string, services []string) (*DockerActionResult, error) {
	return runDockerAction(projectPath, "start", func(compose string) error {
		svcs := services
		if len(svcs) == 0 {
			ov := dockerpkg.LoadOverview(projectPath)
			for _, c := range ov.Containers {
				if c.Service != "" {
					svcs = append(svcs, c.Service)
				}
			}
		}
		if len(svcs) == 0 {
			return fmt.Errorf("nenhum serviço para start")
		}
		return app.RunDockerStart(app.DockerOptions{
			WorkDir:     projectPath,
			ComposeFile: compose,
			Services:    svcs,
		})
	})
}

// DockerRecreate force-recreates a service (default: first/default service).
func DockerRecreate(projectPath, service string) (*DockerActionResult, error) {
	return runDockerAction(projectPath, "recreate", func(compose string) error {
		svc := strings.TrimSpace(service)
		if svc == "" {
			ov := dockerpkg.LoadOverview(projectPath)
			svc = ov.DefaultService()
		}
		if svc == "" {
			return fmt.Errorf("informe o serviço para recreate")
		}
		return app.RunDockerRecreate(app.DockerOptions{
			WorkDir:     projectPath,
			ComposeFile: compose,
			Service:     svc,
		})
	})
}

func runDockerAction(projectPath, action string, fn func(compose string) error) (*DockerActionResult, error) {
	if strings.TrimSpace(projectPath) == "" {
		return nil, fmt.Errorf("no project open")
	}
	ov := dockerpkg.LoadOverview(projectPath)
	if !ov.Available {
		return nil, fmt.Errorf("docker CLI não encontrado")
	}
	if !ov.DaemonRunning {
		return nil, fmt.Errorf("Docker daemon não está rodando")
	}
	if ov.ComposeFile == "" {
		return nil, fmt.Errorf("compose file não encontrado no projeto")
	}
	if err := fn(ov.ComposeFile); err != nil {
		return nil, err
	}
	dash, err := LoadDashboard(projectPath)
	if err != nil {
		return nil, err
	}
	docker, err := LoadDockerStatus(projectPath)
	if err != nil {
		return nil, err
	}
	dash.Docker = docker
	dash.HasDocker = docker.Available
	return &DockerActionResult{
		Action:    action,
		Message:   fmt.Sprintf("docker %s ok", action),
		Dashboard: dash,
		Docker:    docker,
	}, nil
}
