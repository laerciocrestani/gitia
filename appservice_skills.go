package main

import "github.com/laerciocrestani/openbench/internal/desktop"

// ListSkills returns builtin + user chat skills.
func (s *AppService) ListSkills() (*desktop.SkillsListView, error) {
	return desktop.ListSkills()
}

// GetSkill returns one skill by id.
func (s *AppService) GetSkill(id string) (*desktop.SkillView, error) {
	return desktop.GetSkill(id)
}

// SaveSkill creates or updates a skill.
func (s *AppService) SaveSkill(id, name, description, body string) error {
	return desktop.SaveSkill(id, name, description, body)
}

// SetSkillEnabled toggles a skill for chat injection.
func (s *AppService) SetSkillEnabled(id string, enabled bool) error {
	return desktop.SetSkillEnabled(id, enabled)
}

// ResetSkill restores a builtin skill to the embedded default.
func (s *AppService) ResetSkill(id string) error {
	return desktop.ResetSkill(id)
}

// DeleteSkill removes a custom (non-builtin) skill.
func (s *AppService) DeleteSkill(id string) error {
	return desktop.DeleteSkill(id)
}
