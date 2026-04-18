package cmd

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/lyarwood/cctv/internal/claude"
	"github.com/lyarwood/cctv/internal/tui"
)

var Version = "dev"

var (
	claudeDir string
	usePWD    bool
)

var rootCmd = &cobra.Command{
	Use:   "cctv",
	Short: "Claude Code TUI Viewer - browse and resume Claude Code conversations",
	Long:  "cctv discovers Claude Code conversations from the local filesystem and displays them in a TUI. Select a session to resume it.",
	RunE: func(cmd *cobra.Command, args []string) error {
		discoverer := claude.NewDiscoverer(claudeDir)
		sessions, err := discoverer.DiscoverAll()
		if err != nil {
			return fmt.Errorf("discovering sessions: %w", err)
		}

		initialFilter := ""
		if usePWD {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("getting working directory: %w", err)
			}
			initialFilter = "cwd:" + cwd
			sessions = filterByCWD(sessions, cwd)
		}

		model := tui.NewModel(sessions, discoverer, initialFilter)
		p := tea.NewProgram(model, tea.WithAltScreen())
		_, err = p.Run()
		return err
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&claudeDir, "claude-dir", "", "path to Claude Code data directory (default: ~/.claude)")
	rootCmd.Flags().BoolVar(&usePWD, "pwd", false, "filter sessions by present working directory")
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(resumeCmd)
	rootCmd.AddCommand(versionCmd)
}

func filterByCWD(sessions []claude.Session, cwd string) []claude.Session {
	lower := strings.ToLower(cwd)
	var filtered []claude.Session
	for _, s := range sessions {
		if strings.Contains(strings.ToLower(s.ProjectPath), lower) {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
