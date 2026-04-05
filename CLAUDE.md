---
scope: both
---

# desearch-cli

CLI for [Desearch AI](https://desearch.ai) — a contextual AI search engine that aggregates results across web, Hacker News, Reddit, Wikipedia, YouTube, Twitter/X, and arXiv with AI-synthesized answers and citations.

Single binary, no Python/Node dependencies.

## Sources

- **OpenAPI spec**: `openapi.json` (downloaded from `https://api.desearch.ai/openapi.json`)
- **API docs**: https://desearch.ai/api-reference
- **API key**: stored in `~/.config/desearch-cli/config.toml`. Set via `desearch-cli config --api-key KEY` or `DESEARCH_API_KEY` env var.

## Directory Structure

```
desearch-cli/
├── cmd/                          # Cobra CLI commands
│   ├── root.go                   # Root command, config loading, PreRunE dispatch, GNU -- support
│   ├── search.go                 # search command + all search flags + runSearch/runSearchOne/runSearchNormal/runSearchStream
│   ├── completion.go             # ai subcommand (streaming AI) + completion <shell> subcommands
│   ├── config.go                 # config command with show/clear subcommands and --api-key/--default-* flags
│   ├── tools.go                  # resolveTools() — flag → config → default ["web"] tool resolution
│   ├── version.go                # version command
│   ├── docs.go                   # docs command (embeds cmd/README.md via go:embed)
│   ├── skill.go                  # skill print/add commands (Claude Code skill management)
│   ├── README.md                 # Embedded by docs.go (go:embed target)
│   ├── root_test.go
│   ├── completion_test.go
│   ├── config_test.go
│   ├── search_test.go
│   ├── tools_test.go
│   └── gendocs/main.go          # Man page generator
├── pkg/                          # Core packages
│   ├── api/client.go             # Desearch API client, request/response types, MarshalJSON/UnmarshalJSON
│   │   └── client_test.go
│   ├── auth/api_key.go           # XDG config loading/saving, GetAPIKey(), ConfigPath()
│   │   └── api_key_test.go
│   ├── output/
│   │   ├── formatter.go          # HumanFormatter, JSONFormatter, PlaintextFormatter, StreamingFormatter, EvaluateJQ, FilterJSONFields
│   │   ├── formatter_test.go
│   │   ├── sse.go                # ParseSSEEvent() — SSE chunk parsing shared by search and ai commands
│   │   └── sse_test.go
│   └── errors/errors.go          # SystemError + UsageError sentinels, Wrap/WrapF/WrapUsage/IsSystem/IsUsage
│       └── errors_test.go
├── skill/
│   └── SKILL.md                 # Embedded Claude Code skill (go:embed), `skill add` installs to ~/.claude/skills/desearch-cli/
├── docs/
│   └── config.md                # Full configuration schema documentation
├── .github/workflows/
│   └── bump-tap.yml             # GitHub Action: on release published → update homebrew-tap formula
├── integration_test.go           # Integration tests (build tag: integration) using httptest.Server + exec
├── main.go                      # Entry point: cmd.Execute(), SystemError → exit 3, UsageError → exit 2
├── go.mod / go.sum              # Go 1.26.1, module: github.com/roboalchemist/desearch-cli
├── .goreleaser.yaml             # GoReleaser: darwin/linux × arm64/amd64, brews tap config
├── Makefile                     # check, build, test, test-unit, test-integration, man, install targets
└── README.md / GOAL.md / llms.txt
```

## Dependencies

| Library | Purpose |
|---------|---------|
| `spf13/cobra` v1.10.2 | CLI framework |
| `spf13/viper` v1.21.0 | Listed as direct dep (cobra transitively uses it) |
| `pelletier/go-toml/v2` v2.2.4 | TOML config parsing/writing (used directly in auth/) |
| `itchyny/gojq` v0.12.18 | jq expression filtering on JSON output |
| `stretchr/testify` v1.11.1 | Testing assertions |
| `itchyny/timefmt-go` v0.1.7 | Date formatting (indirect) |

## API

- **Base URL**: `https://api.desearch.ai`
- **Auth**: `Authorization: <API_KEY>` header (raw key, **no "Bearer" prefix** — the API rejects Bearer)
- **Endpoint**: `POST /desearch/ai/search`
- **Timeout**: 60 seconds on HTTP client
- **Streaming**: Server sends SSE-style JSON chunks; `SearchStream()` returns `*streamReadCloser` (bufio.Reader + io.Closer)

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
`UnmarshalJSON` handles both map (from API) and sorted-array (from self) formats (round-trip safe).

## SSE Streaming Format

The Desearch API may send multiple SSE events on a single line without newline separators.
Parsing is handled by `output.ParseSSEEvent()` in `pkg/output/sse.go`:
- Splits each read on `"data: "` boundaries
- Only `{"type":"text","content":"..."}` events produce output
- `[DONE]` sentinel, non-JSON garbage, and other event types are silently skipped

## Configuration

- **Path**: `~/.config/desearch-cli/config.toml` (XDG spec; `XDG_CONFIG_HOME` honored)
- **TOML schema**: `api_key`, `default_tools []string`, `default_date_filter`, `default_count int`
- **Env override**: `DESEARCH_API_KEY` takes precedence over config file
- **Permissions**: config written with mode 0600, config dir with 0700

## Tool Resolution Chain (cmd/tools.go)

```
1. --tool flag(s) on command line
2. default_tools in ~/.config/desearch-cli/config.toml
3. Hard-coded fallback: ["web"]
```

`resolveTools(flagTools, cfg)` always returns a non-empty slice. The Desearch API returns 422 if no tools are sent.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | User error or API error |
| 2 | Usage error (unknown flag/command) — `errors.UsageError` |
| 3 | System error (network failure, unreadable config) — `errors.SystemError` |

`main.go` dispatches exit codes: `IsSystem` → 3, `IsUsage` → 2, otherwise 1.

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
- **Live tests**: run with `DESEARCH_API_KEY` set
- Tests use `resetFlags()` helpers to reset package-level flag vars between test cases
- The mock server checks for a non-empty `Authorization` header (does NOT validate Bearer vs raw key)
- `golangci-lint` must pass — run `make lint` before committing

## Building

```bash
make build          # Builds ./desearch-cli binary with version from git describe
goreleaser build --snapshot --clean  # Cross-platform snapshot builds
```

- `CGO_ENABLED=0` (fully static binary)
- GoReleaser targets: darwin/linux × arm64/amd64
- Version injected via `-ldflags "-X github.com/roboalchemist/desearch-cli/cmd.version=<version>"`

## Installation & Release

- **Homebrew**: `roboalchemist/tap` tap on GitHub (`brew tap roboalchemist/tap && brew install desearch-cli`)
- **Release flow**: push git tag → GitHub Action (`.github/workflows/bump-tap.yml`) runs on `release: published` → updates `homebrew-tap` formula via sed on `.rb` file
- **Binary**: download from GitHub releases
- **Source**: `go install` or `make build`
- **Manual install**: `make install` → `sudo install -m 755 desearch /usr/local/bin/`

## Command Tree

```
desearch-cli [--api-key KEY] [--json] [--verbose/-v] [--quiet/-q] [--silent] [--config PATH] [--version] [--help] <command>

Commands:
  search <query>     Search — flags: --tool (repeatable), --date-filter, --start-date, --end-date,
                               --streaming, --count, --result-type, --system-message,
                               --scoring-system-message, --no-ai, --plaintext/-p, --dry-run/-D,
                               --jq, --fields, --stdin
  ai <query>         Streaming AI completion only (no per-source results); --system-message, --json
  completion <shell> Shell completion scripts: bash | zsh | fish | powershell
  config             Manage config — subcommands: show, clear
                     flags on config itself: --api-key, --default-tool (repeatable), --default-date-filter
  config show        Display current config (masked API key, or --json for full)
  config clear       Remove config file; --force/-f flag exists but no confirmation prompt is implemented
  version            Show version
  docs               Print embedded cmd/README.md to stdout (aliases: readme)
  skill print        Print SKILL.md to stdout
  skill add          Install SKILL.md to ~/.claude/skills/desearch-cli/SKILL.md
```

## Key Patterns

- **GNU `--` dispatch**: `desearch -- search query` routes to `search` subcommand; implemented in root `PreRunE` by calling `cmd.Find()` then manually running `ParseFlags`, `PersistentPreRunE`, and `RunE` on the subcommand.
- **No-auth commands**: `version`, `help`, `docs`, `skill` (and `print`/`add`), `completion` (and `bash`/`zsh`/`fish`/`powershell`), `clear`. Checked in `PersistentPreRun` via `isNoAuthCommand()`. Note: `ai` and `search` require auth.
- **Dry-run auth bypass**: `PersistentPreRun` skips the API key check if `--dry-run` or `--fields` is set, or if `hasDryRunInArgs()` detects them after `--` in `os.Args`.
- **Config loading**: `auth.LoadConfig()` called in root `PreRunE` — system errors exit 3, non-system errors print a warning and continue (flags may still provide the key).
- **Output routing**: All output via `fmt.Fprint(os.Stdout)` + `os.Stdout.Sync()` for streaming flush.
- **JSON serialization**: `SearchResponse.MarshalJSON()` sorts `MinerLinkScores` map into `[{key,value}]` array; `UnmarshalJSON` handles both map (from API) and sorted-array (from self) formats.
- **Formatter selection**: `output.NewFormatter(OutputFlags)` returns `JSONFormatter`, `PlaintextFormatter`, or `HumanFormatter` based on flags. `--no-ai` implies JSON mode. `EvaluateJQ` and `FilterJSONFields` applied post-format.
- **Streaming (search --streaming)**: reads `streamReadCloser` line-by-line, splits on `"data: "` boundaries, delegates to `output.ParseSSEEvent()` for each segment, writes via `StreamingFormatter.WriteChunk`.
- **Streaming (ai cmd)**: same SSE parsing as search streaming; handles Ctrl+C via `context.WithCancel` + signal goroutine; `--json` mode accumulates text then outputs JSON at end.
- **`--jq` validation**: requires `--json`, `--no-ai`, or `--dry-run` (all produce JSON). `--fields` requires `--json` or `--dry-run`.
