package tui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/lyarwood/cctv/internal/claude"
)

func renderSessionList(sessions []claude.Session, cursor int, width, height int, filter string, searchResults []claude.SearchResult) string {
	if len(sessions) == 0 {
		msg := "No Claude Code sessions found.\nStart a conversation with `claude` first."
		if len(searchResults) == 0 && filter != "" {
			msg = fmt.Sprintf("No sessions matching %q", filter)
		} else if searchResults != nil {
			msg = "No matching conversations found."
		}
		return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center,
			dimStyle.Render(msg))
	}

	colStatus := 3
	colMsgs := 6
	colTokens := 9
	colCost := 9
	colModified := 12
	colBranch := 18
	colProject := 20
	colPR := 30
	colSummary := width - colStatus - colMsgs - colTokens - colCost - colModified - colBranch - colProject - colPR - 16
	if colSummary < 20 {
		colSummary = 20
	}

	summaryHeader := "SUMMARY"
	if len(searchResults) > 0 {
		summaryHeader = "MATCH"
	}

	header := fmt.Sprintf(" %-*s %-*s %-*s %-*s %-*s %*s %*s %*s %-*s",
		colStatus, " ",
		colSummary, summaryHeader,
		colProject, "PROJECT",
		colBranch, "BRANCH",
		colPR, "PR",
		colTokens, "TOKENS",
		colCost, "COST",
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
		var snippet *claude.SearchResult
		if i < len(searchResults) {
			snippet = &searchResults[i]
		}
		row := formatRow(s, colSummary, colProject, colBranch, colPR, colTokens, colCost, colMsgs, colModified, snippet)
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

func formatRow(s claude.Session, colSummary, colProject, colBranch, colPR, colTokens, colCost, colMsgs, colModified int, snippet *claude.SearchResult) string {
	status := "  "
	if s.IsRunning {
		status = runningStyle.Render("* ")
	}

	var summaryCol string
	if snippet != nil {
		summaryCol = renderSnippet(*snippet, colSummary)
	} else {
		summary := s.Summary
		if summary == "" {
			summary = claude.SanitizePrompt(s.FirstPrompt)
		}
		summaryCol = truncateStr(summary, colSummary)
	}

	project := truncateStr(filepath.Base(s.ProjectPath), colProject)
	branch := truncateStr(s.GitBranch, colBranch)
	pr := truncateStr(formatPRLinks(s.PRLinks), colPR)

	tokens := fmt.Sprintf("%*s", colTokens, "")
	if s.TotalTokens > 0 {
		tokens = colorizeTokens(s.TotalTokens, colTokens)
	}

	cost := fmt.Sprintf("%*s", colCost, "")
	if s.LastInputTokens > 0 && s.LastModel != "" {
		coldCost, _ := claude.EstimateResumeCost(s.LastModel, s.LastInputTokens)
		if coldCost > 0 {
			cost = padRight(formatCost(coldCost), colCost)
		}
	}

	msgs := fmt.Sprintf("%*s", colMsgs, "")
	if s.MessageCount > 0 {
		msgs = fmt.Sprintf("%*d", colMsgs, s.MessageCount)
	}

	modified := fmt.Sprintf("%-*s", colModified, formatRelativeTime(s.Modified))

	return fmt.Sprintf(" %s %-*s %-*s %-*s %-*s %s %s %s %s",
		status,
		colSummary, summaryCol,
		colProject, project,
		colBranch, branch,
		colPR, pr,
		tokens,
		cost,
		msgs,
		modified)
}

func formatCost(cost float64) string {
	if cost >= 1.0 {
		return fmt.Sprintf("$%.2f", cost)
	}
	return fmt.Sprintf("$%.4f", cost)
}

func renderSnippet(sr claude.SearchResult, maxWidth int) string {
	runes := []rune(sr.Snippet)
	if len(runes) <= maxWidth {
		before := string(runes[:sr.MatchPos])
		matchEnd := sr.MatchPos + sr.MatchLen
		if matchEnd > len(runes) {
			matchEnd = len(runes)
		}
		matched := string(runes[sr.MatchPos:matchEnd])
		after := string(runes[matchEnd:])
		return before + searchMatchStyle.Render(matched) + after
	}

	matchEnd := sr.MatchPos + sr.MatchLen
	if matchEnd > len(runes) {
		matchEnd = len(runes)
	}

	if matchEnd <= maxWidth-3 {
		before := string(runes[:sr.MatchPos])
		matched := string(runes[sr.MatchPos:matchEnd])
		after := string(runes[matchEnd : maxWidth-3])
		return before + searchMatchStyle.Render(matched) + after + "..."
	}

	start := sr.MatchPos - (maxWidth-sr.MatchLen-6)/2
	if start < 0 {
		start = 0
	}
	end := start + maxWidth - 6
	if end > len(runes) {
		end = len(runes)
		start = max(end-maxWidth+3, 0)
	}

	newMatchPos := sr.MatchPos - start
	newMatchEnd := matchEnd - start
	if newMatchEnd > end-start {
		newMatchEnd = end - start
	}

	sub := runes[start:end]
	before := string(sub[:newMatchPos])
	matched := string(sub[newMatchPos:newMatchEnd])
	after := string(sub[newMatchEnd:])

	prefix := ""
	suffix := ""
	if start > 0 {
		prefix = "..."
	}
	if end < len(runes) {
		suffix = "..."
	}
	return prefix + before + searchMatchStyle.Render(matched) + after + suffix
}

func colorizeTokens(tokens int64, width int) string {
	text := fmt.Sprintf("%*s", width, formatTokenCount(tokens))
	style := tokensLowStyle
	switch {
	case tokens >= 500_000:
		style = tokensHighStyle
	case tokens >= 100_000:
		style = tokensMedStyle
	}
	return style.Render(text)
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	return fmt.Sprintf("%-*s", width, s)
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
