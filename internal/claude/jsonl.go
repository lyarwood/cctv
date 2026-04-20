package claude

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

func ParseJSONLMetadata(path string) (*Session, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening JSONL %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat JSONL %s: %w", path, err)
	}

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

	var session Session
	session.FullPath = path
	session.HasJSONL = true
	session.Source = SourceJSONL
	session.Modified = info.ModTime()

	foundUser := false
	prSeen := make(map[string]bool)

	for scanner.Scan() {
		line := scanner.Bytes()
		var base jsonlLine
		if err := json.Unmarshal(line, &base); err != nil {
			continue
		}

		switch base.Type {
		case "user":
			if !foundUser {
				var u jsonlUser
				if err := json.Unmarshal(line, &u); err != nil {
					continue
				}
				session.SessionID = u.SessionID
				session.FirstPrompt = extractContent(u.Message.Content)
				session.ProjectPath = u.CWD
				session.GitBranch = u.GitBranch
				if ts, err := time.Parse(time.RFC3339Nano, u.Timestamp); err == nil {
					session.Created = ts
				}
				foundUser = true
			}
		case "assistant":
			var a jsonlAssistant
			if err := json.Unmarshal(line, &a); err != nil {
				continue
			}
			session.TotalTokens += a.Message.Usage.InputTokens + a.Message.Usage.OutputTokens
		case "pr-link":
			var pr jsonlPRLink
			if err := json.Unmarshal(line, &pr); err != nil {
				continue
			}
			key := fmt.Sprintf("%s#%d", pr.Repository, pr.Number)
			if !prSeen[key] {
				prSeen[key] = true
				session.PRLinks = append(session.PRLinks, PRLink{
					Number:     pr.Number,
					URL:        pr.URL,
					Repository: pr.Repository,
				})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning JSONL %s: %w", path, err)
	}

	if session.SessionID == "" {
		return nil, fmt.Errorf("no user message found in %s", path)
	}

	return &session, nil
}

func ParseJSONLDetail(path string) (*SessionDetail, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening JSONL %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

	detail := &SessionDetail{}
	modelSet := make(map[string]bool)
	prSeen := make(map[string]bool)

	for scanner.Scan() {
		line := scanner.Bytes()
		var base jsonlLine
		if err := json.Unmarshal(line, &base); err != nil {
			continue
		}

		switch base.Type {
		case "user":
			var u jsonlUser
			if err := json.Unmarshal(line, &u); err != nil {
				continue
			}
			if detail.Version == "" {
				detail.Version = u.Version
			}
			ts, _ := time.Parse(time.RFC3339Nano, u.Timestamp)
			detail.Prompts = append(detail.Prompts, PromptEntry{
				Content:   extractContent(u.Message.Content),
				Timestamp: ts,
			})

		case "assistant":
			var a jsonlAssistant
			if err := json.Unmarshal(line, &a); err != nil {
				continue
			}
			if a.Message.Model != "" && !modelSet[a.Message.Model] {
				modelSet[a.Message.Model] = true
				detail.Models = append(detail.Models, a.Message.Model)
			}
			detail.TotalUsage.InputTokens += a.Message.Usage.InputTokens
			detail.TotalUsage.OutputTokens += a.Message.Usage.OutputTokens
			detail.TotalUsage.CacheCreationInputTokens += a.Message.Usage.CacheCreationInputTokens
			detail.TotalUsage.CacheReadInputTokens += a.Message.Usage.CacheReadInputTokens

		case "last-prompt":
			var lp jsonlLastPrompt
			if err := json.Unmarshal(line, &lp); err != nil {
				continue
			}
			detail.LastPrompt = lp.LastPrompt

		case "pr-link":
			var pr jsonlPRLink
			if err := json.Unmarshal(line, &pr); err != nil {
				continue
			}
			key := fmt.Sprintf("%s#%d", pr.Repository, pr.Number)
			if !prSeen[key] {
				prSeen[key] = true
				detail.PRLinks = append(detail.PRLinks, PRLink{
					Number:     pr.Number,
					URL:        pr.URL,
					Repository: pr.Repository,
				})
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning JSONL %s: %w", path, err)
	}

	return detail, nil
}

func extractContent(content any) string {
	switch v := content.(type) {
	case string:
		return v
	case []any:
		for _, block := range v {
			if m, ok := block.(map[string]any); ok {
				if t, ok := m["type"].(string); ok && t == "text" {
					if text, ok := m["text"].(string); ok {
						return text
					}
				}
			}
		}
	}
	return ""
}
