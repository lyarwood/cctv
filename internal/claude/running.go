package claude

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

func LoadRunning(claudeDir string) (map[string]RunningSession, error) {
	sessionsDir := filepath.Join(claudeDir, "sessions")
	entries, err := os.ReadDir(sessionsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading sessions dir: %w", err)
	}

	result := make(map[string]RunningSession)
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(sessionsDir, entry.Name()))
		if err != nil {
			continue
		}

		var rs RunningSession
		if err := json.Unmarshal(data, &rs); err != nil {
			continue
		}

		if !isProcessAlive(rs.PID) {
			continue
		}

		result[rs.SessionID] = rs
	}

	return result, nil
}

func isProcessAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil || err == syscall.EPERM
}
