package claude

import (
	"bufio"
	"encoding/json"
	"os"
	"regexp"
	"strings"
)

type jsonlSearchMessage struct {
	Type    string           `json:"type"`
	Message searchMsgContent `json:"message"`
}

type searchMsgContent struct {
	Content any `json:"content"`
}

func SearchJSONL(path string, pattern *regexp.Regexp) *SearchResult {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		var base jsonlLine
		if err := json.Unmarshal(line, &base); err != nil {
			continue
		}

		if base.Type != "user" && base.Type != "assistant" {
			continue
		}

		var msg jsonlSearchMessage
		if err := json.Unmarshal(line, &msg); err != nil {
			continue
		}

		for _, text := range extractAllContent(msg.Message.Content) {
			loc := pattern.FindStringIndex(text)
			if loc == nil {
				continue
			}
			snippet, pos, matchLen := buildSnippet(text, loc[0], loc[1]-loc[0], 60)
			return &SearchResult{
				Snippet:  snippet,
				MatchPos: pos,
				MatchLen: matchLen,
			}
		}
	}

	return nil
}

func (d *Discoverer) SearchAll(sessions []Session, query string, maxResults int) ([]SearchResult, error) {
	pattern, err := regexp.Compile("(?i)" + query)
	if err != nil {
		pattern = regexp.MustCompile("(?i)" + regexp.QuoteMeta(query))
	}

	var results []SearchResult
	for _, s := range sessions {
		if !s.HasJSONL {
			continue
		}
		if r := SearchJSONL(s.FullPath, pattern); r != nil {
			r.Session = s
			results = append(results, *r)
			if len(results) >= maxResults {
				break
			}
		}
	}
	return results, nil
}

func extractAllContent(content any) []string {
	switch v := content.(type) {
	case string:
		return []string{v}
	case []any:
		var texts []string
		for _, block := range v {
			if m, ok := block.(map[string]any); ok {
				if t, ok := m["type"].(string); ok && t == "text" {
					if text, ok := m["text"].(string); ok {
						texts = append(texts, text)
					}
				}
			}
		}
		return texts
	}
	return nil
}

func buildSnippet(text string, matchStart, matchLen, context int) (snippet string, snippetMatchPos, snippetMatchLen int) {
	runes := []rune(text)

	runeMatchStart := len([]rune(text[:matchStart]))
	runeMatchLen := len([]rune(text[matchStart : matchStart+matchLen]))

	snippetStart := max(runeMatchStart-context, 0)
	snippetEnd := min(runeMatchStart+runeMatchLen+context, len(runes))

	var b strings.Builder
	if snippetStart > 0 {
		b.WriteString("...")
	}
	b.WriteString(string(runes[snippetStart:snippetEnd]))
	if snippetEnd < len(runes) {
		b.WriteString("...")
	}

	snippetMatchPos = runeMatchStart - snippetStart
	if snippetStart > 0 {
		snippetMatchPos += 3
	}

	return b.String(), snippetMatchPos, runeMatchLen
}
