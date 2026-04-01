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
│   ├── root.go                   # Root command, config loading, PreRunE dispatch, GNU -- support
│   ├── search.go                 # search command + all search flags
│   ├── completion.go             # ai subcommand (streaming AI) + completion <shell> subcommands
│   ├── config.go                 # config command with show/clear subcommands and --api-key/--default-* flags
│   ├── version.go                # version command
│   ├── docs.go                   # docs command (prints embedded README)
│   ├── skill.go                  # skill print/add commands (Claude Code skill management)
│   ├── completion_test.go
│   ├── config_test.go
│   ├── search_test.go
│   └── gendocs/main.go          # Man page generator
├── pkg/                          # Core packages
│   ├── api/client.go             # Desearch API client, request/response types, MarshalJSON/UnmarshalJSON
│   │   └── client_test.go
│   ├── auth/api_key.go           # XDG config loading/saving, GetAPIKey(), ConfigPath()
│   │   └── api_key_test.go
│   ├── output/formatter.go       # HumanFormatter, JSONFormatter, PlaintextFormatter, StreamingFormatter, EvaluateJQ, FilterJSONFields
│   │   └── formatter_test.go
│   └── errors/errors.go          # SystemError sentinel, Wrap/WrapF/IsSystem helpers
│       └── errors_test.go
├── skill/
│   └── SKILL.md                 # Embedded Claude Code skill (go:embed), `skill add` installs to ~/.claude/skills/desearch/
├── docs/
│   └── config.md                # Full configuration schema documentation
├── .gitea/workflows/
│   └── bump-tap.yml             # Gitea Action: on release published → update homebrew-private formula
├── integration_test.go           # Integration tests (build tag: integration) using httptest.Server + exec
├── main.go                      # Entry point: cmd.Execute(), SystemError → exit 3, other errors → exit 1
├── go.mod / go.sum              # Go 1.26.1, module: github.com/roboalchemist/desearch-cli
├── .goreleaser.yaml             # GoReleaser: darwin/linux × arm64/amd64, brews tap config
├── Makefile                     # check, build, test, test-unit, test-integration, man, install targets
└── README.md / GOAL.md / llms.txt
```

## Dependencies

| Library | Purpose |
|---------|---------|
| `spf13/cobra` v1.10.2 | CLI framework |
| `spf13/viper` v1.21.0 | Configuration management |
| `pelletier/go-toml/v2` v2.2.4 | TOML config parsing/writing |
| `itchyny/gojq` v0.12.18 | jq expression filtering on JSON output |
| `stretchr/testify` v1.11.1 | Testing assertions |
| `itchyny/timefmt-go` v0.1.7 | Date formatting (indirect) |

## API

- **Base URL**: `https://api.desearch.ai`
- **Auth**: `Authorization: <API_KEY>` header (raw key, no "Bearer" prefix)
- **Endpoint**: `POST /desearch/ai/search`
- **Timeout**: 60 seconds on HTTP client
- **Streaming**: Server-sends JSON chunks line-by-line; `SearchStream()` returns a `*bufio.Reader`

## Request/Response Types (pkg/api/client.go)

```go
type SearchRequest struct {
    Prompt               string    `json:"prompt"`
    Tools                []string  `json:"tools,omitempty"`
    StartDate            *string   `json:"start_date,omitempty"`
    EndDate              *string   `json:"end_date,omitempty"`
    DateFilter           *string   `json:"date_filter,omitempty"`
    Streaming            *bool     `json:"streaming,omitempty"`
    ResultType           *string   `json:"result_type,omitempty"`
    SystemMessage        *string   `json:"system_message,omitempty"`
    ScoringSystemMessage *string   `json:"scoring_system_message,omitempty"`
    Count                *int      `json:"count,omitempty"`
}

type SearchResponse struct {
    Search           []WebResult
    HackerNewsSearch []HackerNewsResult
    RedditSearch     []RedditResult
    YoutubeSearch    []YoutubeResult
    Tweets           []TweetResult        // rich TweetResult with user, media, entities
    WikipediaSearch  []WikipediaResult
    ArxivSearch      []ArxivResult
    Text             string
    MinerLinkScores  map[string]string
    Completion       string
}
```

All result types (`WebResult`, `HackerNewsResult`, etc.) have `Title`, `Link`, `Snippet` fields.
`TweetResult` is richer: includes `User`, engagement counts, `Entities`, `Media`.

`MarshalJSON` serializes `MinerLinkScores` as a sorted `[{key,value}]` array for deterministic output.
`UnmarshalJSON` handles both map and sorted-array formats (round-trip safe).

## Configuration

- **Path**: `~/.config/desearch-cli/config.toml` (XDG spec; `XDG_CONFIG_HOME` honored)
- **TOML schema**: `api_key`, `default_tools []string`, `default_date_filter`, `default_count int`
- **Env override**: `DESEARCH_API_KEY` takes precedence over config file
- **Permissions**: config written with mode 0600, config dir with 0700

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | User error or API error |
| 2 | Usage error (unknown flag/command) |
| 3 | System error (network failure, unreadable config) |

Exit code 3 is triggered by `errors.SystemError` returned from `auth.LoadConfig()` or `api.Client` network errors.

## Testing

```bash
make test                    # Smoke tests: build + run help/version/docs/skill/completion
make test-unit               # Unit tests with -race and coverage (75% minimum)
make test-integration        # Integration tests (build tag: integration) with httptest mock
make test-integration-live   # Live API tests (requires DESEARCH_API_KEY env var)
make check                   # fmt + lint + test + test-unit (CI gate)
```

- **Unit tests** (`*_test.go` in `cmd/`, `pkg/`): test flag parsing, formatters, auth, error types
- **Integration tests** (`integration_test.go`, build tag `integration`): build binary with `exec.Command`, test via subprocess + `httptest.Server`; use `READONLY=1` to skip filesystem-mutating tests
- **Live tests**: skip with `SKIP_INTEGRATION=1`; run with `DESEARCH_API_KEY` set
- Tests use `resetFlags()` helpers to reset package-level flag vars between test cases

## Building

```bash
make build          # Builds ./desearch binary with version from git describe
goreleaser build --snapshot --clean  # Cross-platform snapshot builds
```

- `CGO_ENABLED=0` (fully static binary)
- GoReleaser targets: darwin/linux × arm64/amd64
- Version injected via `-ldflags "-X github.com/roboalchemist/desearch-cli/cmd.version=<version>"`

## Installation & Release

- **Homebrew**: `roboalchemist/private` tap on Gitea (`ssh://git@gitea.roboalch.com:2222/roboalchemist/homebrew-private.git`)
- **Release flow**: push git tag → Gitea Action (`.gitea/workflows/bump-tap.yml`) runs on `release: published` → clones `homebrew-private`, `sed`-patches version in `Formula/desearch-cli.rb`, pushes
- **Binary**: download from Gitea releases
- **Source**: `go install` or `make build`
- **Manual install**: `make install` → copies to `/usr/local/bin/desearch`

## Command Tree

```
desearch [--api-key KEY] [--json] [--verbose/-v] [--quiet/-q] [--silent] [--config PATH] [--version] [--help] <command>

Commands:
  search <query>     Search — flags: --tool (repeatable), --date-filter, --start-date, --end-date,
                               --streaming, --count, --result-type, --system-message,
                               --no-ai, --plaintext/-p, --dry-run, --jq, --fields, --stdin
  ai <query>         Streaming AI completion only (no per-source results); --system-message, --json
  completion <shell> Shell completion scripts: bash | zsh | fish | powershell
  config             Manage config — subcommands: show, clear
                     flags on config itself: --api-key, --default-tool (repeatable), --default-date-filter
  config show        Display current config (masked API key, or --json for full)
  config clear       Remove config file; --force/-f to skip confirmation
  version            Show version
  docs               Print embedded README to stdout
  skill print        Print SKILL.md to stdout
  skill add          Install SKILL.md to ~/.claude/skills/desearch/SKILL.md
```

## Key Patterns

- **GNU `--` dispatch**: `desearch -- search query` routes to `search` subcommand; implemented in root `PreRunE` by calling `cmd.Find()` then manually running `ParseFlags`, `PersistentPreRunE`, and `RunE` on the subcommand.
- **No-auth commands**: `version`, `help`, `docs`, `skill` (and `print`/`add`), `completion` (and `bash`/`zsh`/`fish`/`powershell`), `ai`, `clear`. Checked in `PersistentPreRun` via `isNoAuthCommand()`.
- **Dry-run auth bypass**: `PersistentPreRun` also skips the API key check if `--dry-run` or `--fields` is set, or if `hasDryRunInArgs()` detects them after `--` in `os.Args`.
- **Config loading**: `auth.LoadConfig()` called in root `PreRunE` — system errors exit 3, non-system errors print a warning and continue (flags may still provide the key).
- **Output routing**: All output via `fmt.Fprint(os.Stdout)` + `os.Stdout.Sync()` for streaming flush.
- **JSON serialization**: `SearchResponse.MarshalJSON()` sorts `MinerLinkScores` map into `[{key,value}]` array; `UnmarshalJSON` handles both map (from API) and sorted-array (from self) formats.
- **Formatter selection**: `output.NewFormatter(OutputFlags)` returns `JSONFormatter`, `PlaintextFormatter`, or `HumanFormatter` based on flags. `--no-ai` implies JSON mode. `EvaluateJQ` and `FilterJSONFields` applied post-format.
- **Streaming (ai cmd)**: reads `bufio.Reader` line-by-line, parses each chunk as JSON `{completion, text}`, writes immediately via `StreamingFormatter.WriteChunk`. Handles Ctrl+C via `context.WithCancel` + signal goroutine.
