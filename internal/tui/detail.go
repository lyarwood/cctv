package tui

import (
	"fmt"
	"strings"

	"github.com/lyarwood/cctv/internal/claude"
)

func renderDetail(session claude.Session, detail *claude.SessionDetail, width, height int) string {
	var b strings.Builder

	row := func(label, value string) {
		b.WriteString(detailLabelStyle.Render(label))
		b.WriteString(detailValueStyle.Render(value))
		b.WriteString("\n")
	}

	row("Session ID", session.SessionID)
	b.WriteString("\n")

	if session.Summary != "" {
		row("Summary", session.Summary)
	}

	row("First Prompt", claude.SanitizePrompt(session.FirstPrompt))
	b.WriteString("\n")

	row("Project", session.ProjectPath)
	row("Branch", session.GitBranch)

	if !session.Created.IsZero() {
		row("Created", session.Created.Format("2006-01-02 15:04:05"))
	}
	if !session.Modified.IsZero() {
		row("Modified", session.Modified.Format("2006-01-02 15:04:05"))
	}

	if session.MessageCount > 0 {
		row("Messages", fmt.Sprintf("%d", session.MessageCount))
	}

	if session.IsRunning {
		row("Status", runningStyle.Render(fmt.Sprintf("Running (PID %d)", session.RunningPID)))
	}

	prLinks := session.PRLinks
	if detail != nil && len(detail.PRLinks) > 0 {
		prLinks = detail.PRLinks
	}
	if len(prLinks) > 0 {
		b.WriteString("\n")
		for _, pr := range prLinks {
			row("PR", fmt.Sprintf("%s#%d", pr.Repository, pr.Number))
			row("  URL", pr.URL)
		}
	}

	if detail != nil {
		b.WriteString("\n")
		if len(detail.Models) > 0 {
			row("Models", strings.Join(detail.Models, ", "))
		}
		if detail.Version != "" {
			row("CC Version", detail.Version)
		}
		if detail.TotalUsage.InputTokens > 0 || detail.TotalUsage.OutputTokens > 0 {
			b.WriteString("\n")
			row("Input Tokens", formatTokens(detail.TotalUsage.InputTokens))
			row("Output Tokens", formatTokens(detail.TotalUsage.OutputTokens))
			if detail.TotalUsage.CacheCreationInputTokens > 0 {
				row("Cache Write", formatTokens(detail.TotalUsage.CacheCreationInputTokens))
			}
			if detail.TotalUsage.CacheReadInputTokens > 0 {
				row("Cache Read", formatTokens(detail.TotalUsage.CacheReadInputTokens))
			}
			if detail.LastInputTokens > 0 && len(detail.Models) > 0 {
				model := detail.Models[len(detail.Models)-1]
				coldCost, _ := claude.EstimateResumeCost(model, detail.LastInputTokens)
				if coldCost > 0 {
					row("Resume Cost", fmt.Sprintf("~$%.4f (uncached)", coldCost))
				}
			}
		}

		if len(detail.Prompts) > 0 {
			b.WriteString("\n")
			b.WriteString(detailLabelStyle.Render("Prompts"))
			b.WriteString("\n")
			for i, p := range detail.Prompts {
				ts := ""
				if !p.Timestamp.IsZero() {
					ts = dimStyle.Render(p.Timestamp.Format("15:04:05")) + " "
				}
				prompt := truncateStr(p.Content, 80)
				fmt.Fprintf(&b, "  %d. %s%s\n", i+1, ts, prompt)
			}
		}
	} else if session.HasJSONL {
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("Loading details..."))
	} else {
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("No JSONL file available (archived session)"))
	}

	innerWidth := width - 6
	if innerWidth < 40 {
		innerWidth = 40
	}
	innerHeight := height - 4
	if innerHeight < 10 {
		innerHeight = 10
	}

	return detailBorderStyle.
		Width(innerWidth).
		Height(innerHeight).
		Render(b.String())
}

func formatTokens(n int64) string {
	if n >= 1_000_000 {
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	}
	if n >= 1_000 {
		return fmt.Sprintf("%.1fK", float64(n)/1_000)
	}
	return fmt.Sprintf("%d", n)
}
