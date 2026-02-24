package publisher

import (
	"github.com/manzanita-research/chaparral/internal/config"
)

// WrittenFile describes a file that was written to disk.
type WrittenFile struct {
	Path  string // relative to brand repo root
	IsNew bool   // true if file didn't exist before
}

// FileChange describes a change that would be made to a file.
type FileChange struct {
	Path       string // relative to brand repo root
	Kind       string // "new", "modified", "unchanged"
	OldContent string
	NewContent string
}

// FreshnessResult describes whether a skill's published manifest is stale.
type FreshnessResult struct {
	Skill            string
	Stale            bool
	PublishedVersion string // empty if not yet published
}

// bumpVersion reads the existing plugin.json for a skill and returns the next
// patch version. Returns "0.1.0" if the file is missing or unparseable.
func bumpVersion(brandRepoPath, skillsDir, skillName string) string {
	panic("not implemented")
}

// WriteManifests generates and writes plugin.json for each skill and
// marketplace.json for the org. Returns the list of written files.
func WriteManifests(org config.Org, skills []config.Skill) ([]WrittenFile, error) {
	panic("not implemented")
}

// DiffManifests produces a read-only diff of what WriteManifests would do.
func DiffManifests(org config.Org, skills []config.Skill) ([]FileChange, error) {
	panic("not implemented")
}

// CheckFreshness reports whether each skill's source files are newer than
// its published plugin.json.
func CheckFreshness(org config.Org, skills []config.Skill) ([]FreshnessResult, error) {
	panic("not implemented")
}
