package claude

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

func ParseIndex(path string) ([]Session, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading index %s: %w", path, err)
	}

	var idx IndexFile
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("parsing index %s: %w", path, err)
	}

	var sessions []Session
	for _, e := range idx.Entries {
		if e.IsSidechain {
			continue
		}

		created, _ := time.Parse(time.RFC3339Nano, e.Created)
		modified, _ := time.Parse(time.RFC3339Nano, e.Modified)

		_, jsonlErr := os.Stat(e.FullPath)

		sessions = append(sessions, Session{
			SessionID:    e.SessionID,
			Summary:      e.Summary,
			FirstPrompt:  e.FirstPrompt,
			ProjectPath:  e.ProjectPath,
			GitBranch:    e.GitBranch,
			MessageCount: e.MessageCount,
			Created:      created,
			Modified:     modified,
			FullPath:     e.FullPath,
			HasJSONL:     jsonlErr == nil,
			Source:        SourceIndex,
		})
	}

	return sessions, nil
}
