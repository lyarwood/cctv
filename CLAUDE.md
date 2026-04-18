# CLAUDE.md

## Project overview

cctv is a Go TUI for browsing and resuming Claude Code conversations from the local filesystem (`~/.claude/`). It uses Charm libraries (Bubble Tea, Lip Gloss, Bubbles) for the TUI, Cobra for CLI argument parsing, and Ginkgo/Gomega for testing.

## Build and test

```sh
make build          # build to bin/cctv
make test           # run tests via ginkgo
make lint           # run golangci-lint
make test-cover     # coverage report to coverage.html
```

## Architecture

- `internal/claude/` — data layer: parses `~/.claude/projects/*/sessions-index.json` (fast metadata), `*.jsonl` (full conversations with PR links, token usage, models), and `sessions/<pid>.json` (running session detection). The `Discoverer` merges all sources into a unified `[]Session` sorted by modified time.
- `internal/tui/` — Bubble Tea TUI with three views (list, detail, and stats popup). The stats popup (`s` key) overlays the list with token usage, cache hit rate, and session duration. Filtering is live (updates on every keystroke) and supports prefix syntax (`project:`, `branch:`, `cwd:`, `pr:`). Multiple space-separated terms are ANDed.
- `internal/cmd/` — Cobra commands: root (launches TUI), `list` (non-interactive), `resume` (exec into claude), `version`.
- `internal/claude/sanitize.go` — cleans raw prompts: extracts slash command names from XML tags, replaces `<local-command-caveat>` with `[local command]`, shortens URLs, strips remaining XML.

## Testing conventions

- Use Ginkgo v2 (`github.com/onsi/ginkgo/v2`) and Gomega for tests in `internal/claude/`.
- Standard Go `testing` package for tests in `internal/tui/`.
- Test fixtures live in `internal/claude/testdata/`.
- Each Ginkgo test package has a `*_suite_test.go` bootstrap file.

## Key design decisions

- `tea.ExecProcess` suspends the TUI when resuming a session, returning to cctv after Claude exits. The command's `Dir` is set to `session.ProjectPath`.
- JSONL parsing uses `bufio.Scanner` line-by-line for memory efficiency. `ParseJSONLMetadata` scans the full file to collect PR links. `ParseJSONLDetail` is loaded on-demand for the detail view.
- PR links are deduplicated by `repo#number`.
- Sidechain sessions (subagent conversations) are filtered out.
- Version is injected at build time via `-ldflags`.
