package marketplace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestParsePluginID(t *testing.T) {
	tests := []struct {
		id         string
		wantName   string
		wantMarket string
	}{
		{"superpowers@superpowers-marketplace", "superpowers", "superpowers-marketplace"},
		{"feature-dev@claude-code-plugins", "feature-dev", "claude-code-plugins"},
		{"tarot@esoterica", "tarot", "esoterica"},
		{"no-at-sign", "no-at-sign", ""},
		{"", "", ""},
	}

	for _, tt := range tests {
		name, market := ParsePluginID(tt.id)
		if name != tt.wantName || market != tt.wantMarket {
			t.Errorf("ParsePluginID(%q) = (%q, %q), want (%q, %q)",
				tt.id, name, market, tt.wantName, tt.wantMarket)
		}
	}
}

func TestReadInstalledPlugins_Missing(t *testing.T) {
	// Point to a non-existent path
	orig := pluginsFilePath
	pluginsFilePath = filepath.Join(t.TempDir(), "nonexistent", "installed_plugins.json")
	defer func() { pluginsFilePath = orig }()

	f, err := ReadInstalledPlugins()
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if f.Version != 2 {
		t.Errorf("expected version 2, got %d", f.Version)
	}
	if len(f.Plugins) != 0 {
		t.Errorf("expected empty plugins map, got %d entries", len(f.Plugins))
	}
}

func TestReadInstalledPlugins_Valid(t *testing.T) {
	dir := t.TempDir()
	data := `{
		"version": 2,
		"plugins": {
			"superpowers@superpowers-marketplace": [
				{
					"scope": "user",
					"installPath": "/Users/test/.claude/plugins/cache/superpowers-marketplace/superpowers/3.4.1",
					"version": "3.4.1",
					"installedAt": "2025-11-13T03:50:45.261Z",
					"lastUpdated": "2025-11-13T03:50:45.261Z"
				}
			],
			"feature-dev@claude-code-plugins": [
				{
					"scope": "project",
					"projectPath": "/Users/test/code/myorg/myrepo",
					"installPath": "/Users/test/.claude/plugins/cache/claude-code-plugins/feature-dev/1.0.0",
					"version": "1.0.0",
					"installedAt": "2025-12-18T18:39:54.710Z",
					"lastUpdated": "2025-12-18T18:39:54.710Z"
				}
			]
		}
	}`
	path := filepath.Join(dir, "installed_plugins.json")
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}

	orig := pluginsFilePath
	pluginsFilePath = path
	defer func() { pluginsFilePath = orig }()

	f, err := ReadInstalledPlugins()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Version != 2 {
		t.Errorf("expected version 2, got %d", f.Version)
	}
	if len(f.Plugins) != 2 {
		t.Errorf("expected 2 plugin entries, got %d", len(f.Plugins))
	}
}

func TestScanInstalled(t *testing.T) {
	dir := t.TempDir()
	data := `{
		"version": 2,
		"plugins": {
			"superpowers@superpowers-marketplace": [
				{
					"scope": "user",
					"installPath": "/tmp/cache/superpowers/3.4.1",
					"version": "3.4.1",
					"installedAt": "2025-11-13T03:50:45.261Z",
					"lastUpdated": "2025-11-13T03:50:45.261Z"
				}
			],
			"feature-dev@claude-code-plugins": [
				{
					"scope": "project",
					"projectPath": "/Users/test/code/myorg/myrepo",
					"installPath": "/tmp/cache/feature-dev/1.0.0",
					"version": "1.0.0",
					"installedAt": "2025-12-18T18:39:54.710Z",
					"lastUpdated": "2025-12-18T18:39:54.710Z"
				}
			]
		}
	}`
	path := filepath.Join(dir, "installed_plugins.json")
	os.WriteFile(path, []byte(data), 0644)

	orig := pluginsFilePath
	pluginsFilePath = path
	defer func() { pluginsFilePath = orig }()

	plugins, err := ScanInstalled()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(plugins) != 2 {
		t.Fatalf("expected 2 plugins, got %d", len(plugins))
	}

	// Check that fields are populated
	found := map[string]bool{}
	for _, p := range plugins {
		found[p.Name] = true
		if p.PluginID == "" || p.Version == "" {
			t.Errorf("plugin %q has empty fields: ID=%q Version=%q", p.Name, p.PluginID, p.Version)
		}
	}
	if !found["superpowers"] || !found["feature-dev"] {
		t.Errorf("expected superpowers and feature-dev, got %v", found)
	}
}

func TestPluginsForRepo_UserScope(t *testing.T) {
	plugins := []InstalledPlugin{
		{PluginID: "sp@mp", Name: "sp", Scope: "user", Version: "1.0"},
	}

	result := PluginsForRepo(plugins, "/some/random/repo")
	if len(result) != 1 {
		t.Errorf("user-scoped plugin should apply to any repo, got %d", len(result))
	}
}

func TestPluginsForRepo_ProjectScope(t *testing.T) {
	plugins := []InstalledPlugin{
		{PluginID: "fd@ccp", Name: "fd", Scope: "project", Version: "1.0", ProjectPath: "/Users/test/code/myorg/myrepo"},
	}

	// Matching path
	result := PluginsForRepo(plugins, "/Users/test/code/myorg/myrepo")
	if len(result) != 1 {
		t.Errorf("project-scoped plugin should match exact path, got %d", len(result))
	}

	// Non-matching path
	result = PluginsForRepo(plugins, "/Users/test/code/myorg/other-repo")
	if len(result) != 0 {
		t.Errorf("project-scoped plugin should not match different path, got %d", len(result))
	}
}

func TestPluginsForRepo_Mixed(t *testing.T) {
	plugins := []InstalledPlugin{
		{PluginID: "sp@mp", Name: "sp", Scope: "user", Version: "1.0"},
		{PluginID: "fd@ccp", Name: "fd", Scope: "project", Version: "1.0", ProjectPath: "/Users/test/code/myorg/myrepo"},
		{PluginID: "other@mp", Name: "other", Scope: "project", Version: "2.0", ProjectPath: "/Users/test/code/myorg/other-repo"},
	}

	result := PluginsForRepo(plugins, "/Users/test/code/myorg/myrepo")
	if len(result) != 2 {
		t.Errorf("expected 2 plugins (1 user + 1 matching project), got %d", len(result))
	}
}

func TestMergeStatus(t *testing.T) {
	installed := []InstalledPlugin{
		{PluginID: "sp@mp", Name: "sp", Marketplace: "mp", Scope: "user", Version: "3.4.1", Enabled: true},
	}
	available := []AvailablePlugin{
		{PluginID: "sp@mp", Name: "sp", Marketplace: "mp", Version: "4.0.0", Description: "Core skills"},
		{PluginID: "chrome@mp", Name: "chrome", Marketplace: "mp", Version: "1.6.1", Description: "Chrome tools"},
	}

	statuses := MergeStatus(installed, available, "/any/repo")
	if len(statuses) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(statuses))
	}

	// Find the installed one
	var spStatus, chromeStatus PluginStatus
	for _, s := range statuses {
		if s.Name == "sp" {
			spStatus = s
		}
		if s.Name == "chrome" {
			chromeStatus = s
		}
	}

	if !spStatus.Installed || !spStatus.Available {
		t.Errorf("sp should be installed and available, got installed=%v available=%v", spStatus.Installed, spStatus.Available)
	}
	if spStatus.Version != "3.4.1" {
		t.Errorf("sp installed version should be 3.4.1, got %s", spStatus.Version)
	}
	if spStatus.AvailableVersion != "4.0.0" {
		t.Errorf("sp available version should be 4.0.0, got %s", spStatus.AvailableVersion)
	}

	if chromeStatus.Installed || !chromeStatus.Available {
		t.Errorf("chrome should be not installed but available, got installed=%v available=%v", chromeStatus.Installed, chromeStatus.Available)
	}
}

// TestInstalledPluginsJSON verifies our types can marshal/unmarshal correctly.
func TestInstalledPluginsJSON(t *testing.T) {
	f := InstalledPluginsFile{
		Version: 2,
		Plugins: map[string][]PluginInstall{
			"test@mp": {
				{Scope: "user", Version: "1.0"},
			},
		},
	}
	data, err := json.Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
	var f2 InstalledPluginsFile
	if err := json.Unmarshal(data, &f2); err != nil {
		t.Fatal(err)
	}
	if f2.Version != 2 || len(f2.Plugins) != 1 {
		t.Errorf("roundtrip failed: %+v", f2)
	}
}
