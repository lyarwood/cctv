package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"github.com/lyarwood/cctv/internal/claude"
)

var (
	listJSON    bool
	listProject string
	listBranch  string
	listCWD     string
	listPWD     bool
	listPR      string
	listLimit   int
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List sessions non-interactively",
	RunE: func(cmd *cobra.Command, args []string) error {
		if listPWD {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}
			listCWD = cwd
		}

		discoverer := claude.NewDiscoverer(claudeDir)
		sessions, err := discoverer.DiscoverAll()
		if err != nil {
			return fmt.Errorf("discovering sessions: %w", err)
		}

		sessions = filterSessions(sessions)

		if listLimit > 0 && len(sessions) > listLimit {
			sessions = sessions[:listLimit]
		}

		if listJSON {
			return printJSON(sessions)
		}
		return printTable(sessions)
	},
}

func init() {
	listCmd.Flags().BoolVar(&listJSON, "json", false, "output as JSON")
	listCmd.Flags().StringVar(&listProject, "project", "", "filter by project path or name")
	listCmd.Flags().StringVar(&listBranch, "branch", "", "filter by git branch")
	listCmd.Flags().StringVar(&listCWD, "cwd", "", "filter by working directory path")
	listCmd.Flags().BoolVar(&listPWD, "pwd", false, "filter sessions by present working directory")
	listCmd.Flags().StringVar(&listPR, "pr", "", "filter by PR repository or number")
	listCmd.Flags().IntVar(&listLimit, "limit", 0, "maximum number of sessions to show")
}

func filterSessions(sessions []claude.Session) []claude.Session {
	if listProject == "" && listBranch == "" && listCWD == "" && listPR == "" {
		return sessions
	}
	var filtered []claude.Session
	for _, s := range sessions {
		if listProject != "" {
			if !flexMatch(filepath.Base(s.ProjectPath), listProject) &&
				!flexMatch(s.ProjectPath, listProject) {
				continue
			}
		}
		if listBranch != "" {
			if !flexMatch(s.GitBranch, listBranch) {
				continue
			}
		}
		if listCWD != "" {
			if !flexMatch(s.ProjectPath, listCWD) {
				continue
			}
		}
		if listPR != "" {
			if !matchPRLinks(s.PRLinks, listPR) {
				continue
			}
		}
		filtered = append(filtered, s)
	}
	return filtered
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

func printJSON(sessions []claude.Session) error {
	type jsonPR struct {
		Number     int    `json:"number"`
		URL        string `json:"url"`
		Repository string `json:"repository"`
	}
	type jsonSession struct {
		SessionID    string   `json:"sessionId"`
		Summary      string   `json:"summary"`
		FirstPrompt  string   `json:"firstPrompt"`
		ProjectPath  string   `json:"projectPath"`
		GitBranch    string   `json:"gitBranch"`
		MessageCount int      `json:"messageCount"`
		Created      string   `json:"created"`
		Modified     string   `json:"modified"`
		IsRunning    bool     `json:"isRunning"`
		PRLinks      []jsonPR `json:"prLinks,omitempty"`
	}

	out := make([]jsonSession, len(sessions))
	for i, s := range sessions {
		out[i] = jsonSession{
			SessionID:    s.SessionID,
			Summary:      s.Summary,
			FirstPrompt:  s.FirstPrompt,
			ProjectPath:  s.ProjectPath,
			GitBranch:    s.GitBranch,
			MessageCount: s.MessageCount,
			Created:      s.Created.Format(time.RFC3339),
			Modified:     s.Modified.Format(time.RFC3339),
			IsRunning:    s.IsRunning,
		}
		for _, pr := range s.PRLinks {
			out[i].PRLinks = append(out[i].PRLinks, jsonPR{
				Number:     pr.Number,
				URL:        pr.URL,
				Repository: pr.Repository,
			})
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func printTable(sessions []claude.Session) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "STATUS\tSUMMARY\tPROJECT\tBRANCH\tPR\tMSGS\tMODIFIED")

	for _, s := range sessions {
		status := " "
		if s.IsRunning {
			status = "*"
		}
		summary := s.Summary
		if summary == "" {
			summary = truncate(claude.SanitizePrompt(s.FirstPrompt), 50)
		}
		project := filepath.Base(s.ProjectPath)
		modified := relativeTime(s.Modified)
		pr := formatCLIPRLinks(s.PRLinks)

		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%d\t%s\n",
			status, summary, project, s.GitBranch, pr, s.MessageCount, modified)
	}
	return w.Flush()
}

func formatCLIPRLinks(links []claude.PRLink) string {
	if len(links) == 0 {
		return ""
	}
	first := fmt.Sprintf("%s#%d", links[0].Repository, links[0].Number)
	if len(links) > 1 {
		return fmt.Sprintf("%s +%d", first, len(links)-1)
	}
	return first
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func relativeTime(t time.Time) string {
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
