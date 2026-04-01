---
name: desearch
description: CLI for DeSearch AI — web search across multiple sources with AI summarization, streaming responses, and JSON output. Use when searching the web, researching topics, or aggregating results from Hacker News, Reddit, Wikipedia, YouTube, Twitter, and ArXiv.
scope: personal
allowed-tools: Bash(desearch:*)
---

# desearch

CLI for [DeSearch AI](https://desearch.ai) — a contextual AI search engine that aggregates results across multiple sources with AI-powered summarization.

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap roboalchemist/private ssh://git@gitea.roboalch.com:2222/roboalchemist/homebrew-private.git
brew install desearch-cli
```

### Binary

Download prebuilt binaries from the GitHub releases page.

### Build from Source

```bash
git clone ssh://git@gitea.roboalch.com:2222/roboalchemist/desearch-cli.git
cd desearch-cli
go install
```

Or with goreleaser:

```bash
goreleaser build --snapshot --clean
./dist/desearch-cli_darwin_amd64_v1/desearch
```

## Authentication

DeSearch requires an API key. Sign up at [https://console.desearch.ai](https://console.desearch.ai).

### Configure API Key

```bash
desearch config --api-key YOUR_API_KEY
```

The key is stored in `~/.config/desearch-cli/config.toml`.

## Usage

### Basic Search

```bash
desearch "golang best practices"
```

### Search Specific Sources

```bash
# Only Hacker News
desearch "rust vs go" --tool hackernews

# Multiple sources
desearch "AI news" --tool web --tool reddit --tool hackernews

# Available sources: web, hackernews, reddit, wikipedia, youtube, twitter, arxiv
```

### Date Filtering

```bash
# Use predefined filters
desearch "latest AI news" --date-filter PAST_24_HOURS
desearch "this week in tech" --date-filter PAST_WEEK

# Custom date range (ISO8601 UTC)
desearch "recent research" --start-date 2026-01-01 --end-date 2026-03-01

# Available filters:
# PAST_24_HOURS, PAST_2_DAYS, PAST_WEEK, PAST_2_WEEKS,
# PAST_MONTH, PAST_2_MONTHS, PAST_YEAR, PAST_2_YEARS
```

### Streaming Results

Stream results as they arrive:

```bash
desearch "explain transformers" --streaming
```

### JSON Output

```bash
desearch "golang concurrency" --json
desearch "rust memory model" --json --no-ai  # Raw results without AI summary
```

### Control Result Count

```bash
desearch "best practices" --count 20  # 10-200 results per source
```

### AI Completion Only (No Per-Source Links)

```bash
desearch completion "what is bittensor"
desearch completion "explain transformers" --system-message "Summarize in simple terms"
```

## Output Format

### Default (Human-Readable)

Returns AI-generated summary with source links.

### JSON (`--json`)

Returns structured JSON with:

```json
{
  "results": [...],
  "summary": "...",
  "sources": [...]
}
```

### With `--no-ai`

Returns raw per-source results without AI summarization.

## Configuration

For the full configuration schema, see [docs/config.md](../../docs/config.md).

### Show Current Config

```bash
desearch config show
```

### Set Defaults

```bash
desearch config --default-tool web --default-tool hackernews --default-date-filter PAST_WEEK
```

### Clear Config

```bash
desearch config clear
```

## Global Flags

| Flag | Description |
|------|-------------|
| `--api-key KEY` | API key (overrides config file) |
| `--config PATH` | Config file path (default `~/.config/desearch-cli/config.toml`) |
| `--json` | Output in JSON format |
| `--help`, `-h` | Show help |
| `--version` | Show version |

## Examples

<examples>
<example>
Task: Quick web search

```bash
desearch "why use golang"
```

Output:
```
=== WEB ===
[Why Go? : r/golang - Reddit](https://www.reddit.com/r/golang/comments/11c9wv1/why_go/)
  Go scales easily to millions of network connections per box and allows for better utilization of hardware...

[What's so great about Go? - Stack Overflow - StackOverflow blog](https://stackoverflow.blog/2020/11/02/go-golang-learn-fast-programming-languages/)
  Go is compilable on nearly any machine, so you can use it to create a full webapp or a tool...

[Why Go - The Go Programming Language](https://go.dev/solutions/)
  Common problems companies solve with Go. Use Cases. Stories about how and why companies use Go.

<!-- (additional results and AI summary omitted for brevity) -->
```

</example>

<example>
Task: Search with streaming

```bash
desearch "explain attention mechanism" --streaming
```

Output: Results stream in real-time as the AI processes them.

</example>

<example>
Task: Get JSON for scripting

```bash
desearch "golang best practices" --json 2>/dev/null | jq '.summary'
```

</example>

<example>
Task: Research across multiple sources

```bash
desearch "AI Agents research 2026" --tool web --tool arxiv --tool hackernews --date-filter PAST_MONTH --count 20
```

</example>

<example>
Task: Quick completion without links

```bash
desearch completion "what is retrieval augmented generation"
```

</example>
</examples>

## Troubleshooting

### "No API key found"

Run `desearch config --api-key YOUR_KEY` to configure authentication.

### Version Check

```bash
desearch version
```
