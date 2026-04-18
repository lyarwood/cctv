package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/lyarwood/cctv/internal/claude"
)

var (
	statsPopupStyle = lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("205")).
			Padding(1, 3).
			Background(lipgloss.Color("235"))

	statsSectionStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("205")).
				MarginTop(1)

	statsLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Width(20)

	statsValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	statsBarFill = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42"))

	statsBarEmpty = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))
)

func renderStats(session claude.Session, detail *claude.SessionDetail, width, height int) string {
	var b strings.Builder

	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).Render("Session Stats"))
	b.WriteString("\n\n")

	row := func(label, value string) {
		b.WriteString(statsLabelStyle.Render(label))
		b.WriteString(statsValueStyle.Render(value))
		b.WriteString("\n")
	}

	summary := session.Summary
	if summary == "" {
		summary = claude.SanitizePrompt(session.FirstPrompt)
	}
	row("Session", truncateStr(summary, 40))

	if !session.Created.IsZero() && !session.Modified.IsZero() {
		duration := session.Modified.Sub(session.Created)
		row("Duration", formatDuration(duration))
	}

	if session.MessageCount > 0 {
		row("Messages", fmt.Sprintf("%d", session.MessageCount))
	}

	if len(session.PRLinks) > 0 {
		for i, pr := range session.PRLinks {
			label := "PR"
			if i > 0 {
				label = ""
			}
			row(label, fmt.Sprintf("%s#%d", pr.Repository, pr.Number))
		}
	}

	if detail != nil {
		if len(detail.Prompts) > 0 {
			row("User Prompts", fmt.Sprintf("%d", len(detail.Prompts)))
		}

		if len(detail.Models) > 0 {
			row("Models", strings.Join(detail.Models, ", "))
		}
		if detail.Version != "" {
			row("CC Version", detail.Version)
		}

		u := detail.TotalUsage
		totalTokens := u.InputTokens + u.OutputTokens
		if totalTokens > 0 {
			b.WriteString("\n")
			b.WriteString(statsSectionStyle.Render("Token Usage"))
			b.WriteString("\n")

			row("Input", formatTokens(u.InputTokens))
			row("Output", formatTokens(u.OutputTokens))
			row("Total", formatTokens(totalTokens))

			if u.CacheCreationInputTokens > 0 || u.CacheReadInputTokens > 0 {
				b.WriteString("\n")
				b.WriteString(statsSectionStyle.Render("Cache"))
				b.WriteString("\n")

				row("Cache Write", formatTokens(u.CacheCreationInputTokens))
				row("Cache Read", formatTokens(u.CacheReadInputTokens))

				totalCacheInput := u.CacheCreationInputTokens + u.CacheReadInputTokens
				if totalCacheInput > 0 {
					hitRate := float64(u.CacheReadInputTokens) / float64(totalCacheInput) * 100
					row("Hit Rate", fmt.Sprintf("%.1f%%", hitRate))
					b.WriteString(statsLabelStyle.Render(""))
					b.WriteString(renderBar(hitRate, 30))
					b.WriteString("\n")
				}
			}

			b.WriteString("\n")
			b.WriteString(statsSectionStyle.Render("Ratio"))
			b.WriteString("\n")
			outputRatio := float64(u.OutputTokens) / float64(totalTokens) * 100
			row("Output/Total", fmt.Sprintf("%.1f%%", outputRatio))
			b.WriteString(statsLabelStyle.Render(""))
			b.WriteString(renderBar(outputRatio, 30))
			b.WriteString("\n")
		}
	} else if session.HasJSONL {
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("Loading stats..."))
	} else {
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("No JSONL file available"))
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("esc:close  d:detail  enter:resume"))

	popupWidth := width/2 - 2
	if popupWidth < 50 {
		popupWidth = 50
	}

	return statsPopupStyle.
		Width(popupWidth).
		Render(b.String())
}

func renderBar(pct float64, width int) string {
	filled := int(pct / 100 * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled
	return statsBarFill.Render(strings.Repeat("█", filled)) +
		statsBarEmpty.Render(strings.Repeat("░", empty))
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm %ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	return fmt.Sprintf("%dd %dh", days, hours)
}

func overlayCenter(bg, popup string, width, height int) string {
	bgLines := strings.Split(bg, "\n")
	popupLines := strings.Split(popup, "\n")

	popupHeight := len(popupLines)
	popupWidth := 0
	for _, line := range popupLines {
		if w := lipgloss.Width(line); w > popupWidth {
			popupWidth = w
		}
	}

	startRow := (height - popupHeight) / 2
	if startRow < 0 {
		startRow = 0
	}
	startCol := (width - popupWidth) / 2
	if startCol < 0 {
		startCol = 0
	}

	for len(bgLines) < height {
		bgLines = append(bgLines, strings.Repeat(" ", width))
	}

	for i, popupLine := range popupLines {
		row := startRow + i
		if row >= len(bgLines) {
			break
		}
		bgLine := bgLines[row]
		bgRunes := []rune(bgLine)

		for len(bgRunes) < width {
			bgRunes = append(bgRunes, ' ')
		}

		prefix := string(bgRunes[:startCol])
		popupEnd := startCol + popupWidth
		suffix := ""
		if popupEnd < len(bgRunes) {
			suffix = string(bgRunes[popupEnd:])
		}

		bgLines[row] = prefix + popupLine + suffix
	}

	return strings.Join(bgLines[:height], "\n")
}
