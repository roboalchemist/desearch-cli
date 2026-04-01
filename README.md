# desearch-cli

A fast, scriptable CLI for [Desearch AI](https://desearch.ai) — a contextual AI search engine that aggregates results across web, Hacker News, Reddit, Wikipedia, YouTube, Twitter/X, and arXiv with AI-synthesized answers and citations.

Single binary, no Python or Node.js dependencies.

## Install

### Homebrew (recommended)

```bash
brew tap roboalchemist/private ssh://git@gitea.roboalch.com:2222/roboalchemist/homebrew-private.git
brew install desearch
```

### Binary

Download the latest release for your platform from the [releases page](https://git.roboalch.com/roboalchemist/desearch-cli/releases), then:

```bash
# macOS (Intel or Apple Silicon)
curl -L https://git.roboalch.com/roboalchemist/desearch-cli/releases/latest/download/desearch-cli_darwin_amd64.tar.gz | tar xz
sudo mv desearch /usr/local/bin/

# Linux
curl -L https://git.roboalch.com/roboalchemist/desearch-cli/releases/latest/download/desearch-cli_linux_amd64.tar.gz | tar xz
sudo mv desearch /usr/local/bin/
```

### Build from source

```bash
# Requires Go 1.26+
go install github.com/roboalchemist/desearch-cli@latest
```

Or use [goreleaser](https://goreleaser.com/):

```bash
goreleaser build --snapshot --clean
```

## Homebrew Releases

Releases to the private Homebrew tap require:
1. Set `HOMEBREW_TAP_TOKEN` secret in Gitea Actions (API token with repo scope)
2. Push a version tag: `git tag v0.1.0 && git push origin v0.1.0`
3. GoReleaser builds and pushes the formula automatically

## Setup

Get an API key at [https://console.desearch.ai](https://console.desearch.ai), then configure it:

```bash
desearch config --api-key <YOUR_KEY>
```

That's it. The API key is stored in your config file at `~/.config/desearch-cli/config.toml`.

## Basic Usage

```bash
desearch "golang best practices"
```

## Commands

### `search` (default)

```bash
desearch [query] [flags]
```

Search across all or specified sources. This is the default command, so `desearch "query"` and `desearch search "query"` are equivalent.

### `completion`

```bash
desearch completion <query> [flags]
```

Stream an AI-generated summary without per-source search results.

### `config`

```bash
desearch config [flags]
```

Manage API key and default settings.

### `version`

```bash
desearch version
```

Show version information.

### `ai`

```bash
desearch ai [query] [flags]
```

Streaming AI completion without per-source search results — streams the AI-generated response as it is generated.

### `docs`

```bash
desearch docs
```

Print the embedded README documentation to stdout. Useful for offline reference.

### `skill`

```bash
desearch skill [print|add]
```

Claude Code skill management. `print` outputs the SKILL.md to stdout; `add` installs the skill to `~/.claude/skills/desearch-cli/`.

## Search Flags

| Flag | Type | Description |
|------|------|-------------|
| `--query string` | string | Search prompt (also the positional arg) |
| `--tool strings` | []string | Sources to query: `web`, `hackernews`, `reddit`, `wikipedia`, `youtube`, `twitter`, `arxiv` (default: all) |
| `--date-filter string` | string | Predefined date range. One of: `PAST_24_HOURS`, `PAST_2_DAYS`, `PAST_WEEK`, `PAST_2_WEEKS`, `PAST_MONTH`, `PAST_2_MONTHS`, `PAST_YEAR`, `PAST_2_YEARS` |
| `--start-date string` | string | ISO8601 UTC start bound, e.g. `2026-01-01T00:00:00Z` |
| `--end-date string` | string | ISO8601 UTC end bound |
| `--streaming` | bool | Stream results as they arrive |
| `--result-type string` | string | `ONLY_LINKS` or `LINKS_WITH_FINAL_SUMMARY` (default: `LINKS_WITH_FINAL_SUMMARY`) |
| `--count int` | int | Results per source, 10-200 (default: 10) |
| `--system-message string` | string | Override system prompt to influence AI behavior |
| `--json` | bool | Output raw JSON instead of formatted human-readable |
| `--no-ai` | bool | Skip AI completion/summary |

## Config Flags

| Flag | Description |
|------|-------------|
| `--api-key string` | Set API key (get at https://console.desearch.ai) |
| `--default-tool strings` | Set default sources (can be specified multiple times) |
| `--default-date-filter string` | Set default date filter |
| `--show` | Display current configuration (alias: `desearch config show`) |
| `--clear` | Reset configuration to defaults |

## Examples

### Search with specific sources

```bash
desearch "llm benchmarks" --tool web --tool hackernews
desearch "rust vs go performance" --tool reddit --tool youtube
```

### Date filtering

```bash
desearch "AI news" --date-filter PAST_2_DAYS
desearch "golang updates" --start-date 2026-01-01T00:00:00Z --end-date 2026-03-01T00:00:00Z
```

### Limit results

```bash
desearch "react patterns" --tool web --count 20
```

### Custom system message

```bash
desearch "explain quantum computing" --system-message "Summarize in simple terms for a non-technical audience"
```

### Streaming output

```bash
desearch "latest AI research" --streaming
```

### AI completion only (no per-source results)

```bash
desearch completion "what is bittensor"
desearch completion "explain transformers" --system-message "Summarize in simple terms"
```

### JSON output

```bash
desearch "golangci-lint" --json | jq '.search[0].title'
desearch "linux kernel news" --tool hackernews --json
```

### Skip AI summary

```bash
desearch "hacker news top posts" --no-ai
```

### Set defaults

```bash
desearch config --api-key sk-xxxx --default-tool web --default-date-filter PAST_WEEK
```

### View current config

```bash
desearch config show
```

### Clear config

```bash
desearch config clear
```

## Configuration File

See [docs/config.md](docs/config.md) for full configuration documentation including the TOML schema, JSONC schema with comments, environment variable overrides, and defaults table.

Config is stored at `~/.config/desearch-cli/config.toml` (XDG spec).

## Output Format

By default, output is formatted with section headers per source:

```
=== WEB ===
[Article Title](https://example.com)
  Article snippet text

[Another Article](https://another.com)
  Another snippet

=== HACKERNEWS ===
[HN Post Title](https://news.ycombinator.com/item?id=123)
  Post snippet

=== REDDIT ===
[Reddit Post Title](https://reddit.com/r/...)
  Post snippet

=== TWITTER ===
@username
  Tweet text
  Link: https://x.com/...
  42 likes, 12 retweets, 5 replies

=== AI SUMMARY ===
AI-generated synthesis of the results...
```

Twitter results include engagement metrics (likes, retweets, replies, quotes, bookmarks) when available.

Use `--json` for raw JSON output.

## License

See project repository.

## Issue Tracker

Report bugs and request features at: [https://git.roboalch.com/roboalchemist/desearch-cli/issues](https://git.roboalch.com/roboalchemist/desearch-cli/issues)
