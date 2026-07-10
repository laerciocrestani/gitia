package ai

import (
	"path/filepath"
	"strings"
)

// ChangeArea representa um agrupamento semântico de arquivos alterados.
type ChangeArea struct {
	Key  string
	Path string
}

// ChangeAreasFromStat agrupa arquivos do git diff --stat por área de mudança.
func ChangeAreasFromStat(stat string) []ChangeArea {
	seen := make(map[string]bool)
	var areas []ChangeArea

	for _, line := range strings.Split(stat, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, " ") {
			continue
		}
		pipe := strings.Index(line, "|")
		if pipe < 0 {
			continue
		}
		path := strings.TrimSpace(line[:pipe])
		if path == "" {
			continue
		}

		key := changeAreaKey(path)
		if seen[key] {
			continue
		}
		seen[key] = true
		areas = append(areas, ChangeArea{Key: key, Path: path})
	}

	return areas
}

func changeAreaKey(path string) string {
	dir := filepath.ToSlash(filepath.Dir(path))
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	lowerDir := strings.ToLower(dir)

	switch {
	case strings.Contains(lowerDir, "/commands"):
		return dir + "/" + base
	case strings.Contains(lowerDir, "/controllers"):
		return dir + "/" + base
	case strings.Contains(lowerDir, "/models"):
		return dir + "/" + base
	case strings.Contains(lowerDir, "/services"):
		return dir + "/" + base
	case strings.Contains(lowerDir, "/handlers"):
		return dir + "/" + base
	default:
		return dir
	}
}

// ShouldSuggestSplit indica se as alterações parecem abranger áreas distintas.
func ShouldSuggestSplit(areas []ChangeArea) bool {
	return len(areas) >= 2
}

// FormatSplitSuggestion gera mensagem orientando commits atômicos.
func FormatSplitSuggestion(areas []ChangeArea) string {
	if len(areas) < 2 {
		return ""
	}

	var names []string
	for _, area := range areas {
		names = append(names, area.Key)
	}

	return "Alterações em " + strings.Join(names, ", ") +
		" — considere commits separados (git add -p ou git add <paths>)."
}
