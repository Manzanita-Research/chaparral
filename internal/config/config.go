package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Manifest represents a chaparral.json file in a brand repo.
type Manifest struct {
	Org       string   `json:"org"`
	ClaudeMD  string   `json:"claude_md"`
	SkillsDir string   `json:"skills_dir"`
	Exclude   []string `json:"exclude"`
}

// Org represents a discovered organization directory.
type Org struct {
	Name      string
	Path      string   // absolute path to org directory
	BrandRepo string   // name of the brand repo within the org
	Manifest  Manifest
	Repos     []string // sibling repo names (excluding brand and excluded)
}

// Skill represents a single skill directory.
type Skill struct {
	Name string
	Path string // absolute path to the skill directory
}

// LoadManifest reads and parses a chaparral.json file.
func LoadManifest(path string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("reading manifest: %w", err)
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return Manifest{}, fmt.Errorf("parsing manifest: %w", err)
	}

	return m, nil
}

// SkillsPath returns the absolute path to the skills directory.
func (o *Org) SkillsPath() string {
	return filepath.Join(o.Path, o.BrandRepo, o.Manifest.SkillsDir)
}

// ClaudeMDPath returns the absolute path to the org-level CLAUDE.md.
func (o *Org) ClaudeMDPath() string {
	return filepath.Join(o.Path, o.BrandRepo, o.Manifest.ClaudeMD)
}

// IsExcluded checks if a repo name should be excluded from linking.
func (o *Org) IsExcluded(repo string) bool {
	for _, ex := range o.Manifest.Exclude {
		if ex == repo {
			return true
		}
	}
	return false
}
