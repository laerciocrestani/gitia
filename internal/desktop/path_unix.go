//go:build !windows

package desktop

func unixCLIBinDirs() []string {
	return []string{
		"/opt/homebrew/bin",
		"/usr/local/bin",
		"/usr/local/sbin",
	}
}
