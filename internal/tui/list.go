package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/lyarwood/cctv/internal/claude"
)

func renderSessionList(sessions []claude.Session, cursor int, width, height int, filter string) string {
	if len(sessions) == 0 {
		msg := "No Claude Code sessions found.\nStart a conversation with `claude` first."
		if filter != "" {
			msg = fmt.Sprintf("No sessions matching %q", filter)
		}
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center,
			dimStyle.Render(msg))
	}

	colStatus := 3
	colMsgs := 6
	colModified := 12
	colBranch := 18
	colProject := 20
	colPR := 30
	colSummary := width - colStatus - colMsgs - colModified - colBranch - colProject - colPR - 12
	if colSummary < 20 {
		colSummary = 20
	}

	header := fmt.Sprintf(" %-*s %-*s %-*s %-*s %-*s %*s %-*s",
		colStatus, " ",
		colSummary, "SUMMARY",
		colProject, "PROJECT",
		colBranch, "BRANCH",
		colPR, "PR",
		colMsgs, "MSGS",
		colModified, "MODIFIED")
	header = headerStyle.Width(width).Render(header)

	availableHeight := height - 4
	if availableHeight < 1 {
		availableHeight = 1
	}

	scrollOffset := 0
	if cursor >= availableHeight {
		scrollOffset = cursor - availableHeight + 1
	}

	var rows []string
	end := scrollOffset + availableHeight
	if end > len(sessions) {
		end = len(sessions)
	}

	for i := scrollOffset; i < end; i++ {
		s := sessions[i]
		row := formatRow(s, colStatus, colSummary, colProject, colBranch, colPR, colMsgs, colModified)
		if i == cursor {
			row = selectedStyle.Width(width).Render(row)
		} else {
			row = lipgloss.NewStyle().Width(width).Render(row)
		}
		rows = append(rows, row)
	}

	content := strings.Join(rows, "\n")

	position := fmt.Sprintf(" %d/%d", cursor+1, len(sessions))
	statusBar := statusBarStyle.Width(width).Render(position)

	return lipgloss.JoinVertical(lipgloss.Left, header, content, statusBar)
}

func formatRow(s claude.Session, colStatus, colSummary, colProject, colBranch, colPR, colMsgs, colModified int) string {
	status := "  "
	if s.IsRunning {
		status = runningStyle.Render("* ")
	}

	summary := s.Summary
	if summary == "" {
		summary = claude.SanitizePrompt(s.FirstPrompt)
	}
	summary = truncateStr(summary, colSummary)

	project := truncateStr(filepath.Base(s.ProjectPath), colProject)
	branch := truncateStr(s.GitBranch, colBranch)
	pr := truncateStr(formatPRLinks(s.PRLinks), colPR)

	msgs := ""
	if s.MessageCount > 0 {
		msgs = fmt.Sprintf("%d", s.MessageCount)
	}

	modified := formatRelativeTime(s.Modified)

	return fmt.Sprintf(" %s %-*s %-*s %-*s %-*s %*s %-*s",
		status,
		colSummary, summary,
		colProject, project,
		colBranch, branch,
		colPR, pr,
		colMsgs, msgs,
		colModified, modified)
}

func formatPRLinks(links []claude.PRLink) string {
	if len(links) == 0 {
		return ""
	}
	first := fmt.Sprintf("%s#%d", links[0].Repository, links[0].Number)
	if len(links) > 1 {
		return fmt.Sprintf("%s +%d", first, len(links)-1)
	}
	return first
}

func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}

func formatRelativeTime(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	default:
		return t.Format("2006-01-02")
	}
}
