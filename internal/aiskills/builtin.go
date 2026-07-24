package aiskills

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
)

//go:embed builtin/*.skill.md
var builtinFS embed.FS

func loadBuiltinSkills() ([]Skill, error) {
	entries, err := fs.ReadDir(builtinFS, "builtin")
	if err != nil {
		return nil, err
	}
	var out []Skill
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".skill.md") {
			continue
		}
		rel := path.Join("builtin", e.Name())
		data, err := builtinFS.ReadFile(rel)
		if err != nil {
			return nil, fmt.Errorf("ler builtin %s: %w", e.Name(), err)
		}
		s, err := ParseSkillMarkdown(string(data))
		if err != nil {
			return nil, fmt.Errorf("parse builtin %s: %w", e.Name(), err)
		}
		s.Builtin = true
		s.Customized = false
		s.Enabled = true
		out = append(out, *s)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func builtinByID(id string) (*Skill, error) {
	all, err := loadBuiltinSkills()
	if err != nil {
		return nil, err
	}
	for i := range all {
		if all[i].ID == id {
			s := all[i]
			return &s, nil
		}
	}
	return nil, fmt.Errorf("skill builtin %q não encontrada", id)
}
