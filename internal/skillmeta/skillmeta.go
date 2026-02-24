package skillmeta

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Frontmatter holds the parsed metadata from a SKILL.md file.
type Frontmatter struct {
	Name        string
	Description string
	License     string
}

// ParseFrontmatter reads SKILL.md frontmatter (key: value pairs between --- delimiters).
// The file must start with --- on the first line. Parsing ends at the closing --- or EOF.
func ParseFrontmatter(path string) (Frontmatter, error) {
	f, err := os.Open(path)
	if err != nil {
		return Frontmatter{}, fmt.Errorf("can't open %s: %w", path, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	// First line must be ---
	if !scanner.Scan() || strings.TrimSpace(scanner.Text()) != "---" {
		return Frontmatter{}, fmt.Errorf("SKILL.md at %s has no frontmatter (missing opening ---)", path)
	}

	var fm Frontmatter
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			break
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "name":
			fm.Name = val
		case "description":
			fm.Description = val
		case "license":
			fm.License = val
		}
	}

	if err := scanner.Err(); err != nil {
		return Frontmatter{}, fmt.Errorf("reading %s: %w", path, err)
	}

	return fm, nil
}
