package setup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/laerciocrestani/gitia/internal/ui"
)

const moduleID = "github.com/laerciocrestani/gitia"

func Install() error {
	sess := ui.New("install", false)
	sess.Header()

	if err := sess.Step("Checking Go toolchain", func() error {
		return requireGo(sess)
	}); err != nil {
		return err
	}

	sess.Step("Checking optional tools", func() error {
		checkOptionalTools(sess)
		return nil
	})

	var root string
	if err := sess.Step("Locating repository", func() error {
		var err error
		root, err = FindRepoRoot()
		return err
	}); err != nil {
		return err
	}

	if err := sess.Step("Building binary", func() error {
		return goInstall(root)
	}); err != nil {
		return err
	}

	var bin string
	if err := sess.Step("Verifying installation", func() error {
		var err error
		bin, err = GitiaBin()
		if err != nil {
			return fmt.Errorf("instalação falhou — binário não encontrado em %s", GoBinDir())
		}
		sess.Detail(bin)
		return nil
	}); err != nil {
		return err
	}

	if err := sess.Step("Configuring PATH", func() error {
		return ensurePath(sess)
	}); err != nil {
		return err
	}

	sess.Detail("Próximo passo: gitia config")
	sess.Success("Installation complete 🚀")
	return nil
}

func Update() error {
	sess := ui.New("update", false)
	sess.Header()

	if err := sess.Step("Checking Go toolchain", func() error {
		return requireGo(sess)
	}); err != nil {
		return err
	}

	var root string
	if err := sess.Step("Locating repository", func() error {
		var err error
		root, err = FindRepoRoot()
		if err != nil {
			return fmt.Errorf("rode update dentro do clone do repositório gitia")
		}
		return nil
	}); err != nil {
		return err
	}

	before, err := gitShortHash(root)
	if err != nil {
		return err
	}

	branch, err := gitOutput(root, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return err
	}

	if err := sess.Step("Fetching updates", func() error {
		if err := gitRun(root, "fetch", "origin", branch); err != nil {
			_ = gitRun(root, "fetch", "origin")
		}
		return nil
	}); err != nil {
		return err
	}

	if err := sess.Step("Pulling changes", func() error {
		if err := gitRun(root, "pull", "--ff-only", "origin", branch); err != nil {
			return gitRun(root, "pull", "--ff-only")
		}
		return nil
	}); err != nil {
		return err
	}

	after, err := gitShortHash(root)
	if err != nil {
		return err
	}

	if err := sess.Step("Rebuilding binary", func() error {
		return goInstall(root)
	}); err != nil {
		return err
	}

	bin, err := GitiaBin()
	if err != nil {
		return fmt.Errorf("reinstalação falhou")
	}

	if before == after {
		sess.Info(fmt.Sprintf("Already on latest version (%s)", after))
	} else {
		sess.Detail(fmt.Sprintf("%s → %s", before, after))
		if line, err := gitOutput(root, "log", "-1", "--oneline"); err == nil {
			sess.Detail(line)
		}
	}
	sess.Detail(bin)
	sess.Success("Update complete 🚀")
	return nil
}

func FindRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		modPath := filepath.Join(dir, "go.mod")
		data, err := os.ReadFile(modPath)
		if err == nil && strings.Contains(string(data), moduleID) {
			if _, err := os.Stat(filepath.Join(dir, "cmd", "gitia", "main.go")); err == nil {
				return dir, nil
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("repositório gitia não encontrado — rode dentro do clone ou use: go run ./cmd/gitia install")
}

func GoBinDir() string {
	if out, err := exec.Command("go", "env", "GOPATH").Output(); err == nil {
		return filepath.Join(strings.TrimSpace(string(out)), "bin")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "go", "bin")
}

func GitiaBin() (string, error) {
	candidate := filepath.Join(GoBinDir(), binaryName())
	if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
		if runtime.GOOS != "windows" {
			if st.Mode()&0111 != 0 {
				return candidate, nil
			}
		} else {
			return candidate, nil
		}
	}

	if path, err := exec.LookPath(binaryName()); err == nil {
		return path, nil
	}

	return "", fmt.Errorf("binário não encontrado")
}

func ensurePath(sess *ui.Session) error {
	goBin := GoBinDir()
	if pathContains(goBin) {
		sess.Detail("PATH already includes " + goBin)
		return nil
	}

	shellRC := shellRCFile()
	if shellRC != "" {
		if data, err := os.ReadFile(shellRC); err == nil && strings.Contains(string(data), goBin) {
			sess.Detail("PATH entry already exists in " + shellRC)
			return nil
		}
	}

	if shellRC == "" {
		sess.Warn("Could not detect ~/.zshrc or ~/.bashrc")
		sess.Detail(`Add manually: export PATH="$PATH:` + goBin + `"`)
		return nil
	}

	f, err := os.OpenFile(shellRC, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	block := fmt.Sprintf("\n# gitia (Go bin)\nexport PATH=\"$PATH:%s\"\n", goBin)
	if _, err := f.WriteString(block); err != nil {
		return err
	}

	sess.Detail("Added to " + shellRC)
	sess.Warn("Run: source " + shellRC)
	return nil
}

func goInstall(root string) error {
	version, _ := gitOutput(root, "describe", "--tags", "--always", "--dirty")
	ldflags := ""
	if version != "" {
		ldflags = fmt.Sprintf("-X github.com/laerciocrestani/gitia/internal/ui.buildVersion=%s", version)
	}

	args := []string{"install"}
	if ldflags != "" {
		args = append(args, "-ldflags", ldflags)
	}
	args = append(args, "./cmd/gitia")

	cmd := exec.Command("go", args...)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func requireGo(sess *ui.Session) error {
	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("Go não encontrado — instale em https://go.dev/dl/")
	}
	if out, err := exec.Command("go", "env", "GOVERSION").Output(); err == nil {
		sess.Detail(strings.TrimPrefix(strings.TrimSpace(string(out)), "go"))
	}
	return nil
}

func checkOptionalTools(sess *ui.Session) {
	if _, err := exec.LookPath("git"); err != nil {
		sess.Warn("git not found — required for gitia")
	}
	if _, err := exec.LookPath("gh"); err != nil {
		sess.Warn("gh not found — required only for gitia pr")
		return
	}
	cmd := exec.Command("gh", "auth", "status")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		sess.Warn("gh not authenticated — run: gh auth login")
	}
}

func gitShortHash(root string) (string, error) {
	return gitOutput(root, "rev-parse", "--short", "HEAD")
}

func gitOutput(root string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(out)), nil
}

func gitRun(root string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func shellRCFile() string {
	if runtime.GOOS == "windows" {
		return ""
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	if shell := os.Getenv("SHELL"); strings.Contains(shell, "zsh") {
		return filepath.Join(home, ".zshrc")
	}
	if _, err := os.Stat(filepath.Join(home, ".zshrc")); err == nil {
		return filepath.Join(home, ".zshrc")
	}
	return filepath.Join(home, ".bashrc")
}

func pathContains(dir string) bool {
	for _, part := range filepath.SplitList(os.Getenv("PATH")) {
		if part == dir {
			return true
		}
	}
	return false
}

func binaryName() string {
	if runtime.GOOS == "windows" {
		return "gitia.exe"
	}
	return "gitia"
}
