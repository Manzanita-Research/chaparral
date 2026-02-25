package marketplace

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// InstalledPlugin represents a plugin installed via Claude Code.
type InstalledPlugin struct {
	PluginID    string // e.g. "superpowers@superpowers-marketplace"
	Name        string // e.g. "superpowers"
	Marketplace string // e.g. "superpowers-marketplace"
	Version     string
	Scope       string // "user" or "project"
	Enabled     bool
	ProjectPath string // only set for project-scoped plugins
	InstallPath string
}

// AvailablePlugin represents a plugin available from a marketplace.
type AvailablePlugin struct {
	PluginID    string `json:"pluginId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Marketplace string `json:"marketplaceName"`
	Version     string `json:"version"`
}

// PluginStatus combines installed state with available info for display.
type PluginStatus struct {
	PluginID         string
	Name             string
	Marketplace      string
	Version          string // installed version (empty if not installed)
	AvailableVersion string // marketplace version (empty if not in marketplace)
	Scope            string
	Enabled          bool
	Installed        bool
	Available        bool
	Description      string
}

// ParsePluginID splits "name@marketplace" into its parts.
func ParsePluginID(id string) (name, marketplace string) {
	if id == "" {
		return "", ""
	}
	parts := strings.SplitN(id, "@", 2)
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[1]
}

// availableListResponse is the JSON structure from `claude plugin list --available --json`.
type availableListResponse struct {
	Installed []json.RawMessage `json:"installed"`
	Available []AvailablePlugin `json:"available"`
}

// QueryAvailable runs `claude plugin list --available --json` and returns available plugins.
func QueryAvailable() ([]AvailablePlugin, error) {
	cmd := exec.Command("claude", "plugin", "list", "--available", "--json")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("querying marketplace: %w", err)
	}
	if len(out) == 0 {
		return nil, nil
	}

	var resp availableListResponse
	if err := json.Unmarshal(out, &resp); err != nil {
		return nil, fmt.Errorf("parsing marketplace response: %w", err)
	}
	return resp.Available, nil
}

// MergeStatus combines installed plugins (filtered for a repo) with available plugins
// into a unified status list for display.
func MergeStatus(installed []InstalledPlugin, available []AvailablePlugin, repoPath string) []PluginStatus {
	// Filter installed for this repo
	repoPlugins := PluginsForRepo(installed, repoPath)

	// Build map of installed by plugin ID
	installedMap := make(map[string]InstalledPlugin)
	for _, p := range repoPlugins {
		installedMap[p.PluginID] = p
	}

	// Build map of available by plugin ID
	availableMap := make(map[string]AvailablePlugin)
	for _, a := range available {
		availableMap[a.PluginID] = a
	}

	// Merge: start with installed
	seen := make(map[string]bool)
	var statuses []PluginStatus

	for _, p := range repoPlugins {
		seen[p.PluginID] = true
		s := PluginStatus{
			PluginID:  p.PluginID,
			Name:      p.Name,
			Marketplace: p.Marketplace,
			Version:   p.Version,
			Scope:     p.Scope,
			Enabled:   p.Enabled,
			Installed: true,
		}
		if a, ok := availableMap[p.PluginID]; ok {
			s.Available = true
			s.AvailableVersion = a.Version
			s.Description = a.Description
		}
		statuses = append(statuses, s)
	}

	// Add available-only plugins
	for _, a := range available {
		if seen[a.PluginID] {
			continue
		}
		statuses = append(statuses, PluginStatus{
			PluginID:         a.PluginID,
			Name:             a.Name,
			Marketplace:      a.Marketplace,
			AvailableVersion: a.Version,
			Available:        true,
			Description:      a.Description,
		})
	}

	return statuses
}
