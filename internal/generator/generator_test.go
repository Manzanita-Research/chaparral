package generator

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/manzanita-research/chaparral/internal/config"
)

func makeSkill(t *testing.T, dir, name, content string) config.Skill {
	t.Helper()
	skillDir := filepath.Join(dir, name)
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return config.Skill{Name: name, Path: skillDir}
}

func TestGeneratePlugin_Complete(t *testing.T) {
	dir := t.TempDir()
	skill := makeSkill(t, dir, "brand-voice", "---\nname: brand-voice\ndescription: Write in the Manzanita Research voice\nlicense: MIT\n---\n")

	manifest, err := GeneratePlugin(skill)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if manifest.Name != "brand-voice" {
		t.Errorf("name = %q, want %q", manifest.Name, "brand-voice")
	}
	if manifest.Description != "Write in the Manzanita Research voice" {
		t.Errorf("description = %q", manifest.Description)
	}
	if manifest.Version != "0.1.0" {
		t.Errorf("version = %q, want %q", manifest.Version, "0.1.0")
	}
	if manifest.License != "MIT" {
		t.Errorf("license = %q, want %q", manifest.License, "MIT")
	}
	if manifest.Skills != "./" {
		t.Errorf("skills = %q, want %q", manifest.Skills, "./")
	}
}

func TestGeneratePlugin_MinimalFrontmatter(t *testing.T) {
	dir := t.TempDir()
	skill := makeSkill(t, dir, "tui-design", "---\nname: tui-design\ndescription: Build terminal user interfaces\n---\n")

	manifest, err := GeneratePlugin(skill)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if manifest.Name != "tui-design" {
		t.Errorf("name = %q", manifest.Name)
	}
	if manifest.License != "" {
		t.Errorf("license = %q, want empty", manifest.License)
	}
}

func TestGeneratePlugin_JSONOutput(t *testing.T) {
	dir := t.TempDir()
	skill := makeSkill(t, dir, "brand-voice", "---\nname: brand-voice\ndescription: Write in the voice\nlicense: MIT\n---\n")

	data, err := GeneratePluginJSON(skill)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Must be valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, data)
	}

	// Must be pretty-printed (contains newlines and indentation)
	s := string(data)
	if len(s) < 10 {
		t.Error("JSON too short, expected pretty-printed output")
	}
}

func TestGeneratePlugin_VersionAlways0_1_0(t *testing.T) {
	dir := t.TempDir()
	skill := makeSkill(t, dir, "test-skill", "---\nname: test-skill\ndescription: test\n---\n")

	manifest, err := GeneratePlugin(skill)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if manifest.Version != "0.1.0" {
		t.Errorf("version = %q, want %q", manifest.Version, "0.1.0")
	}
}

func TestGenerateMarketplace_SingleSkill(t *testing.T) {
	dir := t.TempDir()
	skills := []config.Skill{
		makeSkill(t, dir, "brand-voice", "---\nname: brand-voice\ndescription: Write in the voice\n---\n"),
	}
	org := config.Org{
		Name: "manzanita-research",
		Manifest: config.Manifest{
			Org:       "manzanita-research",
			SkillsDir: "org/skills",
		},
	}

	marketplace, err := GenerateMarketplace(org, skills)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(marketplace.Plugins) != 1 {
		t.Fatalf("plugins count = %d, want 1", len(marketplace.Plugins))
	}
	if marketplace.Plugins[0].Name != "brand-voice" {
		t.Errorf("plugin name = %q", marketplace.Plugins[0].Name)
	}
}

func TestGenerateMarketplace_MultipleSkills(t *testing.T) {
	dir := t.TempDir()
	skills := []config.Skill{
		makeSkill(t, dir, "brand-voice", "---\nname: brand-voice\ndescription: Voice\n---\n"),
		makeSkill(t, dir, "tui-design", "---\nname: tui-design\ndescription: TUI\n---\n"),
		makeSkill(t, dir, "pitch-md", "---\nname: pitch-md\ndescription: Pitch\n---\n"),
	}
	org := config.Org{
		Name:     "manzanita-research",
		Manifest: config.Manifest{Org: "manzanita-research", SkillsDir: "org/skills"},
	}

	marketplace, err := GenerateMarketplace(org, skills)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(marketplace.Plugins) != 3 {
		t.Fatalf("plugins count = %d, want 3", len(marketplace.Plugins))
	}
}

func TestGenerateMarketplace_NameFromOrg(t *testing.T) {
	dir := t.TempDir()
	skills := []config.Skill{
		makeSkill(t, dir, "test", "---\nname: test\ndescription: test\n---\n"),
	}
	org := config.Org{
		Name:     "manzanita-research",
		Manifest: config.Manifest{Org: "manzanita-research", SkillsDir: "org/skills"},
	}

	marketplace, err := GenerateMarketplace(org, skills)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if marketplace.Name != "manzanita-research" {
		t.Errorf("marketplace name = %q, want %q", marketplace.Name, "manzanita-research")
	}
}

func TestGenerateMarketplace_OwnerFromOrg(t *testing.T) {
	dir := t.TempDir()
	skills := []config.Skill{
		makeSkill(t, dir, "test", "---\nname: test\ndescription: test\n---\n"),
	}
	org := config.Org{
		Name:     "manzanita-research",
		Manifest: config.Manifest{Org: "manzanita-research", SkillsDir: "org/skills"},
	}

	marketplace, err := GenerateMarketplace(org, skills)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if marketplace.Owner.Name != "manzanita-research" {
		t.Errorf("owner name = %q, want %q", marketplace.Owner.Name, "manzanita-research")
	}
}

func TestGenerateMarketplace_PluginSourcePaths(t *testing.T) {
	dir := t.TempDir()
	skills := []config.Skill{
		makeSkill(t, dir, "brand-voice", "---\nname: brand-voice\ndescription: Voice\n---\n"),
	}
	org := config.Org{
		Name:     "manzanita-research",
		Manifest: config.Manifest{Org: "manzanita-research", SkillsDir: "org/skills"},
	}

	marketplace, err := GenerateMarketplace(org, skills)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "./org/skills/brand-voice"
	if marketplace.Plugins[0].Source != expected {
		t.Errorf("source = %q, want %q", marketplace.Plugins[0].Source, expected)
	}
}

func TestGenerateMarketplace_JSONOutput(t *testing.T) {
	dir := t.TempDir()
	skills := []config.Skill{
		makeSkill(t, dir, "test", "---\nname: test\ndescription: test\n---\n"),
	}
	org := config.Org{
		Name:     "test-org",
		Manifest: config.Manifest{Org: "test-org", SkillsDir: "skills"},
	}

	data, err := GenerateMarketplaceJSON(org, skills)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, data)
	}
}
