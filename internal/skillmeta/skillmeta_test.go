package skillmeta

import (
	"os"
	"path/filepath"
	"testing"
)

func writeSkillMD(t *testing.T, dir string, content string) string {
	t.Helper()
	path := filepath.Join(dir, "SKILL.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestParseFrontmatter_ValidComplete(t *testing.T) {
	dir := t.TempDir()
	path := writeSkillMD(t, dir, `---
name: brand-voice
description: Write in the Manzanita Research voice
license: Complete terms in LICENSE.txt
---

Body content here.
`)

	fm, err := ParseFrontmatter(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm.Name != "brand-voice" {
		t.Errorf("name = %q, want %q", fm.Name, "brand-voice")
	}
	if fm.Description != "Write in the Manzanita Research voice" {
		t.Errorf("description = %q, want %q", fm.Description, "Write in the Manzanita Research voice")
	}
	if fm.License != "Complete terms in LICENSE.txt" {
		t.Errorf("license = %q, want %q", fm.License, "Complete terms in LICENSE.txt")
	}
}

func TestParseFrontmatter_MinimalValid(t *testing.T) {
	dir := t.TempDir()
	path := writeSkillMD(t, dir, `---
name: tui-design
description: Build terminal user interfaces
---
`)

	fm, err := ParseFrontmatter(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm.Name != "tui-design" {
		t.Errorf("name = %q, want %q", fm.Name, "tui-design")
	}
	if fm.Description != "Build terminal user interfaces" {
		t.Errorf("description = %q, want %q", fm.Description, "Build terminal user interfaces")
	}
	if fm.License != "" {
		t.Errorf("license = %q, want empty", fm.License)
	}
}

func TestParseFrontmatter_MissingDelimiter(t *testing.T) {
	dir := t.TempDir()
	path := writeSkillMD(t, dir, `name: brand-voice
description: some text
`)

	_, err := ParseFrontmatter(path)
	if err == nil {
		t.Fatal("expected error for missing frontmatter delimiter")
	}
}

func TestParseFrontmatter_NoClosingDelimiter(t *testing.T) {
	dir := t.TempDir()
	path := writeSkillMD(t, dir, `---
name: brand-voice
description: Write in the Manzanita Research voice
`)

	fm, err := ParseFrontmatter(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm.Name != "brand-voice" {
		t.Errorf("name = %q, want %q", fm.Name, "brand-voice")
	}
}

func TestParseFrontmatter_EmptyValues(t *testing.T) {
	dir := t.TempDir()
	path := writeSkillMD(t, dir, `---
name:
description: some text
---
`)

	fm, err := ParseFrontmatter(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm.Name != "" {
		t.Errorf("name = %q, want empty", fm.Name)
	}
}

func TestParseFrontmatter_DescriptionWithColons(t *testing.T) {
	dir := t.TempDir()
	path := writeSkillMD(t, dir, `---
name: review-code
description: Review code for bugs, security, and performance: be thorough
---
`)

	fm, err := ParseFrontmatter(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm.Description != "Review code for bugs, security, and performance: be thorough" {
		t.Errorf("description = %q, want %q", fm.Description, "Review code for bugs, security, and performance: be thorough")
	}
}
