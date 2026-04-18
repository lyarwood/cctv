# CLAUDE.md

## Project overview

cctv is a Go TUI for browsing and resuming Claude Code conversations from the local filesystem (`~/.claude/`). It uses Charm libraries (Bubble Tea, Lip Gloss, Bubbles) for the TUI, Cobra for CLI argument parsing, and Ginkgo/Gomega for testing.

## Build and test

```sh
make build          # build to bin/cctv
make test           # run tests via ginkgo
make lint           # run golangci-lint
make test-cover     # coverage report to coverage.html
make demo           # regenerate demo.gif (requires vhs)
```

## Architecture

- `internal/claude/` — data layer: parses `~/.claude/projects/*/sessions-index.json` (fast metadata), `*.jsonl` (full conversations with PR links, token usage, models), and `sessions/<pid>.json` (running session detection). The `Discoverer` merges all sources into a unified `[]Session` sorted by modified time.
- `internal/tui/` — Bubble Tea TUI with three views (list, detail, and stats popup). The stats popup (`s` key) shows token usage, cache hit rate, and session duration. Filtering is live (updates on every keystroke) with regex support and prefix syntax (`project:`, `branch:`, `cwd:`, `pr:`). Multiple space-separated terms are ANDed. Invalid regex falls back to substring matching. Theming is in `theme.go` — a `Theme` struct defines named colors, `styles.go` derives all Lip Gloss styles from the active theme via `applyTheme()`. 5 built-in themes: default, catppuccin, dracula, nord, light. Selected via `--theme` flag.
- `internal/cmd/` — Cobra commands: root (launches TUI), `list` (non-interactive), `resume` (exec into claude), `version`.
- `internal/claude/sanitize.go` — cleans raw prompts: extracts slash command names from XML tags, replaces `<local-command-caveat>` with `[local command]`, shortens URLs, strips remaining XML.

## Testing conventions

- Use Ginkgo v2 (`github.com/onsi/ginkgo/v2`) and Gomega for all tests.
- Tests use external test packages (`package claude_test`, `package tui_test`) to test only public APIs.
- Test fixtures live in `internal/claude/testdata/`.
- Each Ginkgo test package has a `*_suite_test.go` bootstrap file.

## Key design decisions

- `tea.ExecProcess` suspends the TUI when resuming a session, returning to cctv after Claude exits. The command's `Dir` is set to `session.ProjectPath` if the directory exists, otherwise it inherits the current directory.
- JSONL parsing uses `bufio.Scanner` line-by-line for memory efficiency. `ParseJSONLMetadata` scans the full file to collect PR links. `ParseJSONLDetail` is loaded on-demand for the detail view.
- PR links are deduplicated by `repo#number`. When a session exists in both the index and as a JSONL file, discovery parses the JSONL to extract PR links.
- Filter values are compiled as case-insensitive regex patterns. Invalid patterns fall back to substring matching.
- Sidechain sessions (subagent conversations) are filtered out.
- Version is injected at build time via `-ldflags`.
- Adding a new theme: add an entry to the `themes` map in `theme.go` — all styles are automatically derived.
- `demo/` contains fake session data and a VHS tape for reproducible demo GIF generation via `make demo`.
