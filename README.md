# cctv - Claude Code TUI Viewer

A terminal UI for browsing and resuming [Claude Code](https://docs.anthropic.com/en/docs/claude-code) conversations from the local filesystem.

![cctv demo](demo.gif)

## Why resume conversations?

Claude Code conversations accumulate valuable context as you work — the files you've discussed, decisions you've made, bugs you've investigated, and the mental model Claude has built of your codebase. Starting a fresh conversation throws all of that away, forcing you to re-explain context and watch Claude re-discover things it already knew.

Resuming a conversation lets you:

- **Continue where you left off** — pick up a debugging session, code review, or refactor without re-establishing context
- **Reduce token usage** — resumed sessions carry forward their prompt cache, avoiding the cost of Claude re-reading your codebase from scratch
- **Maintain coherence across sessions** — when you step away and come back hours or days later, the conversation still remembers what was agreed, what was tried, and what's left to do
- **Switch between workstreams** — jump between a PR review on one branch and a feature implementation on another, each with its own accumulated context

Claude Code's built-in `--resume` picker is functional but minimal. cctv gives you a richer view across all your projects — filterable by project, branch, PR, or working directory — so you can quickly find and resume the right conversation.

## Features

- Lists all Claude Code sessions across all projects with metadata (summary, project, branch, PR links, message count, timestamps, running status)
- Live filtering with regex support and prefix syntax (`project:`, `branch:`, `cwd:`, `pr:`)
- Content search (`\`) — search through JSONL conversation content for strings or regex, with match snippets highlighted
- Resume cost estimation — shows estimated API cost to resume each session based on context size, model, and prompt cache state (warm/cold)
- Stats popup with token usage, cache hit rate, session duration, model breakdown, and resume cost
- Detail view with model info, token usage, prompt history, PR details, and resume cost
- Resume any session directly from the TUI — suspends cctv, launches Claude Code, returns when done
- Non-interactive `list` subcommand with `--json` output for scripting
- Configurable model pricing via `~/.config/cctv/pricing.json` — override built-in rates for custom providers (e.g., Google Vertex AI, AWS Bedrock)
- 5 built-in color themes (`default`, `catppuccin`, `dracula`, `nord`, `light`) via `--theme`
- Sanitizes raw prompts — slash commands, local commands, and XML tags are cleaned up for readability

## Installation

```sh
make install
```

Or build locally:

```sh
make build
./bin/cctv
```

## Usage

### TUI (default)

```sh
cctv              # browse all sessions
cctv --pwd        # filter to sessions from the present working directory
```

#### Key bindings

| Key | Action |
|---|---|
| `enter` | Resume selected session |
| `d` / `space` | View session details |
| `s` | Session stats popup (token usage, cache hit rate, resume cost) |
| `/` | Open filter input |
| `\` | Search conversation content |
| `tab` | Cycle filter prefix (`project:`, `branch:`, `cwd:`, `pr:`) |
| `r` | Refresh session list |
| `?` | Toggle help |
| `esc` | Back / cancel filter |
| `q` / `ctrl+c` | Quit |

#### Filtering

Type bare text to search across all fields, or use a prefix to target a specific field. All filter values support regex:

```
project:kubevirt$            # exact project name (regex anchor)
project:kubevirt             # substring match (also valid regex)
branch:main                  # match git branch
cwd:/home/user/project       # match working directory path
pr:enhancements#242          # match PR repository or number
branch:^feature/             # branches starting with feature/
```

Multiple terms are ANDed together:

```
project:kubevirt branch:main
pr:242 branch:fix
```

Invalid regex patterns fall back to substring matching.

#### Resume cost estimation

The list view, detail view, and stats popup show an estimated API cost to resume each session. The estimate is based on the last turn's context size (input + cached tokens) and the model's per-million-token pricing. The stats popup also indicates whether the prompt cache is likely warm (modified < 5 minutes ago) or cold.

Built-in pricing covers Anthropic's direct API rates. To override for custom providers (e.g., Google Vertex AI with regional endpoint premium), create `~/.config/cctv/pricing.json`:

```json
{
  "claude-opus-4-6": {"input": 5.50, "cache_hit": 0.55},
  "claude-sonnet-4-6": {"input": 3.30, "cache_hit": 0.33},
  "claude-haiku-4-5": {"input": 1.10, "cache_hit": 0.11}
}
```

Entries merge into the built-in table — only list models you want to override. Prices are per million tokens. You can also pass `--pricing /path/to/pricing.json` to use a specific file.

#### Themes

```sh
cctv --theme catppuccin    # pastel purple/green
cctv --theme dracula       # pink/purple from Dracula
cctv --theme nord          # cool blue/green from Nord
cctv --theme light         # high contrast for light terminals
cctv --theme default       # the default theme
```

### CLI

```sh
cctv list                              # table output
cctv list --json                       # JSON output
cctv list --project kubevirt           # filter by project
cctv list --branch main                # filter by branch
cctv list --cwd /path/to/project       # filter by working directory
cctv list --pr enhancements            # filter by PR repo or number
cctv list --pwd                        # filter by present working directory
cctv list --limit 10                   # limit results
cctv resume <session-id>               # resume a session directly
cctv version                           # print version
```

## Data Sources

cctv reads Claude Code's local storage under `~/.claude/`:

| Source | Path | Content |
|---|---|---|
| Session index | `projects/*/sessions-index.json` | Fast metadata: summary, message count, timestamps |
| Session JSONL | `projects/*/<uuid>.jsonl` | Full conversation: messages, models, token usage, PR links |
| Running PIDs | `sessions/<pid>.json` | Active session detection with stale PID validation |

## Development

### Prerequisites

- Go 1.24+
- [Ginkgo](https://onsi.github.io/ginkgo/) test framework
- [golangci-lint](https://golangci-lint.run/)

### Makefile targets

| Target | Description |
|---|---|
| `make build` | Build binary to `bin/cctv` |
| `make test` | Run all tests via Ginkgo |
| `make test-verbose` | Run tests with verbose output |
| `make test-cover` | Run tests with coverage report (`coverage.html`) |
| `make lint` | Run golangci-lint |
| `make fmt` | Format code |
| `make vet` | Run go vet |
| `make clean` | Remove build artifacts |
| `make install` | Build and install to `$GOPATH/bin` |
| `make run` | Build and run |
| `make demo` | Regenerate `demo.gif` from `demo/` data and `demo.tape` (requires [vhs](https://github.com/charmbracelet/vhs)) |

### Project structure

```
cmd/cctv/main.go                  # Entrypoint
internal/
  claude/                         # Data model and parsing
    types.go                      # Session, PRLink, TokenUsage, etc.
    index.go                      # sessions-index.json parser
    jsonl.go                      # JSONL streaming parser
    running.go                    # Running session detection
    discovery.go                  # Unified session discovery
    sanitize.go                   # Prompt text sanitization
    pricing.go                    # Model pricing table and cost estimation
    search.go                     # Content search across JSONL files
  cmd/                            # Cobra CLI commands
    root.go                       # Root command (launches TUI)
    list.go                       # Non-interactive listing
    resume.go                     # Direct session resume
    version.go                    # Version output
  tui/                            # Bubble Tea TUI
    model.go                      # Main model (Init/Update/View)
    keys.go                       # Key bindings
    theme.go                      # Theme definitions (default, catppuccin, dracula, nord, light)
    styles.go                     # Lip Gloss styles (derived from active theme)
    list.go                       # Session list view
    detail.go                     # Session detail view
    stats.go                      # Stats popup view
demo/                             # Fake session data for demo recording
demo.tape                         # VHS tape for generating demo.gif
```

## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Terminal styling
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Ginkgo](https://github.com/onsi/ginkgo) / [Gomega](https://github.com/onsi/gomega) - Testing

## Contributing

Contributions are welcome! This project was built around my own workflow, so there are almost certainly use cases, filters, or data sources I've missed. If you have ideas or run into issues, please open an [issue](https://github.com/lyarwood/cctv/issues) or submit a PR.

Some areas that could use help:

- Additional themes or theme customization via config file
- New filter prefixes (e.g. `model:`, `status:`)
- Session sorting options (by tokens, message count, duration)
- Support for other conversation metadata not yet surfaced
- Packaging for Homebrew, Nix, or other package managers
