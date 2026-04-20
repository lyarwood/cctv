package claude

import "time"

type SessionSource int

const (
	SourceIndex SessionSource = iota
	SourceJSONL
	SourceBoth
)

type Session struct {
	SessionID    string
	Summary      string
	FirstPrompt  string
	ProjectPath  string
	GitBranch    string
	MessageCount int
	Created      time.Time
	Modified     time.Time
	IsSidechain  bool
	IsRunning    bool
	RunningPID   int
	FullPath     string
	HasJSONL     bool
	Source       SessionSource
	PRLinks      []PRLink
	TotalTokens  int64
}

type PRLink struct {
	Number     int    `json:"prNumber"`
	URL        string `json:"prUrl"`
	Repository string `json:"prRepository"`
}

type SessionDetail struct {
	Models     []string
	TotalUsage TokenUsage
	LastPrompt string
	Version    string
	Prompts    []PromptEntry
	PRLinks    []PRLink
}

type TokenUsage struct {
	InputTokens              int64
	OutputTokens             int64
	CacheCreationInputTokens int64
	CacheReadInputTokens     int64
}

type PromptEntry struct {
	Content   string
	Timestamp time.Time
}

type IndexFile struct {
	Version int          `json:"version"`
	Entries []IndexEntry `json:"entries"`
}

type IndexEntry struct {
	SessionID    string `json:"sessionId"`
	FullPath     string `json:"fullPath"`
	FileMtime    int64  `json:"fileMtime"`
	FirstPrompt  string `json:"firstPrompt"`
	Summary      string `json:"summary"`
	MessageCount int    `json:"messageCount"`
	Created      string `json:"created"`
	Modified     string `json:"modified"`
	GitBranch    string `json:"gitBranch"`
	ProjectPath  string `json:"projectPath"`
	IsSidechain  bool   `json:"isSidechain"`
}

type RunningSession struct {
	PID        int    `json:"pid"`
	SessionID  string `json:"sessionId"`
	CWD        string `json:"cwd"`
	StartedAt  int64  `json:"startedAt"`
	Kind       string `json:"kind"`
	Entrypoint string `json:"entrypoint"`
}

type jsonlLine struct {
	Type string `json:"type"`
}

type jsonlUser struct {
	Type      string      `json:"type"`
	Message   userMessage `json:"message"`
	Timestamp string      `json:"timestamp"`
	CWD       string      `json:"cwd"`
	SessionID string      `json:"sessionId"`
	Version   string      `json:"version"`
	GitBranch string      `json:"gitBranch"`
}

type userMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

type jsonlAssistant struct {
	Type    string           `json:"type"`
	Message assistantMessage `json:"message"`
}

type assistantMessage struct {
	Model string     `json:"model"`
	Usage usageBlock `json:"usage"`
}

type usageBlock struct {
	InputTokens              int64 `json:"input_tokens"`
	OutputTokens             int64 `json:"output_tokens"`
	CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
}

type jsonlLastPrompt struct {
	Type       string `json:"type"`
	LastPrompt string `json:"lastPrompt"`
	SessionID  string `json:"sessionId"`
}

type jsonlPRLink struct {
	Type       string `json:"type"`
	Number     int    `json:"prNumber"`
	URL        string `json:"prUrl"`
	Repository string `json:"prRepository"`
}
