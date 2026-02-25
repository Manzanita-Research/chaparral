package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/manzanita-research/chaparral/internal/config"
	"github.com/manzanita-research/chaparral/internal/discovery"
	"github.com/manzanita-research/chaparral/internal/generator"
	"github.com/manzanita-research/chaparral/internal/linker"
	"github.com/manzanita-research/chaparral/internal/publisher"
	"github.com/manzanita-research/chaparral/internal/tui"
	"github.com/manzanita-research/chaparral/internal/validator"
)

func main() {
	basePath := defaultBasePath()

	if len(os.Args) < 2 {
		// No subcommand — launch TUI
		if err := tui.Run(basePath); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	switch os.Args[1] {
	case "sync":
		runSync(basePath)
	case "status":
		runStatus(basePath)
	case "validate":
		runValidate(basePath)
	case "generate":
		runGenerate(basePath)
	case "publish":
		runPublish(basePath)
	case "unlink":
		runUnlink(basePath)
	case "help", "--help", "-h":
		printHelp()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printHelp()
		os.Exit(1)
	}
}

func defaultBasePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	return filepath.Join(home, "code")
}

func loadOrgs(basePath string) []config.Org {
	orgs, err := discovery.FindOrgs(basePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error discovering orgs: %v\n", err)
		os.Exit(1)
	}
	if len(orgs) == 0 {
		fmt.Println("no orgs found. add a chaparral.json to a brand repo to get started.")
		os.Exit(0)
	}
	return orgs
}

func runSync(basePath string) {
	orgs := loadOrgs(basePath)

	for _, org := range orgs {
		fmt.Printf("%s\n", org.Name)
		results, err := linker.SyncOrg(org)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  error: %v\n", err)
			continue
		}

		for _, r := range results {
			if r.Action == "exists" {
				continue
			}
			detail := ""
			if r.Detail != "" {
				detail = " (" + r.Detail + ")"
			}
			fmt.Printf("  %s %s/%s%s\n", actionIcon(r.Action), r.Repo, r.Skill, detail)
		}

		created := countAction(results, "created")
		existed := countAction(results, "exists")
		fmt.Printf("  %d linked, %d already up to date\n\n", created, existed)
	}
}

func runStatus(basePath string) {
	orgs := loadOrgs(basePath)

	for _, org := range orgs {
		fmt.Printf("%s (%s/)\n", org.Name, filepath.Base(org.Path))

		statuses, err := linker.StatusOrg(org)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  error: %v\n", err)
			continue
		}

		// Group by skill
		skillMap := make(map[string][]linker.LinkStatus)
		var skillOrder []string
		for _, st := range statuses {
			if _, seen := skillMap[st.Skill]; !seen {
				skillOrder = append(skillOrder, st.Skill)
			}
			skillMap[st.Skill] = append(skillMap[st.Skill], st)
		}

		for _, skill := range skillOrder {
			sts := skillMap[skill]
			if skill == "CLAUDE.md" {
				fmt.Printf("  %s CLAUDE.md — %s\n", stateIcon(sts[0].State), sts[0].State)
				continue
			}

			var linked, missing []string
			for _, st := range sts {
				if st.State == "linked" {
					linked = append(linked, st.Repo)
				} else {
					missing = append(missing, st.Repo)
				}
			}

			parts := []string{fmt.Sprintf("  %s %s", stateIcon(sts[0].State), skill)}
			if len(linked) > 0 {
				parts = append(parts, fmt.Sprintf("linked: %s", strings.Join(linked, ", ")))
			}
			if len(missing) > 0 {
				parts = append(parts, fmt.Sprintf("missing: %s", strings.Join(missing, ", ")))
			}
			fmt.Println(strings.Join(parts, "  "))
		}
		fmt.Println()
	}
}

func runValidate(basePath string) {
	orgs := loadOrgs(basePath)
	hasErrors := false

	for _, org := range orgs {
		fmt.Printf("%s (%s/)\n", org.Name, filepath.Base(org.Path))

		results, err := validator.ValidateOrg(org)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  %v\n", err)
			hasErrors = true
			continue
		}

		if len(results) == 0 {
			fmt.Println("  no skills found")
			fmt.Println()
			continue
		}

		for _, r := range results {
			if r.IsValid() && len(r.Warnings) == 0 {
				fmt.Printf("  ✓ %s\n", r.Skill)
			} else if r.IsValid() {
				fmt.Printf("  ~ %s\n", r.Skill)
			} else {
				fmt.Printf("  ✕ %s\n", r.Skill)
				hasErrors = true
			}

			for _, e := range r.Errors {
				fmt.Printf("    ✕ %s\n", e)
			}
			for _, w := range r.Warnings {
				fmt.Printf("    ~ %s\n", w)
			}
		}
		fmt.Println()
	}

	if hasErrors {
		os.Exit(1)
	}
}

func runGenerate(basePath string) {
	orgs := loadOrgs(basePath)

	// Check for --marketplace flag
	showMarketplace := false
	for _, arg := range os.Args[2:] {
		if arg == "--marketplace" {
			showMarketplace = true
		}
	}

	for _, org := range orgs {
		fmt.Printf("%s\n", org.Name)

		skills, err := discovery.FindSkills(org.SkillsPath())
		if err != nil {
			fmt.Fprintf(os.Stderr, "  %v\n", err)
			continue
		}

		if len(skills) == 0 {
			fmt.Println("  no skills found")
			fmt.Println()
			continue
		}

		for _, skill := range skills {
			data, err := generator.GeneratePluginJSON(skill)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  %s/plugin.json — %v\n", skill.Name, err)
				continue
			}

			fmt.Printf("  %s/plugin.json\n", skill.Name)
			for _, line := range strings.Split(string(data), "\n") {
				fmt.Printf("    %s\n", line)
			}
			fmt.Println()
		}

		if showMarketplace {
			data, err := generator.GenerateMarketplaceJSON(org, skills)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  marketplace.json — %v\n", err)
			} else {
				fmt.Println("  marketplace.json")
				for _, line := range strings.Split(string(data), "\n") {
					fmt.Printf("    %s\n", line)
				}
				fmt.Println()
			}
		}
	}
}

func runUnlink(basePath string) {
	orgs := loadOrgs(basePath)

	for _, org := range orgs {
		fmt.Printf("%s\n", org.Name)
		results, err := linker.UnlinkOrg(org)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  error: %v\n", err)
			continue
		}

		for _, r := range results {
			fmt.Printf("  removed %s/%s\n", r.Repo, r.Skill)
		}

		if len(results) == 0 {
			fmt.Println("  nothing to unlink")
		}
		fmt.Println()
	}
}

func runPublish(basePath string) {
	orgs := loadOrgs(basePath)

	// Parse flags
	checkOnly := false
	writeOnly := false
	for _, arg := range os.Args[2:] {
		switch arg {
		case "--check":
			checkOnly = true
		case "--write-only":
			writeOnly = true
		}
	}

	for _, org := range orgs {
		fmt.Printf("%s\n", org.Name)

		skills, err := discovery.FindSkills(org.SkillsPath())
		if err != nil {
			fmt.Fprintf(os.Stderr, "  %v\n", err)
			continue
		}

		if len(skills) == 0 {
			fmt.Println("  no skills found")
			fmt.Println()
			continue
		}

		if checkOnly {
			runPublishCheck(org, skills)
			continue
		}

		if writeOnly {
			runPublishWriteOnly(org, skills)
			continue
		}

		runPublishFull(org, skills)
	}
}

func runPublishCheck(org config.Org, skills []config.Skill) {
	results, err := publisher.CheckFreshness(org, skills)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  %v\n", err)
		return
	}

	upToDate := 0
	for _, r := range results {
		if r.Stale {
			fmt.Printf("  ○ %s (stale)\n", r.Skill)
		} else {
			version := r.PublishedVersion
			if version == "" {
				version = "unpublished"
			}
			fmt.Printf("  ✓ %s (v%s)\n", r.Skill, version)
			upToDate++
		}
	}
	fmt.Printf("  %d of %d skills are up to date\n\n", upToDate, len(results))
}

func runPublishWriteOnly(org config.Org, skills []config.Skill) {
	changes, err := publisher.DiffManifests(org, skills)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  %v\n", err)
		return
	}

	allUnchanged := true
	for _, c := range changes {
		switch c.Kind {
		case "new":
			fmt.Printf("  + %s (new)\n", c.Path)
			allUnchanged = false
		case "modified":
			fmt.Printf("  ~ %s (modified)\n", c.Path)
			allUnchanged = false
		default:
			fmt.Printf("    %s (unchanged)\n", c.Path)
		}
	}

	if allUnchanged {
		fmt.Println("  nothing to write")
		fmt.Println()
		return
	}

	written, err := publisher.WriteManifests(org, skills)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  %v\n", err)
		return
	}
	fmt.Printf("  wrote %d files\n\n", len(written))
}

func runPublishFull(org config.Org, skills []config.Skill) {
	// Show diff preview
	changes, err := publisher.DiffManifests(org, skills)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  %v\n", err)
		return
	}

	allUnchanged := true
	changeCount := 0
	for _, c := range changes {
		switch c.Kind {
		case "new":
			fmt.Printf("  + %s (new)\n", c.Path)
			allUnchanged = false
			changeCount++
		case "modified":
			fmt.Printf("  ~ %s (modified)\n", c.Path)
			allUnchanged = false
			changeCount++
		default:
			fmt.Printf("    %s (unchanged)\n", c.Path)
		}
	}

	if allUnchanged {
		fmt.Println("  already up to date — nothing to publish")
		fmt.Println()
		return
	}

	// Extract version from the first change's new content
	version := extractVersion(changes)

	// Get remote URL
	brandRepoPath := filepath.Join(org.Path, org.BrandRepo)
	remoteURL, err := publisher.RemoteURL(brandRepoPath)
	if err != nil {
		remoteURL = "(no remote configured)"
	}

	fmt.Println()
	fmt.Printf("  Publish marketplace v%s to %s?\n", version, remoteURL)
	fmt.Printf("  Files to write: %d\n\n", changeCount)

	if !confirm("  Push to GitHub?") {
		fmt.Println("  cancelled.")
		fmt.Println()
		return
	}

	if !confirm("  Confirm push (this will update the live marketplace):") {
		fmt.Println("  cancelled.")
		fmt.Println()
		return
	}

	// Write manifests
	written, err := publisher.WriteManifests(org, skills)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  %v\n", err)
		return
	}

	// Commit and push
	err = publisher.CommitAndPush(brandRepoPath, written, version)
	if err != nil {
		if errors.Is(err, publisher.ErrNoChanges) {
			fmt.Println("  already up to date")
			fmt.Println()
			return
		}
		fmt.Fprintf(os.Stderr, "  %v\n", err)
		fmt.Println()
		return
	}

	fmt.Printf("  published v%s to %s\n\n", version, remoteURL)
}

// extractVersion parses the version from the first change's NewContent JSON.
func extractVersion(changes []publisher.FileChange) string {
	for _, c := range changes {
		if c.Kind == "unchanged" {
			continue
		}
		var pm struct {
			Version string `json:"version"`
		}
		if err := json.Unmarshal([]byte(c.NewContent), &pm); err == nil && pm.Version != "" {
			return pm.Version
		}
	}
	return "0.1.0"
}

// confirm prompts the user and returns true if they type "y".
func confirm(prompt string) bool {
	fmt.Printf("%s [y/N] ", prompt)
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(strings.ToLower(line)) == "y"
}

func printHelp() {
	fmt.Println(`chaparral — the connective tissue between your projects

usage:
  chaparral            launch interactive dashboard
  chaparral sync       link skills to all sibling repos
  chaparral status     show current link state
  chaparral validate   check skill structure for errors
  chaparral generate   generate plugin manifests (dry run to stdout)
    --marketplace      also generate marketplace.json catalog
  chaparral publish    write manifests and push marketplace to GitHub
    --check            check if local skills are newer than published
    --write-only       write manifests without pushing to GitHub
  chaparral unlink     remove all managed symlinks
  chaparral help       show this message`)
}

func actionIcon(action string) string {
	switch action {
	case "created":
		return "+"
	case "skipped":
		return "~"
	case "error":
		return "!"
	case "removed":
		return "-"
	default:
		return " "
	}
}

func stateIcon(state string) string {
	switch state {
	case "linked":
		return "✓"
	case "missing":
		return "○"
	case "stale":
		return "◐"
	case "conflict":
		return "✕"
	default:
		return "?"
	}
}

func countAction(results []linker.LinkResult, action string) int {
	n := 0
	for _, r := range results {
		if r.Action == action {
			n++
		}
	}
	return n
}
