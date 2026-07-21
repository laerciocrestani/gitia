package desktop

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var augmentPathOnce sync.Once

// AugmentUserPath prepends common CLI install locations to PATH.
// GUI apps on macOS often inherit a minimal PATH and miss Homebrew (/opt/homebrew/bin).
func AugmentUserPath() {
	augmentPathOnce.Do(func() {
		path := os.Getenv("PATH")
		seen := map[string]bool{}
		for _, p := range filepath.SplitList(path) {
			if p != "" {
				seen[p] = true
			}
		}

		var extras []string
		if home, err := os.UserHomeDir(); err == nil {
			extras = append(extras,
				filepath.Join(home, ".local", "bin"),
				filepath.Join(home, "bin"),
				filepath.Join(home, "go", "bin"),
			)
		}
		extras = append(extras, unixCLIBinDirs()...)

		var prepend []string
		for _, dir := range extras {
			if dir == "" || seen[dir] {
				continue
			}
			if st, err := os.Stat(dir); err != nil || !st.IsDir() {
				continue
			}
			seen[dir] = true
			prepend = append(prepend, dir)
		}
		if len(prepend) == 0 {
			return
		}
		_ = os.Setenv("PATH", strings.Join(append(prepend, path), string(os.PathListSeparator)))
	})
}
