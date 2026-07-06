package ui

import (
	"runtime/debug"
	"strings"
)

// buildVersion pode ser injetada via -ldflags no go install.
var buildVersion string

func Version() string {
	if v := strings.TrimSpace(buildVersion); v != "" {
		return normalizeVersion(v)
	}

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}

	if v := info.Main.Version; v != "" && v != "(devel)" {
		return normalizeVersion(v)
	}

	if rev := vcsSetting(info, "vcs.revision"); len(rev) >= 7 {
		return rev[:7]
	}
	if rev := vcsSetting(info, "vcs.revision"); rev != "" {
		return rev
	}

	return "dev"
}

func SetBuildVersion(v string) {
	buildVersion = v
}

func vcsSetting(info *debug.BuildInfo, key string) string {
	for _, s := range info.Settings {
		if s.Key == key {
			return s.Value
		}
	}
	return ""
}

func normalizeVersion(v string) string {
	v = strings.TrimSpace(v)
	if v == "" || v == "(devel)" {
		return "dev"
	}
	if strings.HasPrefix(v, "v") {
		return v
	}
	if strings.Contains(v, ".") && !strings.Contains(v, "/") {
		return "v" + v
	}
	return v
}
