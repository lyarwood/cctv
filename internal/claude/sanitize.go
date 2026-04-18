package claude

import (
	"net/url"
	"regexp"
	"strings"
)

var (
	reCommandName = regexp.MustCompile(`<command-name>\s*/?([\w:/-]+)\s*</command-name>`)
	reCommandArgs = regexp.MustCompile(`<command-args>(.*?)</command-args>`)
	reLocalCaveat = regexp.MustCompile(`(?s)<local-command-caveat>.*?</local-command-caveat>`)
	reBashInput   = regexp.MustCompile(`<bash-input>(.*?)</bash-input>`)
	reXMLTags     = regexp.MustCompile(`<[^>]+>`)
	reWhitespace  = regexp.MustCompile(`\s+`)
	reURL         = regexp.MustCompile(`https?://\S+`)
)

func SanitizePrompt(raw string) string {
	if raw == "" {
		return ""
	}

	if reCommandName.MatchString(raw) {
		return extractCommand(raw)
	}

	if strings.Contains(raw, "<local-command-caveat>") {
		after := reLocalCaveat.ReplaceAllString(raw, "")
		after = strings.TrimSpace(after)
		if after != "" {
			return sanitizeText(after)
		}
		return "[local command]"
	}

	if reBashInput.MatchString(raw) {
		if m := reBashInput.FindStringSubmatch(raw); len(m) > 1 {
			return "$ " + strings.TrimSpace(m[1])
		}
	}

	return sanitizeText(raw)
}

func extractCommand(raw string) string {
	name := ""
	if m := reCommandName.FindStringSubmatch(raw); len(m) > 1 {
		name = "/" + strings.TrimLeft(m[1], "/")
	}

	args := ""
	if m := reCommandArgs.FindStringSubmatch(raw); len(m) > 1 {
		args = strings.TrimSpace(m[1])
		args = shortenURLs(args)
	}

	if args != "" {
		return name + " " + args
	}
	return name
}

func sanitizeText(s string) string {
	s = reXMLTags.ReplaceAllString(s, "")
	s = shortenURLs(s)
	s = reWhitespace.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)
	return s
}

func shortenURLs(s string) string {
	return reURL.ReplaceAllStringFunc(s, func(raw string) string {
		u, err := url.Parse(raw)
		if err != nil {
			return raw
		}
		path := u.Path
		if len(path) > 40 {
			parts := strings.Split(strings.Trim(path, "/"), "/")
			if len(parts) > 3 {
				path = "/" + strings.Join(parts[:2], "/") + "/.../" + parts[len(parts)-1]
			}
		}
		short := u.Host + path
		if short == "" {
			return raw
		}
		return short
	})
}
