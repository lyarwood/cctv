package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/spf13/cobra"
)

var resumeCmd = &cobra.Command{
	Use:   "resume <session-id>",
	Short: "Resume a Claude Code session by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sessionID := args[0]
		claudeBin, err := exec.LookPath("claude")
		if err != nil {
			return fmt.Errorf("claude not found in PATH: %w", err)
		}

		return syscall.Exec(claudeBin, []string{"claude", "--resume", sessionID}, os.Environ())
	},
}
