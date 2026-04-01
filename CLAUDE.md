---
scope: both
---

# desearch-cli

CLI for [Desearch AI](https://desearch.ai) — a contextual AI search engine that aggregates results across web, Hacker News, Reddit, Wikipedia, YouTube, Twitter/X, and arXiv with AI-synthesized answers and citations.

Single binary, no Python/Node dependencies.

## Sources

- **OpenAPI spec**: `openapi.json` (downloaded from `https://api.desearch.ai/openapi.json`)
- **API docs**: https://desearch.ai/api-reference
- **API key**: stored in `.env` (sourced from 1Password vault `Agents`, item `DESEARCH_API_KEY`). Never commit `.env`.

## Directory Structure

```
desearch-cli/
├── cmd/                          # Cobra CLI commands
│   ├── root.go                   # Root command, config loading, PreRunE dispatch
│   ├── search.go                 # search command + flags (default command)
│   ├── completion.go             # ai + completion subcommands (streaming)
│   ├── config.go                 # config show/clear/set commands
│   ├── version.go                # version command
│   ├── docs.go                   # docs command (prints embedded README)
│   ├── skill.go                  # skill print/add commands (Claude Code skill)
│   ├── completion_test.go
│   ├── config_test.go
│   ├── search_test.go
│   └── gendocs/main.go          # Man page generator
├── pkg/                          # Core packages
│   ├── api/client.go             # Desearch API client, request/response types
│   │   └── client_test.go
│   ├── auth/api_key.go           # XDG config loading/saving, GetAPIKey()
│   │   └── api_key_test.go
│   ├── output/formatter.go       # HumanFormatter, JSONFormatter, PlaintextFormatter, StreamingFormatter
│   │   └── formatter_test.go
│   └── errors/errors.go          # SystemError sentinel, exit code 3
│       └── errors_test.go
├── skill/
│   └── SKILL.md                 # Embedded Claude Code skill (go:embed)
├── docs/
│   └── config.md                # Full configuration documentation
├── integration_test.go           # Integration tests with mock httptest server
├── main.go                      # Entry point: calls cmd.Execute(), handles exit codes
├── go.mod / go.sum              # Dependencies
├── .goreleaser.yaml             # goreleaser build config
├── Makefile                     # check, build, test targets
└── README.md / GOAL.md / llms.txt
```

## Dependencies

| Library | Purpose |
|---------|---------|
| `spf13/cobra` | CLI framework |
| `spf13/viper` | Configuration management |
| `pelletier/go-toml/v2` | TOML config parsing |
| `itchyny/gojq` | jq expression filtering on JSON output |
| `stretchr/testify` | Testing assertions |
| `itchyny/timefmt-go` | Date formatting |

## API

- **Base URL**: `https://api.desearch.ai`
- **Auth**: `Authorization: Bearer <API_KEY>` header
- **Endpoint**: `POST /desearch/ai/search`
- **Streaming**: SSE-like streaming via `SearchStream()`, chunks printed line-by-line

## Request/Response Types (pkg/api/client.go)

```go
type SearchRequest struct {
    Prompt, Tools, StartDate, EndDate, DateFilter, ResultType, SystemMessage string
    Streaming *bool
    Count *int
}

type SearchResponse struct {
    Search           []WebResult
    HackerNewsSearch []HackerNewsResult
    RedditSearch     []RedditResult
    YoutubeSearch    []YoutubeResult
    Tweets           []TweetResult
    WikipediaSearch  []WikipediaResult
    ArxivSearch      []ArxivResult
    Text             string
    MinerLinkScores  map[string]string
    Completion       string
}
```

## Configuration

- **Path**: `~/.config/desearch-cli/config.toml` (XDG spec)
- **TOML schema**: `api_key`, `default_tools`, `default_date_filter`, `default_count`
- **Env override**: `DESEARCH_API_KEY` takes precedence over config file
- **Permissions**: 0600 on config file

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | User error or API error |
| 2 | Usage error (unknown flag/command) |
| 3+ | System error (network failure, unreadable config) |

## Testing

```bash
make test           # Smoke tests (no API key)
make test-unit      # Unit tests with coverage
make test-integration  # Integration tests with mock server
make check          # fmt + lint + test + test-unit
```

- Unit tests: `*_test.go` files in `cmd/` and `pkg/`
- Integration tests: `integration_test.go` (build tag `integration`), uses `httptest.NewServer` mock
- Live API tests: `SKIP_INTEGRATION=1` to skip, `DESEARCH_API_KEY=...` for live tests
- Coverage target: 75% minimum

## Building

```bash
make build          # Builds ./desearch binary
goreleaser build --snapshot --clean  # Cross-platform builds via goreleaser
```

- `CGO_ENABLED=0` (static binary, no cgo)
- GoReleaser builds: darwin/linux × arm64/amd64
- Version injected via `-ldflags "-X .../cmd.version=$(git describe --tags)"`

## Installation

- **Homebrew**: `roboalchemist/private` tap (push git tag → goreleaser auto-releases)
- **Binary**: Download from Gitea releases
- **Source**: `go install` or `go build`

## Command Tree

```
desearch [--api-key KEY] [--json] [--verbose] [--quiet] [--config PATH] [--version] [--help] <command>

Commands:
  search [query]      Search (default) — --tool, --date-filter, --start/ end-date, --streaming, --count, --system-message, --json, --no-ai, --plaintext, --dry-run, --jq, --fields, --stdin
  ai <query>          Streaming AI completion only (no per-source results)
  completion <shell>  Shell completion scripts (bash/zsh/fish/powershell)
  config [--api-key KEY] [--default-tool TOOL] [--default-date-filter FILTER] [--show] [--clear]
  version             Show version
  docs                Print embedded README to stdout
  skill [print|add]   Claude Code skill management
```

## Key Patterns

- **GNU `--` dispatch**: `desearch -- search query` routes to `search` subcommand (PreRunE manually dispatches)
- **No-auth commands**: `version`, `help`, `docs`, `skill`, `completion`, `ai`, shell completions, `config clear`
- **Config loading**: `auth.LoadConfig()` in root `PreRunE`, non-system errors are warnings (flags may override)
- **Output routing**: All output via `fmt.Fprint(os.Stdout/...)` and `os.Stdout.Sync()` for streaming
- **JSON sorting**: `SearchResponse.MarshalJSON()` sorts `MinerLinkScores` keys for deterministic output
