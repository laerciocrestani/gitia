package desktop

import (
	"github.com/laerciocrestani/openbench/internal/aiskills"
)

// SkillView is the UI/API representation of a chat skill.
type SkillView struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Body        string `json:"body"`
	Enabled     bool   `json:"enabled"`
	Builtin     bool   `json:"builtin"`
	Customized  bool   `json:"customized"`
}

// SkillsListView is returned by ListSkills.
type SkillsListView struct {
	Skills []SkillView `json:"skills"`
	Dir    string      `json:"dir"`
}

func skillToView(s aiskills.Skill) SkillView {
	return SkillView{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
		Body:        s.Body,
		Enabled:     s.Enabled,
		Builtin:     s.Builtin,
		Customized:  s.Customized,
	}
}

// ListSkills returns all skills for the settings UI.
func ListSkills() (*SkillsListView, error) {
	list, err := aiskills.List()
	if err != nil {
		return nil, err
	}
	dir, _ := aiskills.SkillsDirString()
	out := make([]SkillView, 0, len(list))
	for _, s := range list {
		out = append(out, skillToView(s))
	}
	return &SkillsListView{Skills: out, Dir: dir}, nil
}

// GetSkill returns one skill by id.
func GetSkill(id string) (*SkillView, error) {
	s, err := aiskills.Get(id)
	if err != nil {
		return nil, err
	}
	v := skillToView(*s)
	return &v, nil
}

// SaveSkill creates or updates a skill (user override for builtins).
func SaveSkill(id, name, description, body string) error {
	return aiskills.Save(id, name, description, body)
}

// SetSkillEnabled enables or disables a skill.
func SetSkillEnabled(id string, enabled bool) error {
	return aiskills.SetEnabled(id, enabled)
}

// ResetSkill restores a builtin skill body and re-enables it.
func ResetSkill(id string) error {
	return aiskills.Reset(id)
}

// DeleteSkill removes a custom skill.
func DeleteSkill(id string) error {
	return aiskills.Delete(id)
}
