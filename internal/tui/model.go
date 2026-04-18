package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/lyarwood/cctv/internal/claude"
)

func resumeSession(session claude.Session) tea.Cmd {
	c := exec.Command("claude", "--resume", session.SessionID)
	if info, err := os.Stat(session.ProjectPath); err == nil && info.IsDir() {
		c.Dir = session.ProjectPath
	}
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return sessionResumedMsg{err: err}
	})
}

type viewState int

const (
	viewList viewState = iota
	viewDetail
	viewStats
)

type detailLoadedMsg struct {
	detail *claude.SessionDetail
	err    error
}

type sessionsRefreshedMsg struct {
	sessions []claude.Session
	err      error
}

type sessionResumedMsg struct {
	err error
}

type Model struct {
	view        viewState
	sessions    []claude.Session
	filtered    []claude.Session
	cursor      int
	detail      *claude.SessionDetail
	err         error
	width       int
	height      int
	filterInput textinput.Model
	filtering   bool
	filterText  string
	showHelp    bool
	discoverer  *claude.Discoverer
}

func NewModel(sessions []claude.Session, discoverer *claude.Discoverer, initialFilter string) Model {
	ti := textinput.New()
	ti.Placeholder = "filter sessions..."
	ti.CharLimit = 100

	if initialFilter != "" {
		ti.SetValue(initialFilter)
	}

	return Model{
		sessions:    sessions,
		filtered:    sessions,
		filterInput: ti,
		filterText:  initialFilter,
		discoverer:  discoverer,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case detailLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.detail = msg.detail
		}
		return m, nil

	case sessionsRefreshedMsg:
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.sessions = msg.sessions
			m.applyFilter(m.filterText)
		}
		return m, nil

	case sessionResumedMsg:
		return m, nil

	case tea.KeyMsg:
		if m.filtering {
			return m.updateFiltering(msg)
		}
		switch m.view {
		case viewDetail:
			return m.updateDetail(msg)
		case viewStats:
			return m.updateStats(msg)
		default:
			return m.updateList(msg)
		}
	}

	return m, nil
}

func (m Model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case matchKey(msg, keys.Quit):
		return m, tea.Quit

	case matchKey(msg, keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}

	case matchKey(msg, keys.Down):
		if m.cursor < len(m.filtered)-1 {
			m.cursor++
		}

	case matchKey(msg, keys.Enter):
		if len(m.filtered) > 0 {
			session := m.filtered[m.cursor]
			return m, resumeSession(session)
		}

	case matchKey(msg, keys.Detail):
		if len(m.filtered) > 0 {
			m.view = viewDetail
			m.detail = nil
			m.err = nil
			session := m.filtered[m.cursor]
			return m, m.loadDetail(session)
		}

	case matchKey(msg, keys.Stats):
		if len(m.filtered) > 0 {
			m.view = viewStats
			m.detail = nil
			m.err = nil
			session := m.filtered[m.cursor]
			return m, m.loadDetail(session)
		}

	case matchKey(msg, keys.Filter):
		m.filtering = true
		m.filterInput.Focus()
		return m, textinput.Blink

	case matchKey(msg, keys.Refresh):
		return m, m.refreshSessions()

	case matchKey(msg, keys.Help):
		m.showHelp = !m.showHelp
	}

	return m, nil
}

func (m Model) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case matchKey(msg, keys.Back), matchKey(msg, keys.Quit):
		m.view = viewList
		m.detail = nil

	case matchKey(msg, keys.Enter):
		if len(m.filtered) > 0 {
			session := m.filtered[m.cursor]
			return m, resumeSession(session)
		}
	}

	return m, nil
}

func (m Model) updateStats(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case matchKey(msg, keys.Back), matchKey(msg, keys.Quit), matchKey(msg, keys.Stats):
		m.view = viewList
		m.detail = nil

	case matchKey(msg, keys.Enter):
		if len(m.filtered) > 0 {
			return m, resumeSession(m.filtered[m.cursor])
		}

	case matchKey(msg, keys.Detail):
		m.view = viewDetail
	}

	return m, nil
}

func (m Model) updateFiltering(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.filtering = false
		m.filterInput.Blur()
		m.filterText = m.filterInput.Value()
		return m, nil

	case "esc":
		m.filtering = false
		m.filterInput.Blur()
		m.filterInput.SetValue(m.filterText)
		m.applyFilter(m.filterText)
		return m, nil

	case "tab":
		m.cycleFilterPrefix()
		m.applyFilter(m.filterInput.Value())
		return m, nil
	}

	var cmd tea.Cmd
	m.filterInput, cmd = m.filterInput.Update(msg)
	m.applyFilter(m.filterInput.Value())
	return m, cmd
}

var filterPrefixes = []string{"", "project:", "branch:", "cwd:", "pr:"}

func (m *Model) cycleFilterPrefix() {
	val := m.filterInput.Value()
	currentPrefix := ""
	rest := val
	for _, p := range filterPrefixes[1:] {
		if strings.HasPrefix(val, p) {
			currentPrefix = p
			rest = val[len(p):]
			break
		}
	}
	nextIdx := 0
	for i, p := range filterPrefixes {
		if p == currentPrefix {
			nextIdx = (i + 1) % len(filterPrefixes)
			break
		}
	}
	m.filterInput.SetValue(filterPrefixes[nextIdx] + rest)
}

func (m *Model) applyFilter(text string) {
	m.filterText = text
	if text == "" {
		m.filtered = m.sessions
	} else {
		m.filtered = matchSessions(m.sessions, text)
	}
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
}

func matchSessions(sessions []claude.Session, text string) []claude.Session {
	var filtered []claude.Session
	for _, s := range sessions {
		if matchSession(s, text) {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

func matchSession(s claude.Session, text string) bool {
	for _, term := range strings.Fields(text) {
		if !matchTerm(s, term) {
			return false
		}
	}
	return true
}

func matchTerm(s claude.Session, term string) bool {
	prefix, query, hasPrefix := strings.Cut(term, ":")
	if hasPrefix {
		switch strings.ToLower(prefix) {
		case "project":
			return flexMatch(filepath.Base(s.ProjectPath), query) ||
				flexMatch(s.ProjectPath, query)
		case "branch":
			return flexMatch(s.GitBranch, query)
		case "cwd":
			return flexMatch(s.ProjectPath, query)
		case "pr":
			return matchPRLinks(s.PRLinks, query)
		}
	}

	return flexMatch(s.Summary, term) ||
		flexMatch(s.FirstPrompt, term) ||
		flexMatch(s.ProjectPath, term) ||
		flexMatch(s.GitBranch, term)
}

func matchPRLinks(links []claude.PRLink, query string) bool {
	for _, pr := range links {
		display := fmt.Sprintf("%s#%d", pr.Repository, pr.Number)
		if flexMatch(display, query) || flexMatch(pr.URL, query) {
			return true
		}
	}
	return false
}

func flexMatch(text, pattern string) bool {
	re, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		return strings.Contains(strings.ToLower(text), strings.ToLower(pattern))
	}
	return re.MatchString(text)
}

func (m Model) loadDetail(session claude.Session) tea.Cmd {
	return func() tea.Msg {
		detail, err := m.discoverer.LoadDetail(session)
		return detailLoadedMsg{detail: detail, err: err}
	}
}

func (m Model) refreshSessions() tea.Cmd {
	return func() tea.Msg {
		sessions, err := m.discoverer.DiscoverAll()
		return sessionsRefreshedMsg{sessions: sessions, err: err}
	}
}

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	title := titleStyle.Render("cctv - Claude Code TUI Viewer")

	var content string
	switch m.view {
	case viewList:
		content = renderSessionList(m.filtered, m.cursor, m.width, m.height-3, m.filterText)
	case viewDetail:
		if len(m.filtered) > 0 {
			content = renderDetail(m.filtered[m.cursor], m.detail, m.width, m.height-3)
		}
	case viewStats:
		if len(m.filtered) > 0 {
			popup := renderStats(m.filtered[m.cursor], m.detail, m.width)
			content = lipgloss.Place(m.width, m.height-3, lipgloss.Center, lipgloss.Center, popup)
		}
	}

	if m.err != nil {
		content += "\n" + errorStyle.Render("Error: "+m.err.Error())
	}

	var footer string
	if m.filtering {
		footer = "Filter: " + m.filterInput.View() + "  " + helpStyle.Render("tab:cycle prefix (project:/branch:/cwd:)  enter:apply  esc:cancel")
	} else if m.filterText != "" {
		footer = helpStyle.Render("active filter: "+m.filterText+"  /:edit  ?:help  q:quit")
	} else if m.showHelp {
		footer = helpStyle.Render("enter:resume  d/space:detail  s:stats  /:filter  r:refresh  ?:help  esc:back  q:quit")
	} else {
		footer = helpStyle.Render("?:help  q:quit")
	}

	return title + "\n" + content + "\n" + footer
}

func matchKey(msg tea.KeyMsg, binding key.Binding) bool {
	for _, k := range binding.Keys() {
		if msg.String() == k {
			return true
		}
	}
	return false
}
