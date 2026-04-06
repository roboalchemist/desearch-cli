# desearch-cli

A fast, scriptable CLI for [Desearch AI](https://desearch.ai) — a contextual AI search engine that aggregates results across web, Hacker News, Reddit, Wikipedia, YouTube, Twitter/X, and arXiv with AI-synthesized answers and citations.

Single binary, no Python or Node.js dependencies.

## Install

### Homebrew (recommended)

```bash
brew tap roboalchemist/tap
brew install desearch-cli
```

### Binary

Download the latest release for your platform from the [releases page](https://github.com/roboalchemist/desearch-cli/releases), then:

```bash
# macOS (Apple Silicon)
curl -L https://github.com/roboalchemist/desearch-cli/releases/latest/download/desearch-cli_Darwin_arm64.tar.gz | tar xz
sudo mv desearch-cli /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/roboalchemist/desearch-cli/releases/latest/download/desearch-cli_Darwin_amd64.tar.gz | tar xz
sudo mv desearch-cli /usr/local/bin/

# Linux
curl -L https://github.com/roboalchemist/desearch-cli/releases/latest/download/desearch-cli_Linux_x86_64.tar.gz | tar xz
sudo mv desearch-cli /usr/local/bin/
```

### Build from source

```bash
go install github.com/roboalchemist/desearch-cli@latest
```

## Setup

Get an API key at [https://console.desearch.ai](https://console.desearch.ai), then configure it:

```bash
desearch-cli config --api-key <YOUR_KEY>
```

The key is stored at `~/.config/desearch-cli/config.toml`. You can also set `DESEARCH_API_KEY` as an environment variable (takes precedence over the config file).

## Basic Usage

```bash
desearch-cli search "golang best practices"
```

## Commands

### `search`

```bash
desearch-cli search [query] [flags]
```

Search across all or specified sources with AI-synthesized results.

### `ai`

```bash
desearch-cli ai <query> [flags]
```

Streaming AI summary only — no per-source result links.

### `config`

```bash
desearch-cli config [flags]
desearch-cli config show
desearch-cli config clear
```

Manage API key, defaults, and history settings. See [Config Flags](#config-flags) below.

### `version`

```bash
desearch-cli version
```

### `docs`

```bash
desearch-cli docs
```

Print the embedded documentation to stdout.

### `skill`

```bash
desearch-cli skill [print|add]
```

Claude Code skill management. `add` installs the skill to `~/.claude/skills/desearch-cli/`.

## Search Flags

| Flag | Type | Description |
|------|------|-------------|
| `--tool strings` | []string | Sources: `web`, `hackernews`, `reddit`, `wikipedia`, `youtube`, `twitter`, `arxiv` (default: `web`) |
| `--date-filter string` | string | `PAST_24_HOURS`, `PAST_2_DAYS`, `PAST_WEEK`, `PAST_2_WEEKS`, `PAST_MONTH`, `PAST_2_MONTHS`, `PAST_YEAR`, `PAST_2_YEARS` |
| `--start-date string` | string | ISO8601 UTC start bound |
| `--end-date string` | string | ISO8601 UTC end bound |
| `--count int` | int | Results per source, 10–200 (default: 10) |
| `--streaming` | bool | Stream results as they arrive |
| `--result-type string` | string | `ONLY_LINKS` or `LINKS_WITH_FINAL_SUMMARY` (default) |
| `--system-message string` | string | Override AI system prompt |
| `--scoring-system-message string` | string | Override scoring/ranking prompt |
| `--no-ai` | bool | Skip AI summary; implies `--json` |
| `--json` | bool | JSON output |
| `--plaintext` / `-p` | bool | Tab-separated title/url/snippet |
| `--fields string` | string | Comma-separated top-level JSON keys to include |
| `--jq string` | string | jq expression (requires `--json`, `--no-ai`, or `--dry-run`) |
| `--streaming` | bool | Stream results as they arrive |
| `--stdin` | bool | Read queries from stdin, one per line |
| `--dry-run` / `-D` | bool | Print request JSON without calling API |
| `--no-history` | bool | Skip writing history for this invocation |

## Config Flags

| Flag | Description |
|------|-------------|
| `--api-key string` | Set API key |
| `--default-tool strings` | Set default sources (repeatable) |
| `--default-date-filter string` | Set default date filter |
| `--default-count int` | Set default result count per source (10–200, or 0 to clear) |
| `--history-enabled` | Enable or disable history logging (`--history-enabled=true` / `--history-enabled=false`) |

## History Logging

**Disabled by default.** When enabled, each search and ai result is saved as a JSON file locally. Useful for building a personal search corpus, mining AI agent patterns, or auditing what queries were made.

```bash
# Enable
desearch-cli config --history-enabled=true

# Disable
desearch-cli config --history-enabled=false

# Skip for a single invocation
desearch-cli search "my query" --no-history
desearch-cli ai "my query" --no-history
```

Files are written to:
```
~/.config/desearch-cli/history/<cmd>/<year>/<month>/<day>/<timestamp>_<slug>_<hostname>.json
```

Each file contains a JSON envelope:
```json
{
  "meta": { "timestamp": "...", "command": "search", "params": {...}, "latency_ms": 1234 },
  "response": { "search": [...], "completion": "..." }
}
```

## Examples

### Search specific sources

```bash
desearch-cli search "llm benchmarks" --tool web --tool hackernews
desearch-cli search "rust vs go" --tool reddit --tool youtube
```

### Date filtering

```bash
desearch-cli search "AI news" --date-filter PAST_2_DAYS
desearch-cli search "golang updates" --start-date 2026-01-01 --end-date 2026-03-01
```

### JSON output for scripting

```bash
desearch-cli search "react patterns" --json 2>/dev/null | jq '.completion'
desearch-cli search "topic" --json --fields search,completion 2>/dev/null
```

### AI summary only

```bash
desearch-cli ai "what is bittensor"
desearch-cli ai "explain transformers" --system-message "Summarize for a non-technical audience"
```

### Set config defaults

```bash
desearch-cli config --api-key sk-xxxx --default-tool web --default-date-filter PAST_WEEK --default-count 20
desearch-cli config show
```

### Enable history and search

```bash
desearch-cli config --history-enabled=true
desearch-cli search "Go concurrency patterns"
# Result saved to ~/.config/desearch-cli/history/search/YYYY/MM/DD/...json
```

## Configuration

Config is stored at `~/.config/desearch-cli/config.toml` (XDG spec, mode 0600).

See [docs/config.md](docs/config.md) for the full TOML schema and all options.

## Output Format

Default human-readable output:

```
=== WEB ===
[Article Title](https://example.com)
  Article snippet text

=== HACKERNEWS ===
[HN Post Title](https://news.ycombinator.com/item?id=123)
  Post snippet

=== AI SUMMARY ===
AI-generated synthesis of the results...
```

Use `--json` for raw structured JSON. Use `--plaintext` for tab-separated title/url/snippet.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | User/API error |
| 2 | Usage error (unknown flag/command) |
| 3 | System error (network, config) |

## Releases

Push a version tag to trigger an automated release:
```bash
git tag v0.1.0 && git push origin v0.1.0
```
GitHub Actions builds cross-platform binaries via GoReleaser and updates the Homebrew tap automatically.

## License

[MIT](LICENSE)

## Issue Tracker

Report bugs and request features at: [https://github.com/roboalchemist/desearch-cli/issues](https://github.com/roboalchemist/desearch-cli/issues)
