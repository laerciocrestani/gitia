package docker

import (
	"context"
	"os/exec"
	"runtime"
	"time"
)

const daemonPingTimeout = 2 * time.Second

// HasDocker reports whether the docker CLI is available on PATH.
func HasDocker() bool {
	name := "docker"
	if runtime.GOOS == "windows" {
		name = "docker.exe"
	}
	_, err := exec.LookPath(name)
	return err == nil
}

// DaemonRunning pings the Docker daemon with a short timeout.
func DaemonRunning() bool {
	if !HasDocker() {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), daemonPingTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", "info")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}
