package generator

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/manzanita-research/chaparral/internal/config"
	"github.com/manzanita-research/chaparral/internal/skillmeta"
)

const initialVersion = "0.1.0"

// PluginManifest is the Claude Code plugin.json format for a single skill.
type PluginManifest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version"`
	License     string `json:"license,omitempty"`
	Skills      string `json:"skills,omitempty"`
}

// MarketplaceManifest is the Claude Code marketplace.json format.
type MarketplaceManifest struct {
	Name    string              `json:"name"`
	Owner   MarketplaceOwner    `json:"owner"`
	Plugins []MarketplacePlugin `json:"plugins"`
}

// MarketplaceOwner identifies the marketplace maintainer.
type MarketplaceOwner struct {
	Name string `json:"name"`
}

// MarketplacePlugin is a single plugin entry in the marketplace catalog.
type MarketplacePlugin struct {
	Name        string `json:"name"`
	Source      string `json:"source"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version"`
}

// GeneratePlugin creates a PluginManifest from a skill's SKILL.md frontmatter.
func GeneratePlugin(skill config.Skill) (PluginManifest, error) {
	skillMDPath := filepath.Join(skill.Path, "SKILL.md")
	fm, err := skillmeta.ParseFrontmatter(skillMDPath)
	if err != nil {
		return PluginManifest{}, fmt.Errorf("reading %s: %w", skill.Name, err)
	}

	return PluginManifest{
		Name:        fm.Name,
		Description: fm.Description,
		Version:     initialVersion,
		License:     fm.License,
		Skills:      "./",
	}, nil
}

// GeneratePluginJSON returns the plugin manifest as pretty-printed JSON.
func GeneratePluginJSON(skill config.Skill) ([]byte, error) {
	manifest, err := GeneratePlugin(skill)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(manifest, "", "  ")
}

// GenerateMarketplace creates a MarketplaceManifest from an org and its skills.
func GenerateMarketplace(org config.Org, skills []config.Skill) (MarketplaceManifest, error) {
	var plugins []MarketplacePlugin

	for _, skill := range skills {
		skillMDPath := filepath.Join(skill.Path, "SKILL.md")
		fm, err := skillmeta.ParseFrontmatter(skillMDPath)
		if err != nil {
			return MarketplaceManifest{}, fmt.Errorf("reading %s: %w", skill.Name, err)
		}

		source := "./" + filepath.ToSlash(filepath.Join(org.Manifest.SkillsDir, skill.Name))

		plugins = append(plugins, MarketplacePlugin{
			Name:        fm.Name,
			Source:      source,
			Description: fm.Description,
			Version:     initialVersion,
		})
	}

	return MarketplaceManifest{
		Name:    org.Manifest.Org,
		Owner:   MarketplaceOwner{Name: org.Manifest.Org},
		Plugins: plugins,
	}, nil
}

// GenerateMarketplaceJSON returns the marketplace manifest as pretty-printed JSON.
func GenerateMarketplaceJSON(org config.Org, skills []config.Skill) ([]byte, error) {
	marketplace, err := GenerateMarketplace(org, skills)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(marketplace, "", "  ")
}
