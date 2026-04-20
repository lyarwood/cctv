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
	colTokens := 9
	colModified := 12
	colBranch := 18
	colProject := 20
	colPR := 30
	colSummary := width - colStatus - colMsgs - colTokens - colModified - colBranch - colProject - colPR - 14
	if colSummary < 20 {
		colSummary = 20
	}

	header := fmt.Sprintf(" %-*s %-*s %-*s %-*s %-*s %*s %*s %-*s",
		colStatus, " ",
		colSummary, "SUMMARY",
		colProject, "PROJECT",
		colBranch, "BRANCH",
		colPR, "PR",
		colTokens, "TOKENS",
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
		row := formatRow(s, colStatus, colSummary, colProject, colBranch, colPR, colTokens, colMsgs, colModified)
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

func formatRow(s claude.Session, colStatus, colSummary, colProject, colBranch, colPR, colTokens, colMsgs, colModified int) string {
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

	tokens := ""
	if s.TotalTokens > 0 {
		tokens = colorizeTokens(s.TotalTokens, colTokens)
	}

	msgs := ""
	if s.MessageCount > 0 {
		msgs = fmt.Sprintf("%d", s.MessageCount)
	}

	modified := formatRelativeTime(s.Modified)

	return fmt.Sprintf(" %s %-*s %-*s %-*s %-*s %*s %*s %-*s",
		status,
		colSummary, summary,
		colProject, project,
		colBranch, branch,
		colPR, pr,
		colTokens, tokens,
		colMsgs, msgs,
		colModified, modified)
}

func colorizeTokens(tokens int64, width int) string {
	text := formatTokenCount(tokens)
	style := tokensLowStyle
	switch {
	case tokens >= 500_000:
		style = tokensHighStyle
	case tokens >= 100_000:
		style = tokensMedStyle
	}
	return style.Render(text)
}

func formatTokenCount(n int64) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
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
