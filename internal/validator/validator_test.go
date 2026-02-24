package validator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/manzanita-research/chaparral/internal/config"
)

func makeSkill(t *testing.T, dir, name, content string, writeFile bool) config.Skill {
	t.Helper()
	skillDir := filepath.Join(dir, name)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}
	if writeFile {
		if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	return config.Skill{Name: name, Path: skillDir}
}

func TestValidateSkill_Valid(t *testing.T) {
	dir := t.TempDir()
	skill := makeSkill(t, dir, "brand-voice", "---\nname: brand-voice\ndescription: Write in the Manzanita Research voice\nlicense: MIT\n---\n", true)

	result := ValidateSkill(skill)
	if !result.IsValid() {
		t.Errorf("expected valid, got errors: %v", result.Errors)
	}
	if len(result.Warnings) > 0 {
		t.Errorf("expected no warnings, got: %v", result.Warnings)
	}
}

func TestValidateSkill_MissingName(t *testing.T) {
	dir := t.TempDir()
	skill := makeSkill(t, dir, "bad-skill", "---\nname:\ndescription: some description\n---\n", true)

	result := ValidateSkill(skill)
	if result.IsValid() {
		t.Error("expected errors for missing name")
	}
	assertHasError(t, result, "missing required field: name")
}

func TestValidateSkill_MissingDescription(t *testing.T) {
	dir := t.TempDir()
	skill := makeSkill(t, dir, "bad-skill", "---\nname: bad-skill\ndescription:\n---\n", true)

	result := ValidateSkill(skill)
	if result.IsValid() {
		t.Error("expected errors for missing description")
	}
	assertHasError(t, result, "missing required field: description")
}

func TestValidateSkill_InvalidName_Spaces(t *testing.T) {
	dir := t.TempDir()
	skill := makeSkill(t, dir, "brand voice", "---\nname: brand voice\ndescription: some description\n---\n", true)

	result := ValidateSkill(skill)
	if result.IsValid() {
		t.Error("expected errors for name with spaces")
	}
}

func TestValidateSkill_InvalidName_Uppercase(t *testing.T) {
	dir := t.TempDir()
	skill := makeSkill(t, dir, "BrandVoice", "---\nname: BrandVoice\ndescription: some description\n---\n", true)

	result := ValidateSkill(skill)
	if result.IsValid() {
		t.Error("expected errors for uppercase name")
	}
}

func TestValidateSkill_NameMismatch(t *testing.T) {
	dir := t.TempDir()
	skill := makeSkill(t, dir, "brand-voice", "---\nname: different-name\ndescription: some description\n---\n", true)

	result := ValidateSkill(skill)
	if !result.IsValid() {
		t.Errorf("name mismatch should be warning not error, got errors: %v", result.Errors)
	}
	assertHasWarning(t, result, `frontmatter name "different-name" differs from directory name "brand-voice"`)
}

func TestValidateSkill_NoLicense(t *testing.T) {
	dir := t.TempDir()
	skill := makeSkill(t, dir, "brand-voice", "---\nname: brand-voice\ndescription: some description\n---\n", true)

	result := ValidateSkill(skill)
	if !result.IsValid() {
		t.Errorf("missing license should be warning not error, got errors: %v", result.Errors)
	}
	assertHasWarning(t, result, "no license specified")
}

func TestValidateSkill_MissingSKILLmd(t *testing.T) {
	dir := t.TempDir()
	skill := makeSkill(t, dir, "empty-skill", "", false)

	result := ValidateSkill(skill)
	if result.IsValid() {
		t.Error("expected error for missing SKILL.md")
	}
}

func assertHasError(t *testing.T, result ValidationResult, msg string) {
	t.Helper()
	for _, e := range result.Errors {
		if e == msg {
			return
		}
	}
	t.Errorf("expected error %q in %v", msg, result.Errors)
}

func assertHasWarning(t *testing.T, result ValidationResult, msg string) {
	t.Helper()
	for _, w := range result.Warnings {
		if w == msg {
			return
		}
	}
	t.Errorf("expected warning %q in %v", msg, result.Warnings)
}
