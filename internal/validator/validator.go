package validator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/manzanita-research/chaparral/internal/config"
	"github.com/manzanita-research/chaparral/internal/discovery"
	"github.com/manzanita-research/chaparral/internal/skillmeta"
)

var kebabCaseRe = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

// ValidationResult holds errors and warnings for a single skill.
type ValidationResult struct {
	Skill    string
	Errors   []string
	Warnings []string
}

// IsValid returns true if no blocking errors were found.
func (r ValidationResult) IsValid() bool {
	return len(r.Errors) == 0
}

// ValidateSkill checks a single skill for structure and metadata issues.
func ValidateSkill(skill config.Skill) ValidationResult {
	result := ValidationResult{Skill: skill.Name}

	// Check SKILL.md exists
	skillMDPath := filepath.Join(skill.Path, "SKILL.md")
	if _, err := os.Stat(skillMDPath); os.IsNotExist(err) {
		result.Errors = append(result.Errors, "missing SKILL.md")
		return result
	}

	// Parse frontmatter
	fm, err := skillmeta.ParseFrontmatter(skillMDPath)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("can't parse frontmatter: %v", err))
		return result
	}

	// Required fields
	if fm.Name == "" {
		result.Errors = append(result.Errors, "missing required field: name")
	}
	if fm.Description == "" {
		result.Errors = append(result.Errors, "missing required field: description")
	}

	// Name format (only check if name is present)
	if fm.Name != "" && !kebabCaseRe.MatchString(fm.Name) {
		result.Errors = append(result.Errors, fmt.Sprintf("name %q must be lowercase kebab-case (e.g., my-skill)", fm.Name))
	}

	// Name mismatch warning
	if fm.Name != "" && fm.Name != skill.Name {
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("frontmatter name %q differs from directory name %q", fm.Name, skill.Name))
	}

	// License warning
	if fm.License == "" {
		result.Warnings = append(result.Warnings, "no license specified")
	}

	return result
}

// ValidateOrg validates all skills in an org.
func ValidateOrg(org config.Org) ([]ValidationResult, error) {
	skills, err := discovery.FindSkills(org.SkillsPath())
	if err != nil {
		return nil, fmt.Errorf("finding skills in %s: %w", org.Name, err)
	}

	var results []ValidationResult
	for _, skill := range skills {
		results = append(results, ValidateSkill(skill))
	}
	return results, nil
}
