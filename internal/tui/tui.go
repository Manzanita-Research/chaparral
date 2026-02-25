package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/manzanita-research/chaparral/internal/config"
	"github.com/manzanita-research/chaparral/internal/discovery"
	"github.com/manzanita-research/chaparral/internal/linker"
	"github.com/manzanita-research/chaparral/internal/marketplace"
)

const maxContentWidth = 80

type view int

const (
	viewDashboard view = iota
	viewSyncing
	viewDone
	viewHelp
	viewInstallPick // picking which plugin to install
	viewInstalling  // install in progress
	viewInstallDone // install result
)

type dashTab int

const (
	tabSkills dashTab = iota
	tabRepos
)

type Model struct {
	basePath string
	orgs     []config.Org
	statuses map[string][]linker.LinkStatus // keyed by org name
	results  []linker.LinkResult
	cursor   int
	view     view
	prevView view // view to return to from help
	tab      dashTab
	err      error
	width    int
	height   int
	spinner  spinner.Model
	noColor  bool

	// Plugin data
	plugins      []marketplace.InstalledPlugin
	available    []marketplace.AvailablePlugin
	pluginErr    error
	pluginLoaded bool

	// Install flow
	installOptions []marketplace.AvailablePlugin // available plugins for selected repo
	installCursor  int
	installPlugin  string
	installRepo    string
	installOutput  string
	installErr     error
}

type orgsLoaded struct {
	orgs     []config.Org
	statuses map[string][]linker.LinkStatus
	err      error
}

type pluginsLoaded struct {
	plugins   []marketplace.InstalledPlugin
	available []marketplace.AvailablePlugin
	err       error
}

type syncDone struct {
	results []linker.LinkResult
	err     error
}

type installDone struct {
	plugin string
	output string
	err    error
}

func NewModel(basePath string) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(colorTerracotta)

	return Model{
		basePath: basePath,
		statuses: make(map[string][]linker.LinkStatus),
		spinner:  s,
		noColor:  hasNoColor(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
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
		},
		func() tea.Msg {
			installed, err := marketplace.ScanInstalled()
			if err != nil {
				return pluginsLoaded{err: err}
			}
			available, availErr := marketplace.QueryAvailable()
			if availErr != nil {
				// Non-fatal — we still have installed data
				return pluginsLoaded{plugins: installed, err: availErr}
			}
			return pluginsLoaded{plugins: installed, available: available}
		},
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		switch m.view {
		case viewInstallPick:
			return m.updateInstallPick(msg)
		case viewInstallDone:
			return m.updateInstallDone(msg)
		default:
			return m.updateDefault(msg)
		}

	case orgsLoaded:
		m.err = msg.err
		m.orgs = msg.orgs
		m.statuses = msg.statuses
		m.view = viewDashboard

	case pluginsLoaded:
		m.pluginErr = msg.err
		m.plugins = msg.plugins
		m.available = msg.available
		m.pluginLoaded = true

	case syncDone:
		m.err = msg.err
		m.results = msg.results
		m.view = viewDone

	case installDone:
		m.installPlugin = msg.plugin
		m.installOutput = msg.output
		m.installErr = msg.err
		m.view = viewInstallDone
	}

	return m, nil
}

func (m Model) updateDefault(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		if m.view == viewHelp {
			m.view = m.prevView
			return m, nil
		}
		return m, tea.Quit
	case "?":
		if m.view == viewHelp {
			m.view = m.prevView
		} else {
			m.prevView = m.view
			m.view = viewHelp
		}
		return m, nil
	case "tab":
		if m.view == viewDashboard {
			if m.tab == tabSkills {
				m.tab = tabRepos
			} else {
				m.tab = tabSkills
			}
			m.cursor = 0
		}
		return m, nil
	case "up", "k":
		if m.view == viewDashboard && m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.view == viewDashboard && m.cursor < len(m.orgs)-1 {
			m.cursor++
		}
	case "s":
		if m.view == viewDashboard && len(m.orgs) > 0 {
			m.view = viewSyncing
			return m, tea.Batch(m.spinner.Tick, m.syncAll())
		}
	case "enter":
		if m.view == viewDashboard && len(m.orgs) > 0 {
			m.view = viewSyncing
			return m, tea.Batch(m.spinner.Tick, m.syncOrg(m.cursor))
		}
	case "r":
		if m.view == viewDashboard {
			m.pluginLoaded = false
			return m, m.Init()
		}
	case "i":
		if m.view == viewDashboard && m.tab == tabRepos && len(m.orgs) > 0 {
			return m.startInstallPick()
		}
	case "esc":
		if m.view == viewHelp {
			m.view = m.prevView
			return m, nil
		}
		if m.view == viewDone {
			m.view = viewDashboard
			m.results = nil
			return m, m.Init()
		}
	}
	return m, nil
}

func (m Model) updateInstallPick(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q":
		m.view = viewDashboard
		return m, nil
	case "up", "k":
		if m.installCursor > 0 {
			m.installCursor--
		}
	case "down", "j":
		if m.installCursor < len(m.installOptions)-1 {
			m.installCursor++
		}
	case "enter":
		if len(m.installOptions) > 0 {
			plugin := m.installOptions[m.installCursor]
			m.installPlugin = plugin.PluginID
			m.view = viewInstalling
			return m, tea.Batch(
				m.spinner.Tick,
				m.doInstall(plugin.PluginID, m.installRepo),
			)
		}
	}
	return m, nil
}

func (m Model) updateInstallDone(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "q", "enter":
		m.view = viewDashboard
		m.installOutput = ""
		m.installErr = nil
		m.pluginLoaded = false
		return m, m.Init()
	}
	return m, nil
}

func (m Model) startInstallPick() (tea.Model, tea.Cmd) {
	if len(m.available) == 0 {
		return m, nil
	}

	org := m.orgs[m.cursor]
	// Pick the first repo for simplicity. In future could let user pick.
	if len(org.Repos) == 0 {
		return m, nil
	}

	repo := org.Repos[0]
	repoPath := filepath.Join(org.Path, repo)

	// Find plugins not yet installed for this repo
	repoPlugins := marketplace.PluginsForRepo(m.plugins, repoPath)
	installedIDs := make(map[string]bool)
	for _, p := range repoPlugins {
		installedIDs[p.PluginID] = true
	}

	var options []marketplace.AvailablePlugin
	for _, a := range m.available {
		if !installedIDs[a.PluginID] {
			options = append(options, a)
		}
	}

	if len(options) == 0 {
		return m, nil
	}

	m.installOptions = options
	m.installCursor = 0
	m.installRepo = repoPath
	m.view = viewInstallPick
	return m, nil
}

// contentWidth returns the usable content width, capped at maxContentWidth.
func (m Model) contentWidth() int {
	w := m.width
	if w <= 0 {
		w = maxContentWidth
	}
	if w > maxContentWidth {
		w = maxContentWidth
	}
	return w
}

// container wraps content with consistent padding.
func (m Model) container(content string) string {
	return lipgloss.NewStyle().
		Padding(0, 2).
		MaxWidth(m.contentWidth()).
		Render(content)
}

func (m Model) View() string {
	if m.err != nil {
		content := titleStyle.Render("chaparral") + "\n\n" +
			skillMissing.Render(fmt.Sprintf("%v", m.err)) + "\n\n" +
			dimStyle.Render("check the path and try again")
		return "\n" + m.container(content) + "\n"
	}

	switch m.view {
	case viewHelp:
		return m.renderHelp()
	case viewSyncing, viewInstalling:
		return m.renderSyncing()
	case viewDone:
		return m.renderResults()
	case viewInstallPick:
		return m.renderInstallPick()
	case viewInstallDone:
		return m.renderInstallDone()
	default:
		return m.renderDashboard()
	}
}

func (m Model) renderSyncing() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render("chaparral"))
	b.WriteString("\n\n")

	label := "syncing links across repos"
	if m.view == viewInstalling {
		name, _ := marketplace.ParsePluginID(m.installPlugin)
		label = fmt.Sprintf("installing %s", name)
	}
	b.WriteString(m.spinner.View() + " " + mutedStyle.Render(label))
	b.WriteString("\n")

	return "\n" + m.container(b.String()) + "\n"
}

func (m Model) renderDashboard() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("chaparral"))
	b.WriteString("\n")

	// Tab bar
	skillsLabel := "skills"
	reposLabel := "repos"
	if m.tab == tabSkills {
		skillsLabel = repoStyle.Render("skills")
		reposLabel = dimStyle.Render("repos")
	} else {
		skillsLabel = dimStyle.Render("skills")
		reposLabel = repoStyle.Render("repos")
	}
	b.WriteString(skillsLabel + "  " + reposLabel + "  " + dimStyle.Render("(tab to switch)"))
	b.WriteString("\n\n")

	if len(m.orgs) == 0 {
		b.WriteString(mutedStyle.Render("no orgs found in "+m.basePath) + "\n")
		b.WriteString(dimStyle.Render("add a chaparral.json to a brand repo to get started") + "\n\n")
		b.WriteString(dimStyle.Render("q quit  ? help"))
		b.WriteString("\n")
		return "\n" + m.container(b.String()) + "\n"
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

		if m.tab == tabSkills {
			m.renderSkillsTab(&b, org, statuses)
		} else {
			m.renderReposTab(&b, org, statuses)
		}

		b.WriteString("\n")
	}

	hint := "enter sync selected  s sync all  tab switch view  r refresh  ? help  q quit"
	if m.tab == tabRepos && len(m.available) > 0 {
		hint = "i install plugin  " + hint
	}
	b.WriteString(dimStyle.Render(hint))
	b.WriteString("\n")

	return "\n" + m.container(b.String()) + "\n"
}

// renderSkillsTab shows skills with linked/total counts.
func (m Model) renderSkillsTab(b *strings.Builder, org config.Org, statuses []linker.LinkStatus) {
	// Show CLAUDE.md status
	for _, st := range statuses {
		if st.Skill == "CLAUDE.md" {
			icon := statusIcon(st.State)
			b.WriteString(fmt.Sprintf("    %s CLAUDE.md %s\n",
				icon, dimStyle.Render(st.State)))
		}
	}

	// Group by skill
	skillRepos := make(map[string][]linker.LinkStatus)
	for _, st := range statuses {
		if st.Skill != "CLAUDE.md" {
			skillRepos[st.Skill] = append(skillRepos[st.Skill], st)
		}
	}

	skillNames := make([]string, 0, len(skillRepos))
	for name := range skillRepos {
		skillNames = append(skillNames, name)
	}
	sort.Strings(skillNames)

	for _, skill := range skillNames {
		repos := skillRepos[skill]
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

		b.WriteString(fmt.Sprintf("    %s %s %s\n",
			icon,
			repoStyle.Render(skill),
			dimStyle.Render(fmt.Sprintf("(%d/%d repos)", linked, total)),
		))
	}

	if len(skillRepos) == 0 && len(statuses) <= 1 {
		b.WriteString("    " + dimStyle.Render("no skills found") + "\n")
	}

	// Marketplace summary
	if m.pluginLoaded {
		installedCount := len(m.plugins)
		availableCount := len(m.available)
		if installedCount > 0 || availableCount > 0 {
			b.WriteString("    " + dimStyle.Render("marketplace") + "\n")
			b.WriteString(fmt.Sprintf("      %s\n",
				dimStyle.Render(fmt.Sprintf("%d installed, %d available", installedCount, availableCount)),
			))
		}
	}
}

// renderReposTab shows repos with their skill statuses and plugin info.
func (m Model) renderReposTab(b *strings.Builder, org config.Org, statuses []linker.LinkStatus) {
	// Group by repo
	repoSkills := make(map[string][]linker.LinkStatus)
	var repoOrder []string
	seen := make(map[string]bool)
	for _, st := range statuses {
		if st.Skill == "CLAUDE.md" {
			continue
		}
		if !seen[st.Repo] {
			repoOrder = append(repoOrder, st.Repo)
			seen[st.Repo] = true
		}
		repoSkills[st.Repo] = append(repoSkills[st.Repo], st)
	}
	sort.Strings(repoOrder)

	// Show CLAUDE.md first since it applies to the whole org
	for _, st := range statuses {
		if st.Skill == "CLAUDE.md" {
			icon := statusIcon(st.State)
			b.WriteString(fmt.Sprintf("    %s CLAUDE.md %s\n",
				icon, dimStyle.Render(st.State)))
		}
	}

	for _, repo := range repoOrder {
		skills := repoSkills[repo]
		linked := 0
		for _, s := range skills {
			if s.State == "linked" {
				linked++
			}
		}

		b.WriteString(fmt.Sprintf("    %s %s\n",
			repoStyle.Render(repo),
			dimStyle.Render(fmt.Sprintf("(%d/%d skills)", linked, len(skills))),
		))

		for _, s := range skills {
			icon := statusIcon(s.State)
			b.WriteString(fmt.Sprintf("      %s %s\n", icon, dimStyle.Render(s.Skill)))
		}

		// Show plugins for this repo
		if m.pluginLoaded {
			repoPath := filepath.Join(org.Path, repo)
			pluginStatuses := marketplace.MergeStatus(m.plugins, m.available, repoPath)
			if len(pluginStatuses) > 0 {
				b.WriteString("      " + dimStyle.Render("plugins") + "\n")
				for _, ps := range pluginStatuses {
					icon := pluginAvailable
					detail := "available"
					if ps.Installed && ps.Enabled {
						icon = pluginInstalled
						detail = fmt.Sprintf("v%s, %s", ps.Version, ps.Scope)
					} else if ps.Installed {
						icon = pluginDisabled
						detail = fmt.Sprintf("v%s, disabled", ps.Version)
					} else if ps.Available {
						detail = fmt.Sprintf("v%s", ps.AvailableVersion)
					}
					b.WriteString(fmt.Sprintf("        %s %s %s\n",
						icon,
						dimStyle.Render(ps.Name),
						dimStyle.Render("("+detail+")"),
					))
				}
			}
		}
	}

	if len(repoOrder) == 0 {
		b.WriteString("    " + dimStyle.Render("no repos found") + "\n")
	}
}

func (m Model) renderResults() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("chaparral"))
	b.WriteString("\n")

	// Count by outcome
	counts := map[string]int{}
	for _, r := range m.results {
		counts[r.Action]++
	}

	if n := counts["created"]; n > 0 {
		b.WriteString(fmt.Sprintf("%s %s\n",
			skillLinked.Render(fmt.Sprintf("%d linked", n)),
			dimStyle.Render("(new)"),
		))
	}
	if n := counts["exists"]; n > 0 {
		b.WriteString(mutedStyle.Render(fmt.Sprintf("%d already linked", n)) + "\n")
	}
	if n := counts["skipped"]; n > 0 {
		b.WriteString(skillStale.Render(fmt.Sprintf("%d skipped", n)) + "\n")
	}
	if n := counts["error"]; n > 0 {
		b.WriteString(skillMissing.Render(fmt.Sprintf("%d failed", n)) + "\n")
	}

	b.WriteString("\n")

	// Group results by outcome: created first, then updated, skipped, errors
	// Skip "exists" since those are just confirmations
	order := []string{"created", "updated", "skipped", "error"}
	for _, action := range order {
		for _, r := range m.results {
			if r.Action != action {
				continue
			}
			icon := statusIcon(r.Action)
			detail := ""
			if r.Detail != "" {
				detail = " " + dimStyle.Render(r.Detail)
			}
			b.WriteString(fmt.Sprintf("%s %s %s%s\n",
				icon,
				repoStyle.Render(r.Repo),
				mutedStyle.Render(r.Skill),
				detail,
			))
		}
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("esc back  ? help  q quit"))
	b.WriteString("\n")

	return "\n" + m.container(b.String()) + "\n"
}

func (m Model) renderInstallPick() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("chaparral"))
	b.WriteString("\n")
	b.WriteString(lavenderStyle.Render("install plugin"))
	b.WriteString("\n\n")

	repoName := filepath.Base(m.installRepo)
	b.WriteString(dimStyle.Render(fmt.Sprintf("into %s:", repoName)))
	b.WriteString("\n\n")

	for i, opt := range m.installOptions {
		cursor := "  "
		if i == m.installCursor {
			cursor = lipgloss.NewStyle().Foreground(colorTerracotta).Render("> ")
		}
		name := repoStyle.Render(opt.Name)
		desc := ""
		if opt.Description != "" {
			// Truncate long descriptions
			d := opt.Description
			if len(d) > 50 {
				d = d[:47] + "..."
			}
			desc = " " + dimStyle.Render(d)
		}
		b.WriteString(fmt.Sprintf("%s%s%s\n", cursor, name, desc))
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("enter install  esc cancel"))
	b.WriteString("\n")

	return "\n" + m.container(b.String()) + "\n"
}

func (m Model) renderInstallDone() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("chaparral"))
	b.WriteString("\n")

	name, _ := marketplace.ParsePluginID(m.installPlugin)
	repoName := filepath.Base(m.installRepo)

	if m.installErr != nil {
		b.WriteString(skillMissing.Render(fmt.Sprintf("could not install %s into %s", name, repoName)))
		b.WriteString("\n\n")
		if m.installOutput != "" {
			for _, line := range strings.Split(strings.TrimSpace(m.installOutput), "\n") {
				b.WriteString("  " + dimStyle.Render(line) + "\n")
			}
		}
	} else {
		b.WriteString(skillLinked.Render(fmt.Sprintf("installed %s into %s", name, repoName)))
		b.WriteString("\n\n")
		if m.installOutput != "" {
			for _, line := range strings.Split(strings.TrimSpace(m.installOutput), "\n") {
				b.WriteString("  " + dimStyle.Render(line) + "\n")
			}
		}
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("esc back  q quit"))
	b.WriteString("\n")

	return "\n" + m.container(b.String()) + "\n"
}

func (m Model) renderHelp() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("chaparral"))
	b.WriteString("\n")
	b.WriteString(lavenderStyle.Render("keybindings"))
	b.WriteString("\n\n")

	keys := []struct{ key, desc string }{
		{"j/k", "navigate orgs"},
		{"tab", "switch skills/repos view"},
		{"enter", "sync selected org"},
		{"s", "sync all orgs"},
		{"i", "install plugin (repos tab)"},
		{"r", "refresh status"},
		{"esc", "back"},
		{"?", "toggle help"},
		{"q", "quit"},
	}

	for _, k := range keys {
		b.WriteString(fmt.Sprintf("  %s  %s\n",
			repoStyle.Render(fmt.Sprintf("%-8s", k.key)),
			dimStyle.Render(k.desc),
		))
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("symbols"))
	b.WriteString("\n\n")

	symbols := []struct{ sym, desc string }{
		{statusLinked, "linked / installed"},
		{statusMissing, "missing"},
		{statusStale, "partially linked / disabled"},
		{pluginAvailable, "available (not installed)"},
		{skillMissing.Render("✕"), "conflict (non-symlink exists)"},
	}

	for _, s := range symbols {
		b.WriteString(fmt.Sprintf("  %s  %s\n", s.sym, dimStyle.Render(s.desc)))
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("esc back  ? close"))
	b.WriteString("\n")

	return "\n" + m.container(b.String()) + "\n"
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

func (m Model) doInstall(pluginID, repoPath string) tea.Cmd {
	return func() tea.Msg {
		output, err := marketplace.Install(pluginID, repoPath)
		return installDone{
			plugin: pluginID,
			output: output,
			err:    err,
		}
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
