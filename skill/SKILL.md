---
name: desearch-cli
description: CLI for DeSearch AI — web search across multiple sources with AI summarization, streaming responses, and JSON output. Use when searching the web, researching topics, or aggregating results from Hacker News, Reddit, Wikipedia, YouTube, Twitter, and ArXiv.
scope: personal
allowed-tools: Bash(desearch-cli:*)
---

# desearch-cli

CLI for [DeSearch AI](https://desearch.ai) — a contextual AI search engine that aggregates results across multiple sources with AI-powered summarization.

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap roboalchemist/tap
brew install desearch-cli
```

### Build from Source

```bash
git clone https://github.com/roboalchemist/desearch-cli.git
cd desearch-cli
go install
```

Or with goreleaser:

```bash
goreleaser build --snapshot --clean
./dist/desearch-cli_darwin_amd64_v1/desearch-cli
```

## Authentication

DeSearch requires an API key. Sign up at [https://console.desearch.ai](https://console.desearch.ai).

```bash
desearch-cli config --api-key YOUR_API_KEY
```

The key is stored in `~/.config/desearch-cli/config.toml`.

## Usage

### Basic Search

```bash
desearch-cli search "golang best practices"
```

### Search Specific Sources

```bash
# Only Hacker News
desearch-cli search "rust vs go" --tool hackernews

# Multiple sources
desearch-cli search "AI news" --tool web --tool reddit --tool hackernews

# Available sources: web, hackernews, reddit, wikipedia, youtube, twitter, arxiv
```

### Date Filtering

```bash
# Predefined filters
desearch-cli search "latest AI news" --date-filter PAST_24_HOURS
desearch-cli search "this week in tech" --date-filter PAST_WEEK

# Custom date range (ISO8601 UTC)
desearch-cli search "recent research" --start-date 2026-01-01 --end-date 2026-03-01

# Available filters:
# PAST_24_HOURS, PAST_2_DAYS, PAST_WEEK, PAST_2_WEEKS,
# PAST_MONTH, PAST_2_MONTHS, PAST_YEAR, PAST_2_YEARS
```

### Streaming Results

```bash
desearch-cli search "explain transformers" --streaming
```

### JSON Output

```bash
desearch-cli search "golang concurrency" --json
desearch-cli search "rust memory model" --json --no-ai  # Raw results without AI summary
```

### Control Result Count

```bash
desearch-cli search "best practices" --count 20  # 10-200 results per source
```

### AI Completion Only (No Per-Source Links)

```bash
desearch-cli ai "what is bittensor"
desearch-cli ai "explain transformers" --system-message "Summarize in simple terms"
```

### History Logging

**Disabled by default.** Must be explicitly enabled. Useful for building a personal search corpus or mining agent query patterns over time.

```bash
# Enable history logging (off by default)
desearch-cli config --history-enabled=true

# Suppress for a single invocation even when enabled
desearch-cli search "my query" --no-history
desearch-cli ai "my query" --no-history

# Disable again
desearch-cli config --history-enabled=false
```

History files are written as JSON envelopes at:
`~/.config/desearch-cli/history/<cmd>/<year>/<month>/<day>/<timestamp>_<slug>_<hostname>.json`

## Configuration

For the full configuration schema, see [docs/config.md](../../docs/config.md).

### Show Current Config

```bash
desearch-cli config show
desearch-cli config show --json  # Machine-readable
```

### Set Defaults

```bash
# Search sources and date filter
desearch-cli config --default-tool web --default-tool hackernews --default-date-filter PAST_WEEK

# Result count (10-200)
desearch-cli config --default-count 20

# History logging
desearch-cli config --history-enabled=true
desearch-cli config --history-enabled=false
```

### Clear Config

```bash
desearch-cli config clear
```

## Global Flags

| Flag | Description |
|------|-------------|
| `--api-key KEY` | API key (overrides config file) |
| `--config PATH` | Config file path (default `~/.config/desearch-cli/config.toml`) |
| `--json` | Output in JSON format |
| `--quiet`, `-q` | Suppress stderr except errors |
| `--verbose`, `-v` | Verbose stderr output |
| `--help`, `-h` | Show help |
| `--version` | Show version |

## Examples

<examples>
<example>
Task: Quick web search

```bash
desearch-cli search "why use golang"
```

Output:
```
=== WEB ===
[Why Go? : r/golang - Reddit](https://www.reddit.com/r/golang/comments/11c9wv1/why_go/)
  Go scales easily to millions of network connections per box...

[What's so great about Go? - Stack Overflow](https://stackoverflow.blog/2020/11/02/go-golang-learn-fast-programming-languages/)
  Go is compilable on nearly any machine...

<!-- (additional results and AI summary omitted for brevity) -->
```

</example>

<example>
Task: Search with streaming

```bash
desearch-cli search "explain attention mechanism" --streaming
```

Output: Results stream in real-time as the AI processes them.

</example>

<example>
Task: Get JSON for scripting

```bash
desearch-cli search "golang best practices" --json 2>/dev/null | jq '.summary'
```

</example>

<example>
Task: Research across multiple sources

```bash
desearch-cli search "AI Agents research 2026" --tool web --tool arxiv --tool hackernews --date-filter PAST_MONTH --count 20
```

</example>

<example>
Task: Streaming AI answer without source links

```bash
desearch-cli ai "what is retrieval augmented generation"
```

</example>

<example>
Task: Enable history and search

```bash
desearch-cli config --history-enabled=true
desearch-cli search "recent Go releases" --tool web
# Writes result to ~/.config/desearch-cli/history/search/YYYY/MM/DD/...json
```

</example>
</examples>

## Troubleshooting

### "No API key found"

Run `desearch-cli config --api-key YOUR_KEY` to configure authentication.

### Version Check

```bash
desearch-cli version
```
