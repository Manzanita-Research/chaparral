# Testing Patterns

**Analysis Date:** 2026-02-24

## Test Framework

**Runner:**
- Standard Go testing framework (`testing` package)
- No third-party test runner configured

**Assertion Library:**
- Standard library only; no assertion library dependency detected

**Run Commands:**
```bash
go test ./...           # Run all tests
go test -v ./...        # Verbose output
go test -cover ./...    # Coverage report
go test -race ./...     # Race detector
```

**Current Status:**
- **No test files present in codebase** (see "Test Coverage" section below)

## Test File Organization

**Expected Location:**
- Test files should be co-located with source code using `_test.go` suffix
- Example patterns:
  - `internal/config/config.go` → `internal/config/config_test.go`
  - `internal/linker/linker.go` → `internal/linker/linker_test.go`
  - `internal/discovery/discovery.go` → `internal/discovery/discovery_test.go`

**Naming Convention:**
- Test functions: `TestFunctionName(t *testing.T)`
- Table-driven tests: `tests := []struct { ... }` with loop over cases
- Subtests: `t.Run("case description", func(t *testing.T) { ... })`

## Test Structure

**Recommended Pattern:**
```go
func TestFunctionName(t *testing.T) {
	tests := []struct {
		name    string
		input   Type
		want    Type
		wantErr bool
	}{
		{
			name:  "basic case",
			input: defaultInput(),
			want:  expectedOutput(),
		},
		{
			name:    "error case",
			input:   badInput(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FunctionUnderTest(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("FunctionUnderTest() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("FunctionUnderTest() = %v, want %v", got, tt.want)
			}
		})
	}
}
```

**Cleanup Pattern:**
```go
func TestWithCleanup(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()

	// Test
	// ... test code ...

	// Cleanup is automatic with t.TempDir()
}
```

## Mocking

**Approach:**
- Go uses interface-based mocking (no mocking library needed)
- Recommended: Use small interfaces for dependencies
- Example pattern for testability:

```go
// In source code, accept interface not concrete type
func ProcessOrg(finder OrgFinder) error {
	// ...
}

// In test, implement minimal interface
type mockOrgFinder struct {
	orgs []config.Org
	err  error
}

func (m *mockOrgFinder) FindOrgs(basePath string) ([]config.Org, error) {
	return m.orgs, m.err
}
```

**What to Mock:**
- Filesystem operations: use `os.TempDir()` / `t.TempDir()` for real temp dirs instead
- External command execution: accept interface over shell execution
- Network operations: accept interface over HTTP client

**What NOT to Mock:**
- Standard library types (strings, io, etc.)
- Small utility functions (date formatting, parsing)
- Data structures that are already simple

## Fixtures and Factories

**Test Data Pattern:**
```go
func newTestOrg() config.Org {
	return config.Org{
		Name:      "test-org",
		Path:      "/tmp/test-org",
		BrandRepo: "brand",
		Manifest:  newTestManifest(),
		Repos:     []string{"repo1", "repo2"},
	}
}

func newTestManifest() config.Manifest {
	return config.Manifest{
		Org:       "test-org",
		ClaudeMD:  "CLAUDE.md",
		SkillsDir: "skills",
		Exclude:   []string{},
	}
}
```

**Location:**
- Place helper functions at end of test file or in separate `_test.go` file
- Group all fixtures together in a helpers section

## Coverage

**Requirements:** Not enforced (no coverage config detected)

**Recommended target:** 70%+ for critical paths
- Critical paths: `linker/`, `discovery/`, `config/`
- Lower priority: `tui/` (UI testing is complex)

**View Coverage:**
```bash
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Test Types

**Unit Tests:**
- Scope: Individual functions in isolation
- Approach: Test pure functions first (config parsing, path operations)
- Example targets:
  - `config.LoadManifest()` with valid/invalid JSON
  - `config.IsExcluded()` with various repo names
  - `discovery.isRepo()` with real temp directories
  - `linker.createSymlink()` with temp filesystem

**Integration Tests:**
- Scope: Multiple packages working together
- Approach: Test full workflows without actual Git/filesystem operations
- Example targets:
  - `SyncOrg()` end-to-end: discover → link → verify
  - `StatusOrg()` with mixed symlink states
  - `UnlinkOrg()` cleanup verification

**E2E Tests:**
- Framework: Not currently used
- Candidates for future E2E:
  - TUI interaction (would require terminal simulation or headless testing)
  - Full CLI commands with real orgs

## Testability Gaps

**Hard to Test (Current Design):**
- `internal/tui/tui.go`: Bubble Tea model logic mixed with view rendering
  - Suggestion: Extract state logic from View() method
  - Create separate functions for state transitions that don't depend on tea.Model

- `cmd/chaparral/main.go`: CLI parsing mixed with logic
  - Suggestion: Create separate functions for each command that accept basePath
  - Example: `func syncCommand(basePath string) error { ... }`

**Opportunities for Improvement:**
1. Accept filesystem operations as interface instead of direct `os.*` calls
2. Move TUI state transition logic into pure functions
3. Extract manifest validation into testable function with specific error types
4. Create explicit error types instead of wrapped fmt.Errorf for error testing

## Recommended Test Writing Sequence

Start with easiest to hardest:

1. **Config tests** (`internal/config/config_test.go`)
   - Test `IsExcluded()` with list of repos
   - Test path calculation methods

2. **Discovery tests** (`internal/discovery/discovery_test.go`)
   - Test file filtering (exclude hidden, non-dirs)
   - Test repo detection (presence of .git)
   - Test skill discovery (SKILL.md validation)

3. **Linker tests** (`internal/linker/linker_test.go`)
   - Test symlink creation logic with temp dirs
   - Test conflict detection
   - Test status checking (existing links, stale links)
   - Test result accumulation in SyncOrg/UnlinkOrg

4. **TUI tests** (`internal/tui/tui_test.go`)
   - Extract and test state transitions
   - Test message handling (keyboard input)
   - Test render logic independently of Model.View()

---

*Testing analysis: 2026-02-24*
