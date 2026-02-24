package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/manzanita-research/chaparral/internal/config"
	"github.com/manzanita-research/chaparral/internal/discovery"
	"github.com/manzanita-research/chaparral/internal/linker"
)

type view int

const (
	viewDashboard view = iota
	viewSyncing
	viewDone
)

type Model struct {
	basePath string
	orgs     []config.Org
	statuses map[string][]linker.LinkStatus // keyed by org name
	results  []linker.LinkResult
	cursor   int
	view     view
	err      error
	width    int
	height   int
}

type orgsLoaded struct {
	orgs     []config.Org
	statuses map[string][]linker.LinkStatus
	err      error
}

type syncDone struct {
	results []linker.LinkResult
	err     error
}

func NewModel(basePath string) Model {
	return Model{
		basePath: basePath,
		statuses: make(map[string][]linker.LinkStatus),
	}
}

func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		orgs, err := discovery.FindOrgs(m.basePath)
		if err != nil {
			return orgsLoaded{err: err}
		}

		statuses := make(map[string][]linker.LinkStatus)
		for _, org := range orgs {
			st, _ := linker.StatusOrg(org)
			statuses[org.Name] = st
		}

		return orgsLoaded{orgs: orgs, statuses: statuses}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.orgs)-1 {
				m.cursor++
			}
		case "s":
			if m.view == viewDashboard && len(m.orgs) > 0 {
				m.view = viewSyncing
				return m, m.syncAll()
			}
		case "enter":
			if m.view == viewDashboard && len(m.orgs) > 0 {
				m.view = viewSyncing
				return m, m.syncOrg(m.cursor)
			}
		case "r":
			return m, m.Init()
		case "esc":
			if m.view == viewDone {
				m.view = viewDashboard
				m.results = nil
				return m, m.Init()
			}
		}

	case orgsLoaded:
		m.err = msg.err
		m.orgs = msg.orgs
		m.statuses = msg.statuses
		m.view = viewDashboard

	case syncDone:
		m.err = msg.err
		m.results = msg.results
		m.view = viewDone
	}

	return m, nil
}

func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\n  %s\n\n  %s\n\n",
			titleStyle.Render("chaparral"),
			skillMissing.Render(fmt.Sprintf("error: %v", m.err)),
		)
	}

	switch m.view {
	case viewSyncing:
		return fmt.Sprintf("\n  %s\n\n  %s\n",
			titleStyle.Render("chaparral"),
			mutedStyle.Render("syncing..."),
		)
	case viewDone:
		return m.renderResults()
	default:
		return m.renderDashboard()
	}
}

func (m Model) renderDashboard() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString("  " + titleStyle.Render("chaparral"))
	b.WriteString("\n")

	if len(m.orgs) == 0 {
		b.WriteString("  " + mutedStyle.Render("no orgs found in "+m.basePath))
		b.WriteString("\n")
		b.WriteString("  " + dimStyle.Render("add a chaparral.json to a brand repo to get started"))
		b.WriteString("\n\n")
		b.WriteString("  " + dimStyle.Render("q quit"))
		b.WriteString("\n")
		return b.String()
	}

	for i, org := range m.orgs {
		cursor := "  "
		if i == m.cursor {
			cursor = lipgloss.NewStyle().Foreground(colorTerracotta).Render("> ")
		}

		b.WriteString(cursor + orgNameStyle.Render(org.Name))
		b.WriteString("  " + dimStyle.Render(fmt.Sprintf("(%s/)", filepath.Base(org.Path))))
		b.WriteString("\n")

		statuses := m.statuses[org.Name]

		// Show CLAUDE.md status
		for _, st := range statuses {
			if st.Skill == "CLAUDE.md" {
				icon := statusIcon(st.State)
				b.WriteString(fmt.Sprintf("    %s CLAUDE.md %s\n",
					icon, dimStyle.Render(st.State)))
			}
		}

		// Group by skill, show which repos have it
		skillRepos := make(map[string][]linker.LinkStatus)
		for _, st := range statuses {
			if st.Skill != "CLAUDE.md" {
				skillRepos[st.Skill] = append(skillRepos[st.Skill], st)
			}
		}

		for skill, repos := range skillRepos {
			linked := 0
			total := len(repos)
			for _, r := range repos {
				if r.State == "linked" {
					linked++
				}
			}

			icon := statusLinked
			if linked == 0 {
				icon = statusMissing
			} else if linked < total {
				icon = statusStale
			}

			repoNames := make([]string, 0, len(repos))
			for _, r := range repos {
				repoNames = append(repoNames, r.Repo)
			}

			b.WriteString(fmt.Sprintf("    %s %s %s %s\n",
				icon,
				repoStyle.Render(skill),
				dimStyle.Render(fmt.Sprintf("(%d/%d)", linked, total)),
				dimStyle.Render(strings.Join(repoNames, ", ")),
			))
		}

		if len(skillRepos) == 0 && len(statuses) <= 1 {
			b.WriteString("    " + dimStyle.Render("no skills found") + "\n")
		}

		b.WriteString("\n")
	}

	b.WriteString("  " + dimStyle.Render("enter sync selected  s sync all  r refresh  q quit"))
	b.WriteString("\n")

	return b.String()
}

func (m Model) renderResults() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString("  " + titleStyle.Render("chaparral"))
	b.WriteString("\n")

	created, existed, skipped, errored := 0, 0, 0, 0
	for _, r := range m.results {
		switch r.Action {
		case "created":
			created++
		case "exists":
			existed++
		case "skipped":
			skipped++
		case "error":
			errored++
		}
	}

	if created > 0 {
		b.WriteString(fmt.Sprintf("  %s %s\n",
			skillLinked.Render(fmt.Sprintf("%d linked", created)),
			dimStyle.Render("(new)"),
		))
	}
	if existed > 0 {
		b.WriteString(fmt.Sprintf("  %s\n",
			mutedStyle.Render(fmt.Sprintf("%d already linked", existed)),
		))
	}
	if skipped > 0 {
		b.WriteString(fmt.Sprintf("  %s\n",
			skillStale.Render(fmt.Sprintf("%d skipped", skipped)),
		))
	}
	if errored > 0 {
		b.WriteString(fmt.Sprintf("  %s\n",
			skillMissing.Render(fmt.Sprintf("%d errors", errored)),
		))
	}

	b.WriteString("\n")

	// Show details for non-"exists" results
	for _, r := range m.results {
		if r.Action == "exists" {
			continue
		}
		icon := statusIcon(r.Action)
		detail := ""
		if r.Detail != "" {
			detail = " " + dimStyle.Render(r.Detail)
		}
		b.WriteString(fmt.Sprintf("  %s %s %s%s\n",
			icon,
			repoStyle.Render(r.Repo),
			mutedStyle.Render(r.Skill),
			detail,
		))
	}

	b.WriteString("\n")
	b.WriteString("  " + dimStyle.Render("esc back  q quit"))
	b.WriteString("\n")

	return b.String()
}

func (m Model) syncAll() tea.Cmd {
	return func() tea.Msg {
		var allResults []linker.LinkResult
		for _, org := range m.orgs {
			results, err := linker.SyncOrg(org)
			if err != nil {
				return syncDone{err: err}
			}
			allResults = append(allResults, results...)
		}
		return syncDone{results: allResults}
	}
}

func (m Model) syncOrg(index int) tea.Cmd {
	return func() tea.Msg {
		if index >= len(m.orgs) {
			return syncDone{}
		}
		results, err := linker.SyncOrg(m.orgs[index])
		return syncDone{results: results, err: err}
	}
}

func statusIcon(state string) string {
	switch state {
	case "linked", "created", "exists":
		return statusLinked
	case "missing", "error", "removed":
		return statusMissing
	case "stale", "skipped", "updated":
		return statusStale
	case "conflict":
		return skillMissing.Render("✕")
	default:
		return statusMissing
	}
}

func Run(basePath string) error {
	if basePath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		basePath = filepath.Join(home, "code")
	}

	p := tea.NewProgram(NewModel(basePath), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
