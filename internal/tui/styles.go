package tui

import "github.com/charmbracelet/lipgloss"

// Warm, muted palette — terracotta, sage, ochre, cream, rust.
var (
	colorTerracotta = lipgloss.Color("#C2704E")
	colorSage       = lipgloss.Color("#8B9F7B")
	colorOchre      = lipgloss.Color("#C49A3C")
	colorCream      = lipgloss.Color("#F5ECD7")
	colorRust       = lipgloss.Color("#A0522D")
	colorRedwood    = lipgloss.Color("#6B3A2A")
	colorMuted      = lipgloss.Color("#8B7D6B")
	colorDim        = lipgloss.Color("#6B6356")

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

	statusLinked  = skillLinked.Render("●")
	statusMissing = skillMissing.Render("○")
	statusStale   = skillStale.Render("◐")
)
