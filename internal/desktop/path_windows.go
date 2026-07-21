//go:build windows

package desktop

import (
	"os"
	"path/filepath"
)

func unixCLIBinDirs() []string {
	var dirs []string
	if pf := os.Getenv("ProgramFiles"); pf != "" {
		dirs = append(dirs, filepath.Join(pf, "GitHub CLI"))
	}
	if pf86 := os.Getenv("ProgramFiles(x86)"); pf86 != "" {
		dirs = append(dirs, filepath.Join(pf86, "GitHub CLI"))
	}
	if local := os.Getenv("LOCALAPPDATA"); local != "" {
		dirs = append(dirs, filepath.Join(local, "GitHub CLI"))
	}
	return dirs
}
