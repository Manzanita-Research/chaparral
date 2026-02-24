package tui

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

// Warm, muted palette — terracotta, sage, ochre, cream, rust, fog, lavender.
var (
	colorTerracotta = lipgloss.Color("#C2704E")
	colorSage       = lipgloss.Color("#8B9F7B")
	colorOchre      = lipgloss.Color("#C49A3C")
	colorCream      = lipgloss.Color("#F5ECD7")
	colorRust       = lipgloss.Color("#A0522D")
	colorRedwood    = lipgloss.Color("#6B3A2A")
	colorMuted      = lipgloss.Color("#8B7D6B")
	colorDim        = lipgloss.Color("#6B6356")
	colorFog        = lipgloss.Color("#E8E5DF")
	colorLavender   = lipgloss.Color("#9B8EA8")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorTerracotta).
			MarginBottom(1)

	orgNameStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorOchre)

	repoStyle = lipgloss.NewStyle().
			Foreground(colorCream)

	skillLinked = lipgloss.NewStyle().
			Foreground(colorSage)

	skillMissing = lipgloss.NewStyle().
			Foreground(colorRust)

	skillStale = lipgloss.NewStyle().
			Foreground(colorOchre)

	mutedStyle = lipgloss.NewStyle().
			Foreground(colorMuted)

	dimStyle = lipgloss.NewStyle().
			Foreground(colorDim)

	lavenderStyle = lipgloss.NewStyle().
			Foreground(colorLavender)

	statusLinked  = skillLinked.Render("●")
	statusMissing = skillMissing.Render("○")
	statusStale   = skillStale.Render("◐")
)

// hasNoColor checks if NO_COLOR is set in the environment.
func hasNoColor() bool {
	_, ok := os.LookupEnv("NO_COLOR")
	return ok
}
