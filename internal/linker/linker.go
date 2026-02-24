package linker

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/manzanita-research/chaparral/internal/config"
	"github.com/manzanita-research/chaparral/internal/discovery"
)

// LinkResult describes what happened for a single link operation.
type LinkResult struct {
	Repo    string
	Skill   string
	Action  string // "created", "exists", "updated", "skipped", "error"
	Detail  string
}

// SyncOrg links all skills and the org CLAUDE.md for a given org.
func SyncOrg(org config.Org) ([]LinkResult, error) {
	var results []LinkResult

	// Link org-level CLAUDE.md to parent directory
	claudeResults := linkClaudeMD(org)
	results = append(results, claudeResults...)

	// Find available skills
	skills, err := discovery.FindSkills(org.SkillsPath())
	if err != nil {
		return results, fmt.Errorf("finding skills: %w", err)
	}

	// Link skills to each sibling repo
	for _, repo := range org.Repos {
		for _, skill := range skills {
			result := linkSkill(org, repo, skill)
			results = append(results, result)
		}
	}

	return results, nil
}

// UnlinkOrg removes all chaparral-managed symlinks for an org.
func UnlinkOrg(org config.Org) ([]LinkResult, error) {
	var results []LinkResult

	// Unlink org-level CLAUDE.md
	claudeDest := filepath.Join(org.Path, "CLAUDE.md")
	if isOurSymlink(claudeDest) {
		os.Remove(claudeDest)
		results = append(results, LinkResult{
			Repo: "(org)", Skill: "CLAUDE.md", Action: "removed",
		})
	}

	// Find skills to know what to unlink
	skills, err := discovery.FindSkills(org.SkillsPath())
	if err != nil {
		return results, err
	}

	for _, repo := range org.Repos {
		for _, skill := range skills {
			linkPath := filepath.Join(org.Path, repo, ".claude", "skills", skill.Name)
			if isOurSymlink(linkPath) {
				os.Remove(linkPath)
				results = append(results, LinkResult{
					Repo: repo, Skill: skill.Name, Action: "removed",
				})
			}
		}
	}

	return results, nil
}

// Status returns the current link state for an org without changing anything.
type LinkStatus struct {
	Repo      string
	Skill     string
	State     string // "linked", "stale", "missing", "conflict"
	LinkTarget string
}

func StatusOrg(org config.Org) ([]LinkStatus, error) {
	var statuses []LinkStatus

	// Check org CLAUDE.md
	claudeDest := filepath.Join(org.Path, "CLAUDE.md")
	claudeSource := org.ClaudeMDPath()
	statuses = append(statuses, checkLink(claudeDest, claudeSource, "(org)", "CLAUDE.md"))

	// Check skills
	skills, err := discovery.FindSkills(org.SkillsPath())
	if err != nil {
		return statuses, err
	}

	for _, repo := range org.Repos {
		for _, skill := range skills {
			linkPath := filepath.Join(org.Path, repo, ".claude", "skills", skill.Name)
			statuses = append(statuses, checkLink(linkPath, skill.Path, repo, skill.Name))
		}
	}

	return statuses, nil
}

func linkClaudeMD(org config.Org) []LinkResult {
	source := org.ClaudeMDPath()
	dest := filepath.Join(org.Path, "CLAUDE.md")

	if _, err := os.Stat(source); os.IsNotExist(err) {
		return []LinkResult{{
			Repo: "(org)", Skill: "CLAUDE.md", Action: "skipped",
			Detail: "source CLAUDE.md not found",
		}}
	}

	return []LinkResult{createSymlink(source, dest, "(org)", "CLAUDE.md")}
}

func linkSkill(org config.Org, repo string, skill config.Skill) LinkResult {
	skillsDir := filepath.Join(org.Path, repo, ".claude", "skills")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return LinkResult{
			Repo: repo, Skill: skill.Name, Action: "error",
			Detail: fmt.Sprintf("creating .claude/skills: %v", err),
		}
	}

	dest := filepath.Join(skillsDir, skill.Name)
	return createSymlink(skill.Path, dest, repo, skill.Name)
}

func createSymlink(source, dest, repo, name string) LinkResult {
	// Check if destination already exists
	info, err := os.Lstat(dest)
	if err == nil {
		// Something exists at dest
		if info.Mode()&os.ModeSymlink != 0 {
			// It's a symlink — check if it points to our source
			target, err := os.Readlink(dest)
			if err == nil && target == source {
				return LinkResult{Repo: repo, Skill: name, Action: "exists"}
			}
			// Stale symlink — update it
			os.Remove(dest)
		} else {
			// Real file/dir — don't overwrite
			return LinkResult{
				Repo: repo, Skill: name, Action: "skipped",
				Detail: "non-symlink file exists at destination",
			}
		}
	}

	if err := os.Symlink(source, dest); err != nil {
		return LinkResult{
			Repo: repo, Skill: name, Action: "error",
			Detail: err.Error(),
		}
	}

	return LinkResult{Repo: repo, Skill: name, Action: "created"}
}

func checkLink(linkPath, expectedTarget, repo, name string) LinkStatus {
	info, err := os.Lstat(linkPath)
	if os.IsNotExist(err) {
		return LinkStatus{Repo: repo, Skill: name, State: "missing"}
	}
	if err != nil {
		return LinkStatus{Repo: repo, Skill: name, State: "missing"}
	}

	if info.Mode()&os.ModeSymlink == 0 {
		return LinkStatus{Repo: repo, Skill: name, State: "conflict"}
	}

	target, err := os.Readlink(linkPath)
	if err != nil {
		return LinkStatus{Repo: repo, Skill: name, State: "missing"}
	}

	if target == expectedTarget {
		return LinkStatus{Repo: repo, Skill: name, State: "linked", LinkTarget: target}
	}

	return LinkStatus{Repo: repo, Skill: name, State: "stale", LinkTarget: target}
}

func isOurSymlink(path string) bool {
	info, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}
