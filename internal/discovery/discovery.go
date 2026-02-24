package discovery

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/manzanita-research/chaparral/internal/config"
)

const manifestFile = "chaparral.json"

// FindOrgs scans a base directory for organization directories.
// An org is any directory containing a repo with a chaparral.json.
func FindOrgs(basePath string) ([]config.Org, error) {
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return nil, err
	}

	var orgs []config.Org
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		orgPath := filepath.Join(basePath, entry.Name())
		org, found, err := scanOrgDir(orgPath)
		if err != nil {
			continue // skip dirs we can't read
		}
		if found {
			orgs = append(orgs, org)
		}
	}

	return orgs, nil
}

// scanOrgDir looks inside an org directory for a repo containing chaparral.json.
func scanOrgDir(orgPath string) (config.Org, bool, error) {
	entries, err := os.ReadDir(orgPath)
	if err != nil {
		return config.Org{}, false, err
	}

	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		manifestPath := filepath.Join(orgPath, entry.Name(), manifestFile)
		manifest, err := config.LoadManifest(manifestPath)
		if err != nil {
			continue
		}

		repos := discoverRepos(orgPath, entries, entry.Name(), manifest.Exclude)

		org := config.Org{
			Name:      manifest.Org,
			Path:      orgPath,
			BrandRepo: entry.Name(),
			Manifest:  manifest,
			Repos:     repos,
		}
		return org, true, nil
	}

	return config.Org{}, false, nil
}

// discoverRepos finds sibling repos, excluding the brand repo and excluded names.
func discoverRepos(orgPath string, entries []os.DirEntry, brandRepo string, exclude []string) []string {
	excluded := make(map[string]bool)
	excluded[brandRepo] = true
	for _, ex := range exclude {
		excluded[ex] = true
	}

	var repos []string
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if excluded[entry.Name()] {
			continue
		}
		// Check it looks like a repo (has .git or any files)
		repoPath := filepath.Join(orgPath, entry.Name())
		if isRepo(repoPath) {
			repos = append(repos, entry.Name())
		}
	}
	return repos
}

// FindSkills returns all skills in a skills directory.
func FindSkills(skillsDir string) ([]config.Skill, error) {
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil, err
	}

	var skills []config.Skill
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		// Verify it has a SKILL.md
		skillMD := filepath.Join(skillsDir, entry.Name(), "SKILL.md")
		if _, err := os.Stat(skillMD); err == nil {
			skills = append(skills, config.Skill{
				Name: entry.Name(),
				Path: filepath.Join(skillsDir, entry.Name()),
			})
		}
	}
	return skills, nil
}

func isRepo(path string) bool {
	gitDir := filepath.Join(path, ".git")
	if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
		return true
	}
	return false
}
