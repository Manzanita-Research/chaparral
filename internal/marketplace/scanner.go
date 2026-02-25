package marketplace

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// pluginsFilePath is the default path to installed_plugins.json.
// Overridable in tests.
var pluginsFilePath = ""

func init() {
	home, err := os.UserHomeDir()
	if err == nil {
		pluginsFilePath = filepath.Join(home, ".claude", "plugins", "installed_plugins.json")
	}
}

// InstalledPluginsFile represents the ~/.claude/plugins/installed_plugins.json format.
type InstalledPluginsFile struct {
	Version int                        `json:"version"`
	Plugins map[string][]PluginInstall `json:"plugins"`
}

// PluginInstall represents a single install entry in installed_plugins.json.
type PluginInstall struct {
	Scope        string `json:"scope"`
	InstallPath  string `json:"installPath"`
	Version      string `json:"version"`
	InstalledAt  string `json:"installedAt"`
	LastUpdated  string `json:"lastUpdated"`
	ProjectPath  string `json:"projectPath,omitempty"`
	GitCommitSha string `json:"gitCommitSha,omitempty"`
}

// ReadInstalledPlugins reads and parses installed_plugins.json.
// Returns an empty file (not error) if the file doesn't exist.
func ReadInstalledPlugins() (*InstalledPluginsFile, error) {
	data, err := os.ReadFile(pluginsFilePath)
	if os.IsNotExist(err) {
		return &InstalledPluginsFile{
			Version: 2,
			Plugins: map[string][]PluginInstall{},
		}, nil
	}
	if err != nil {
		return nil, err
	}

	var f InstalledPluginsFile
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	if f.Plugins == nil {
		f.Plugins = map[string][]PluginInstall{}
	}
	return &f, nil
}

// ScanInstalled reads installed_plugins.json and returns a flat list of installed plugins.
func ScanInstalled() ([]InstalledPlugin, error) {
	f, err := ReadInstalledPlugins()
	if err != nil {
		return nil, err
	}

	var plugins []InstalledPlugin
	for pluginID, installs := range f.Plugins {
		name, marketplace := ParsePluginID(pluginID)
		for _, install := range installs {
			plugins = append(plugins, InstalledPlugin{
				PluginID:    pluginID,
				Name:        name,
				Marketplace: marketplace,
				Version:     install.Version,
				Scope:       install.Scope,
				ProjectPath: install.ProjectPath,
				InstallPath: install.InstallPath,
			})
		}
	}

	return plugins, nil
}

// PluginsForRepo filters installed plugins to those that apply to a given repo path.
// User-scoped plugins apply to all repos. Project-scoped plugins match on projectPath.
func PluginsForRepo(plugins []InstalledPlugin, repoPath string) []InstalledPlugin {
	cleanRepo := cleanPath(repoPath)

	var result []InstalledPlugin
	for _, p := range plugins {
		switch p.Scope {
		case "user":
			result = append(result, p)
		case "project":
			cleanProject := cleanPath(p.ProjectPath)
			if cleanProject == cleanRepo {
				result = append(result, p)
			}
		}
	}
	return result
}

// cleanPath normalizes a path for comparison by resolving symlinks and cleaning.
func cleanPath(p string) string {
	if p == "" {
		return ""
	}
	resolved, err := filepath.EvalSymlinks(p)
	if err != nil {
		return filepath.Clean(p)
	}
	return filepath.Clean(resolved)
}
