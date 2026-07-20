package app

import (
	"fmt"
	"os"
	"strings"

	dockerpkg "github.com/laerciocrestani/openbench/internal/docker"
	"github.com/laerciocrestani/openbench/internal/ui"
)

// DockerOptions holds flags for docker commands.
type DockerOptions struct {
	WorkDir       string // optional project directory for compose discovery
	ComposeFile   string
	Service       string
	Services      []string
	Command       []string
	Build         bool
	Profile       string
	ForceRecreate bool
	NoDeps        bool
	All           bool
	Tail          int
	Follow        bool
	DryRun        bool
	Interactive   bool
}

func resolveComposeFile(opts DockerOptions) (string, error) {
	if opts.ComposeFile != "" {
		return opts.ComposeFile, nil
	}
	start := opts.WorkDir
	if start == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		start = cwd
	}
	compose := dockerpkg.FindComposeFile(start)
	if compose == "" {
		return "", fmt.Errorf("compose file não encontrado no diretório atual")
	}
	return compose, nil
}

func dockerServices(opts DockerOptions) []string {
	if len(opts.Services) > 0 {
		return opts.Services
	}
	if opts.Service != "" {
		return []string{opts.Service}
	}
	return nil
}

// RunDockerStatus prints Docker environment status.
func RunDockerStatus() error {
	sess := ui.New("docker status", false)
	sess.Header()

	ov := dockerpkg.LoadOverview("")
	printDockerOverview(sess, ov)
	return nil
}

// RunDockerPS lists compose containers.
func RunDockerPS(opts DockerOptions) error {
	sess := ui.New("docker ps", false)
	sess.Header()

	compose, err := resolveComposeFile(opts)
	if err != nil {
		return err
	}

	containers, err := dockerpkg.ListComposeContainers(compose)
	if err != nil {
		return err
	}
	if len(containers) == 0 {
		sess.Info("Nenhum container encontrado para este compose.")
		return nil
	}
	for _, c := range containers {
		line := fmt.Sprintf("%-12s %-10s %s", c.Service, c.State, c.Ports)
		if c.Health != "" {
			line += " (" + c.Health + ")"
		}
		sess.Detail(line)
	}
	return nil
}

// RunDockerUp starts compose services.
func RunDockerUp(opts DockerOptions) error {
	compose, err := resolveComposeFile(opts)
	if err != nil {
		return err
	}
	if opts.DryRun {
		fmt.Printf("[dry-run] docker compose up -d (%s)\n", compose)
		return nil
	}
	return dockerpkg.Up(dockerpkg.UpOptions{
		ComposeFile:   compose,
		Build:         opts.Build,
		Profile:       opts.Profile,
		Services:      dockerServices(opts),
		ForceRecreate: opts.ForceRecreate,
		NoDeps:        opts.NoDeps,
		DryRun:        opts.DryRun,
	})
}

// RunDockerDown stops compose services.
func RunDockerDown(opts DockerOptions) error {
	compose, err := resolveComposeFile(opts)
	if err != nil {
		return err
	}
	if opts.DryRun {
		fmt.Printf("[dry-run] docker compose down (%s)\n", compose)
		return nil
	}
	return dockerpkg.Down(dockerpkg.DownOptions{
		ComposeFile: compose,
		DryRun:      opts.DryRun,
	})
}

// RunDockerStop stops one or more compose services.
func RunDockerStop(opts DockerOptions) error {
	compose, err := resolveComposeFile(opts)
	if err != nil {
		return err
	}
	services := dockerServices(opts)
	if len(services) == 0 {
		return fmt.Errorf("informe o serviço: ob docker stop <service>")
	}
	if opts.DryRun {
		fmt.Printf("[dry-run] docker compose stop %s (%s)\n", strings.Join(services, " "), compose)
		return nil
	}
	return dockerpkg.Stop(dockerpkg.ServiceOptions{
		ComposeFile: compose,
		Services:    services,
		DryRun:      opts.DryRun,
	})
}

// RunDockerStart starts one or more compose services.
func RunDockerStart(opts DockerOptions) error {
	compose, err := resolveComposeFile(opts)
	if err != nil {
		return err
	}
	services := dockerServices(opts)
	if len(services) == 0 {
		return fmt.Errorf("informe o serviço: ob docker start <service>")
	}
	if opts.DryRun {
		fmt.Printf("[dry-run] docker compose start %s (%s)\n", strings.Join(services, " "), compose)
		return nil
	}
	return dockerpkg.Start(dockerpkg.ServiceOptions{
		ComposeFile: compose,
		Services:    services,
		DryRun:      opts.DryRun,
	})
}

// RunDockerRecreate force-recreates a compose service.
func RunDockerRecreate(opts DockerOptions) error {
	compose, err := resolveComposeFile(opts)
	if err != nil {
		return err
	}
	service := opts.Service
	if service == "" && len(opts.Services) > 0 {
		service = opts.Services[0]
	}
	if service == "" {
		return fmt.Errorf("informe o serviço")
	}
	if opts.DryRun {
		fmt.Printf("[dry-run] docker compose up -d --force-recreate --no-deps %s (%s)\n", service, compose)
		return nil
	}
	return dockerpkg.Recreate(compose, service, opts.DryRun)
}

// RunDockerExec runs a command in a service container.
func RunDockerExec(opts DockerOptions) error {
	compose, err := resolveComposeFile(opts)
	if err != nil {
		return err
	}
	service := opts.Service
	if service == "" && len(opts.Services) > 0 {
		service = opts.Services[0]
	}
	if service == "" {
		return fmt.Errorf("informe o serviço: ob docker exec <service> -- <command>")
	}
	if len(opts.Command) == 0 {
		return fmt.Errorf("informe o comando: ob docker exec <service> -- <command>")
	}
	if opts.DryRun {
		fmt.Printf("[dry-run] docker compose exec %s %s (%s)\n", service, strings.Join(opts.Command, " "), compose)
		return nil
	}
	return dockerpkg.Exec(dockerpkg.ExecOptions{
		ComposeFile: compose,
		Service:     service,
		Command:     opts.Command,
		Interactive: opts.Interactive,
	})
}

// RunDockerLogs streams or prints service logs.
func RunDockerLogs(opts DockerOptions) error {
	compose, err := resolveComposeFile(opts)
	if err != nil {
		return err
	}
	return dockerpkg.Logs(dockerpkg.LogsOptions{
		ComposeFile: compose,
		Service:     opts.Service,
		Tail:        opts.Tail,
		Follow:      opts.Follow,
	})
}

// RunDockerLogsOutput captures logs for the TUI.
func RunDockerLogsOutput(opts DockerOptions) (string, error) {
	compose, err := resolveComposeFile(opts)
	if err != nil {
		return "", err
	}
	return dockerpkg.LogsOutput(dockerpkg.LogsOptions{
		ComposeFile: compose,
		Service:     opts.Service,
		Tail:        opts.Tail,
	})
}

// RunDockerShell opens an interactive shell in a service.
func RunDockerShell(opts DockerOptions) error {
	compose, err := resolveComposeFile(opts)
	if err != nil {
		return err
	}
	service := opts.Service
	if service == "" && len(opts.Services) > 0 {
		service = opts.Services[0]
	}
	if service == "" {
		ov := dockerpkg.LoadOverview("")
		service = ov.DefaultService()
	}
	if service == "" {
		return fmt.Errorf("nenhum serviço em execução — informe o serviço: ob docker sh <service>")
	}
	return dockerpkg.Shell(compose, service)
}

// DockerComposeFile returns the resolved compose file path from snapshot or cwd.
func DockerComposeFile(snap *WorkspaceSnapshot) string {
	if snap != nil && snap.Docker != nil && snap.Docker.ComposeFile != "" {
		return snap.Docker.ComposeFile
	}
	compose, err := resolveComposeFile(DockerOptions{})
	if err != nil {
		return ""
	}
	return compose
}

func printDockerOverview(sess *ui.Session, ov *dockerpkg.Overview) {
	if ov == nil {
		sess.Info("Docker indisponível")
		return
	}
	sess.MetaRow("CLI", boolLabel(ov.Available))
	sess.MetaRow("Daemon", daemonLabel(ov))
	if ov.ComposeFile != "" {
		sess.MetaRow("Compose", ov.ComposeFile)
		sess.MetaRow("Project", ov.ProjectName)
	}
	if ov.Error != "" {
		sess.Warn(ov.Error)
	}
	for _, c := range ov.Containers {
		line := fmt.Sprintf("%s %s", c.Service, c.State)
		if c.Ports != "" {
			line += " " + c.Ports
		}
		if c.Health != "" {
			line += " (" + c.Health + ")"
		}
		sess.Detail(line)
	}
}

func boolLabel(ok bool) string {
	if ok {
		return "available"
	}
	return "missing"
}

func daemonLabel(ov *dockerpkg.Overview) string {
	if !ov.Available {
		return "n/a"
	}
	if ov.DaemonRunning {
		return "running"
	}
	return "stopped"
}

// CanDockerUp reports whether docker up is available from snapshot.
func CanDockerUp(snap *WorkspaceSnapshot) bool {
	return snap != nil && snap.Docker != nil && snap.Docker.CanUp() && !dockerpkg.HasRunningContainers(snap.Docker.Containers)
}

// CanDockerDown reports whether docker down is available.
func CanDockerDown(snap *WorkspaceSnapshot) bool {
	return snap != nil && snap.Docker != nil && snap.Docker.CanDown()
}

// CanDockerLogs reports whether docker logs view is available.
func CanDockerLogs(snap *WorkspaceSnapshot) bool {
	return snap != nil && snap.Docker != nil && snap.Docker.CanLogs()
}

// CanDockerShell reports whether docker shell is available.
func CanDockerShell(snap *WorkspaceSnapshot) bool {
	return snap != nil && snap.Docker != nil && snap.Docker.CanShell()
}

// CanDockerEnvironment reports whether the environment detail screen is available.
func CanDockerEnvironment(snap *WorkspaceSnapshot) bool {
	return snap != nil && snap.Docker != nil && snap.Docker.CanUp()
}

// CanDockerServiceUp reports whether a service can be started.
func CanDockerServiceUp(snap *WorkspaceSnapshot, service string) bool {
	if snap == nil || snap.Docker == nil || !snap.Docker.CanUp() || service == "" {
		return false
	}
	if c, ok := dockerpkg.ContainerByService(snap.Docker.Containers, service); ok {
		return !dockerpkg.IsRunningState(c.State)
	}
	return true
}

// CanDockerServiceStop reports whether a service can be stopped.
func CanDockerServiceStop(snap *WorkspaceSnapshot, service string) bool {
	if snap == nil || snap.Docker == nil || !snap.Docker.CanUp() || service == "" {
		return false
	}
	if c, ok := dockerpkg.ContainerByService(snap.Docker.Containers, service); ok {
		return dockerpkg.IsRunningState(c.State)
	}
	return false
}

// CanDockerServiceRecreate reports whether a service can be recreated.
func CanDockerServiceRecreate(snap *WorkspaceSnapshot, service string) bool {
	return snap != nil && snap.Docker != nil && snap.Docker.CanUp() && service != ""
}

// CanDockerServiceShell reports whether shell is available for a service.
func CanDockerServiceShell(snap *WorkspaceSnapshot, service string) bool {
	if snap == nil || snap.Docker == nil || !snap.Docker.CanUp() || service == "" {
		return false
	}
	if c, ok := dockerpkg.ContainerByService(snap.Docker.Containers, service); ok {
		return dockerpkg.IsRunningState(c.State)
	}
	return false
}

// CanDockerServiceExec reports whether exec is available for a service.
func CanDockerServiceExec(snap *WorkspaceSnapshot, service string) bool {
	return CanDockerServiceShell(snap, service)
}

// CanDockerProjectUp reports whether the full stack can be started.
func CanDockerProjectUp(snap *WorkspaceSnapshot) bool {
	return snap != nil && snap.Docker != nil && snap.Docker.CanUp()
}

// CanDockerProjectDown reports whether the full stack can be stopped.
func CanDockerProjectDown(snap *WorkspaceSnapshot) bool {
	return CanDockerDown(snap)
}

// DockerDefaultService returns the default service for logs/shell.
func DockerDefaultService(snap *WorkspaceSnapshot) string {
	if snap == nil || snap.Docker == nil {
		return ""
	}
	return snap.Docker.DefaultService()
}

// FormatDockerNote returns a short note for the dashboard when docker is unavailable.
func FormatDockerNote(ov *dockerpkg.Overview) string {
	if ov == nil {
		return ""
	}
	if !ov.Available {
		return "instale Docker — https://docs.docker.com/get-docker/"
	}
	if !ov.DaemonRunning {
		return "inicie o Docker daemon"
	}
	if ov.ComposeFile == "" {
		return ""
	}
	if !dockerpkg.HasRunningContainers(ov.Containers) {
		return "execute: ob docker up"
	}
	return ""
}

// DockerContainersRunning returns count of running containers.
func DockerContainersRunning(ov *dockerpkg.Overview) int {
	if ov == nil {
		return 0
	}
	n := 0
	for _, c := range ov.Containers {
		if strings.EqualFold(c.State, "running") {
			n++
		}
	}
	return n
}

// ParseExecCommand splits a user command string into argv for docker exec.
func ParseExecCommand(input string) []string {
	return strings.Fields(strings.TrimSpace(input))
}
