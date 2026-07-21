package docker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// UpOptions configures docker compose up.
type UpOptions struct {
	ComposeFile   string
	Build         bool
	Profile       string
	Services      []string
	ForceRecreate bool
	NoDeps        bool
	DryRun        bool
}

// DownOptions configures docker compose down.
type DownOptions struct {
	ComposeFile string
	DryRun      bool
}

// ServiceOptions configures docker compose service actions (stop/start).
type ServiceOptions struct {
	ComposeFile string
	Services    []string
	DryRun      bool
}

// LogsOptions configures docker compose logs.
type LogsOptions struct {
	ComposeFile string
	Service     string
	Tail        int
	Follow      bool
}

// ExecOptions configures docker compose exec.
type ExecOptions struct {
	ComposeFile string
	Service     string
	Command     []string
	Interactive bool
}

func upArgs(opts UpOptions) []string {
	args := []string{"up", "-d"}
	if opts.Build {
		args = append(args, "--build")
	}
	if opts.ForceRecreate {
		args = append(args, "--force-recreate")
	}
	if opts.NoDeps {
		args = append(args, "--no-deps")
	}
	if opts.Profile != "" {
		args = append(args, "--profile", opts.Profile)
	}
	return append(args, opts.Services...)
}

func serviceArgs(subcommand string, services []string) []string {
	args := []string{subcommand}
	return append(args, services...)
}

func execArgs(service string, interactive bool, command []string) []string {
	args := []string{"exec"}
	if interactive {
		args = append(args, "-it")
	} else {
		// Disable TTY allocation for capture / non-interactive runs.
		args = append(args, "-T")
	}
	args = append(args, service)
	return append(args, command...)
}

func buildDockerComposeCmd(composeFile string, args ...string) *exec.Cmd {
	dir := composeDir(composeFile)
	full := append([]string{"compose", "-f", filepath.Base(composeFile)}, args...)
	cmd := exec.Command("docker", full...)
	cmd.Dir = dir
	return cmd
}

// BuildExecCommand builds docker compose exec for tea.ExecProcess.
func BuildExecCommand(composeFile, service string, interactive bool, command ...string) (*exec.Cmd, error) {
	if composeFile == "" {
		return nil, fmt.Errorf("compose file não encontrado")
	}
	if service == "" {
		return nil, fmt.Errorf("serviço não informado")
	}
	if len(command) == 0 {
		return nil, fmt.Errorf("comando não informado")
	}
	return buildDockerComposeCmd(composeFile, execArgs(service, interactive, command)...), nil
}

// BuildShellCommand builds an interactive shell exec command (sh).
func BuildShellCommand(composeFile, service string) (*exec.Cmd, error) {
	return BuildExecCommand(composeFile, service, true, "sh")
}

// Up runs docker compose up -d.
func Up(opts UpOptions) error {
	if opts.ComposeFile == "" {
		return fmt.Errorf("compose file não encontrado")
	}
	if opts.DryRun {
		return nil
	}
	cmd := buildDockerComposeCmd(opts.ComposeFile, upArgs(opts)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Down runs docker compose down.
func Down(opts DownOptions) error {
	if opts.ComposeFile == "" {
		return fmt.Errorf("compose file não encontrado")
	}
	if opts.DryRun {
		return nil
	}
	cmd := buildDockerComposeCmd(opts.ComposeFile, "down")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Stop runs docker compose stop for one or more services.
func Stop(opts ServiceOptions) error {
	if opts.ComposeFile == "" {
		return fmt.Errorf("compose file não encontrado")
	}
	if len(opts.Services) == 0 {
		return fmt.Errorf("serviço não informado")
	}
	if opts.DryRun {
		return nil
	}
	cmd := buildDockerComposeCmd(opts.ComposeFile, serviceArgs("stop", opts.Services)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Start runs docker compose start for one or more services.
func Start(opts ServiceOptions) error {
	if opts.ComposeFile == "" {
		return fmt.Errorf("compose file não encontrado")
	}
	if len(opts.Services) == 0 {
		return fmt.Errorf("serviço não informado")
	}
	if opts.DryRun {
		return nil
	}
	cmd := buildDockerComposeCmd(opts.ComposeFile, serviceArgs("start", opts.Services)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Recreate runs docker compose up -d --force-recreate --no-deps for a service.
func Recreate(composeFile, service string, dryRun bool) error {
	return Up(UpOptions{
		ComposeFile:   composeFile,
		Services:      []string{service},
		ForceRecreate: true,
		NoDeps:        true,
		DryRun:        dryRun,
	})
}

// Logs runs docker compose logs.
func Logs(opts LogsOptions) error {
	if opts.ComposeFile == "" {
		return fmt.Errorf("compose file não encontrado")
	}
	dir := composeDir(opts.ComposeFile)
	tail := opts.Tail
	if tail <= 0 {
		tail = 100
	}
	args := []string{"compose", "-f", filepath.Base(opts.ComposeFile), "logs", "--tail", fmt.Sprintf("%d", tail)}
	if opts.Follow {
		args = append(args, "-f")
	}
	if opts.Service != "" {
		args = append(args, opts.Service)
	}
	cmd := exec.Command("docker", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// LogsOutput captures docker compose logs without following.
func LogsOutput(opts LogsOptions) (string, error) {
	if opts.ComposeFile == "" {
		return "", fmt.Errorf("compose file não encontrado")
	}
	dir := composeDir(opts.ComposeFile)
	tail := opts.Tail
	if tail <= 0 {
		tail = 200
	}
	args := []string{"compose", "-f", filepath.Base(opts.ComposeFile), "logs", "--tail", fmt.Sprintf("%d", tail), "--no-color"}
	if opts.Service != "" {
		args = append(args, opts.Service)
	}
	cmd := exec.Command("docker", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// Exec runs a command inside a service container.
func Exec(opts ExecOptions) error {
	if opts.ComposeFile == "" {
		return fmt.Errorf("compose file não encontrado")
	}
	if opts.Service == "" {
		return fmt.Errorf("serviço não informado")
	}
	if len(opts.Command) == 0 {
		return fmt.Errorf("comando não informado")
	}
	cmd, err := BuildExecCommand(opts.ComposeFile, opts.Service, opts.Interactive, opts.Command...)
	if err != nil {
		return err
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// ExecResult is the captured output of a non-interactive compose exec.
type ExecResult struct {
	Service  string
	Command  []string
	Output   string
	ExitCode int
}

// ExecOutput runs a non-interactive command and captures combined stdout/stderr.
func ExecOutput(opts ExecOptions) (*ExecResult, error) {
	if opts.ComposeFile == "" {
		return nil, fmt.Errorf("compose file não encontrado")
	}
	if opts.Service == "" {
		return nil, fmt.Errorf("serviço não informado")
	}
	if len(opts.Command) == 0 {
		return nil, fmt.Errorf("comando não informado")
	}
	cmd, err := BuildExecCommand(opts.ComposeFile, opts.Service, false, opts.Command...)
	if err != nil {
		return nil, err
	}
	out, runErr := cmd.CombinedOutput()
	res := &ExecResult{
		Service: opts.Service,
		Command: append([]string{}, opts.Command...),
		Output:  string(out),
	}
	if runErr != nil {
		if ee, ok := runErr.(*exec.ExitError); ok {
			res.ExitCode = ee.ExitCode()
			return res, nil
		}
		return res, runErr
	}
	return res, nil
}

// Shell opens an interactive shell in the service container.
func Shell(composeFile, service string) error {
	if service == "" {
		return fmt.Errorf("serviço não informado")
	}
	err := Exec(ExecOptions{
		ComposeFile: composeFile,
		Service:     service,
		Command:     []string{"sh"},
		Interactive: true,
	})
	if err == nil {
		return nil
	}
	return Exec(ExecOptions{
		ComposeFile: composeFile,
		Service:     service,
		Command:     []string{"bash"},
		Interactive: true,
	})
}
