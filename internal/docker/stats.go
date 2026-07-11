package docker

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// MemorySummary aggregates compose stack memory usage.
type MemorySummary struct {
	UsedBytes  uint64
	LimitBytes uint64
}

// Percent returns used/limit as 0-100, or 0 when limit is unknown.
func (m MemorySummary) Percent() float64 {
	if m.LimitBytes == 0 {
		return 0
	}
	return float64(m.UsedBytes) * 100 / float64(m.LimitBytes)
}

// AttachContainerStats enriches containers with live memory from docker stats.
func AttachContainerStats(composeFile string, containers []ContainerSummary) MemorySummary {
	if composeFile == "" || len(containers) == 0 {
		return MemorySummary{}
	}

	byName, err := loadStatsByContainerName(composeFile)
	if err != nil || len(byName) == 0 {
		return MemorySummary{}
	}

	var summary MemorySummary
	for i := range containers {
		st, ok := byName[containers[i].Name]
		if !ok {
			continue
		}
		containers[i].MemUsageBytes = st.usage
		containers[i].MemLimitBytes = st.limit
		containers[i].MemPercent = st.percent
		if strings.EqualFold(containers[i].State, "running") {
			summary.UsedBytes += st.usage
		}
		if summary.LimitBytes == 0 && st.limit > 0 {
			summary.LimitBytes = st.limit
		}
	}
	return summary
}

type containerStat struct {
	usage   uint64
	limit   uint64
	percent string
}

func loadStatsByContainerName(composeFile string) (map[string]containerStat, error) {
	ids, err := composeContainerIDs(composeFile)
	if err != nil || len(ids) == 0 {
		return nil, err
	}

	args := append([]string{"stats", "--no-stream", "--format", "{{.Name}}\t{{.MemUsage}}\t{{.MemPerc}}"}, ids...)
	cmd := exec.Command("docker", args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("docker stats: %w", err)
	}

	result := make(map[string]containerStat)
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}
		usage, limit, err := parseMemUsagePair(parts[1])
		if err != nil {
			continue
		}
		percent := ""
		if len(parts) >= 3 {
			percent = strings.TrimSpace(parts[2])
		}
		result[parts[0]] = containerStat{usage: usage, limit: limit, percent: percent}
	}
	return result, nil
}

func composeContainerIDs(composeFile string) ([]string, error) {
	dir := composeDir(composeFile)
	cmd := exec.Command("docker", "compose", "-f", filepath.Base(composeFile), "ps", "-q")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var ids []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			ids = append(ids, line)
		}
	}
	return ids, nil
}

func parseMemUsagePair(raw string) (usage, limit uint64, err error) {
	parts := strings.Split(raw, " / ")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid mem usage: %q", raw)
	}
	usage, err = parseDockerSize(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, err
	}
	limit, err = parseDockerSize(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, err
	}
	return usage, limit, nil
}

func parseDockerSize(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	if s == "" || s == "--" {
		return 0, nil
	}
	i := len(s) - 1
	for i >= 0 && (s[i] == 'B' || s[i] == 'b' || s[i] == 'i' || (s[i] >= 'A' && s[i] <= 'Z') || (s[i] >= 'a' && s[i] <= 'z')) {
		i--
	}
	if i < 0 {
		return 0, fmt.Errorf("invalid size: %q", s)
	}
	numStr := strings.TrimSpace(s[:i+1])
	unit := strings.TrimSpace(s[i+1:])
	value, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return 0, err
	}
	mult := sizeMultiplier(unit)
	return uint64(value * mult), nil
}

func sizeMultiplier(unit string) float64 {
	switch strings.ToUpper(unit) {
	case "B":
		return 1
	case "KIB", "KB":
		return 1024
	case "MIB", "MB":
		return 1024 * 1024
	case "GIB", "GB":
		return 1024 * 1024 * 1024
	case "TIB", "TB":
		return 1024 * 1024 * 1024 * 1024
	default:
		return 1
	}
}

// FormatBytes renders a byte count for display (IEC units).
func FormatBytes(n uint64) string {
	switch {
	case n >= 1024*1024*1024:
		return fmt.Sprintf("%.1f GiB", float64(n)/(1024*1024*1024))
	case n >= 1024*1024:
		return fmt.Sprintf("%.0f MiB", float64(n)/(1024*1024))
	case n >= 1024:
		return fmt.Sprintf("%.0f KiB", float64(n)/1024)
	default:
		return fmt.Sprintf("%d B", n)
	}
}
