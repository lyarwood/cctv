package claude

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Discoverer struct {
	ClaudeDir string
}

func NewDiscoverer(claudeDir string) *Discoverer {
	if claudeDir == "" {
		home, _ := os.UserHomeDir()
		claudeDir = filepath.Join(home, ".claude")
	}
	return &Discoverer{ClaudeDir: claudeDir}
}

func (d *Discoverer) DiscoverAll() ([]Session, error) {
	projectsDir := filepath.Join(d.ClaudeDir, "projects")
	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return nil, fmt.Errorf("reading projects dir: %w", err)
	}

	running, err := LoadRunning(d.ClaudeDir)
	if err != nil {
		running = make(map[string]RunningSession)
	}

	var all []Session
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		projectDir := filepath.Join(projectsDir, entry.Name())
		sessions, err := d.discoverProject(projectDir)
		if err != nil {
			continue
		}
		all = append(all, sessions...)
	}

	for i := range all {
		if rs, ok := running[all[i].SessionID]; ok {
			all[i].IsRunning = true
			all[i].RunningPID = rs.PID
		}
	}

	sort.Slice(all, func(i, j int) bool {
		return all[i].Modified.After(all[j].Modified)
	})

	return all, nil
}

func (d *Discoverer) discoverProject(projectDir string) ([]Session, error) {
	indexed := make(map[string]bool)
	var sessions []Session

	indexPath := filepath.Join(projectDir, "sessions-index.json")
	if indexSessions, err := ParseIndex(indexPath); err == nil {
		for _, s := range indexSessions {
			indexed[s.SessionID] = true
			sessions = append(sessions, s)
		}
	}

	jsonlFiles, _ := filepath.Glob(filepath.Join(projectDir, "*.jsonl"))
	for _, path := range jsonlFiles {
		sessionID := strings.TrimSuffix(filepath.Base(path), ".jsonl")
		if indexed[sessionID] {
			for i := range sessions {
				if sessions[i].SessionID == sessionID {
					sessions[i].HasJSONL = true
					sessions[i].Source = SourceBoth
					break
				}
			}
			continue
		}

		if s, err := ParseJSONLMetadata(path); err == nil {
			sessions = append(sessions, *s)
		}
	}

	return sessions, nil
}

func (d *Discoverer) LoadDetail(session Session) (*SessionDetail, error) {
	if !session.HasJSONL {
		return nil, fmt.Errorf("no JSONL file for session %s", session.SessionID)
	}
	return ParseJSONLDetail(session.FullPath)
}
