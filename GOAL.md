# desearch-cli — Goal

## Overview

CLI tool for [Desearch AI](https://desearch.ai) — a contextual AI search engine that aggregates results across web, hackernews, reddit, wikipedia, youtube, twitter/X, and arxiv with AI-synthesized answers and citations.

## Why

Provides a fast, scriptable terminal interface to Desearch's multi-source AI search. Patterned after `tavily-cli` and `perplexity-cli` — single binary, no Python/node dependencies, API key stored in config.

## Command Tree

```
desearch <command> [flags]

Commands:
  search     Search across all or specified sources (default command)
  completion Stream AI-generated summary for a query
  config     Manage API key and defaults
  version    Show version info

Search flags:
  --query string          Search prompt (or positional arg)
  --tool strings          Sources to query: web, hackernews, reddit, wikipedia, youtube, twitter, arxiv (default: all)
  --date-filter string    PAST_24_HOURS, PAST_2_DAYS, PAST_WEEK, PAST_2_WEEKS, PAST_MONTH, PAST_2_MONTHS, PAST_YEAR, PAST_2_YEARS
  --start-date string     ISO8601 UTC start bound (e.g. 2026-01-01T00:00:00Z)
  --end-date string       ISO8601 UTC end bound
  --streaming             Stream results as they arrive
  --result-type string    ONLY_LINKS or LINKS_WITH_FINAL_SUMMARY (default: LINKS_WITH_FINAL_SUMMARY)
  --count int            Results per source, 10-200 (default: 10)
  --system-message string Override system prompt
  --json                 Output raw JSON (default: formatted human-readable)
  --no-ai                Skip AI completion/summary

Config flags:
  --api-key string       Set API key (get at https://console.desearch.ai)
  --default-tool strings Set default sources
  --show                 Display current config
  --clear                Reset to defaults
```

## Architecture

```
desearch-cli/
├── cmd/
│   ├── root.go          # Root command, config loading
│   ├── search.go        # search command + flags
│   └── completion.go    # completion command (streaming)
├── pkg/
│   ├── api/
│   │   └── client.go    # Desearch API client, request/response types
│   ├── auth/
│   │   └── api_key.go   # API key storage (XDG config dir)
│   └── output/
│       └── formatter.go # Human-readable + JSON output formatters
├── go.mod
├── go.sum
└── README.md
```

### API Reference

**Base URL**: `https://api.desearch.ai`

**Auth**: `Authorization: Bearer <API_KEY>` header

**Endpoint**: `POST /desearch/ai/search`

**Request body**:
```json
{
  "prompt": "string (required)",
  "tools": ["web", "hackernews", "reddit", "wikipedia", "youtube", "twitter", "arxiv"],
  "start_date": "ISO8601 UTC (optional)",
  "end_date": "ISO8601 UTC (optional)",
  "date_filter": "PAST_24_HOURS|PAST_2_DAYS|PAST_WEEK|PAST_2_WEEKS|PAST_MONTH|PAST_2_MONTHS|PAST_YEAR|PAST_2_YEARS",
  "streaming": false,
  "result_type": "ONLY_LINKS|LINKS_WITH_FINAL_SUMMARY",
  "system_message": "string (optional)",
  "scoring_system_message": "string (optional)",
  "count": "10-200"
}
```

**Response**:
```json
{
  "search": [{"link": "", "snippet": "", "title": ""}],
  "hacker_news_search": [...],
  "reddit_search": [...],
  "youtube_search": [...],
  "tweets": [{"id": "", "text": "", "url": "", "user": {...}, "like_count": 0, "retweet_count": 0}],
  "wikipedia_search": [...],
  "text": "string",
  "miner_link_scores": {"url": "HIGH|MEDIUM|LOW"},
  "completion": "AI-generated summary string"
}
```

## Configuration

Config stored at `~/.config/desearch-cli/config.toml` (XDG spec):

```toml
api_key = "YOUR_KEY"
default_tools = ["web", "hackernews", "reddit", "wikipedia", "youtube", "twitter", "arxiv"]
default_date_filter = "PAST_24_HOURS"
default_count = 10
```

## Testing Strategy

- Mock API responses via `httptest` — do not hit real API in tests
- Table-driven tests for output formatters
- Integration test: `go test ./...` with mock server

## Distribution

- **Homebrew**: `roboalchemist/private` tap (goreleaser)
- **Build**: `goreleaser build --snapshot --clean`
- **Single binary**: No external runtime dependencies

## Skill Integration

After install, user can run:
```
/opencode ask "search for latest llama.cpp benchmarks using desearch-cli"
```
Or use as a general-purpose search tool in other agents/workflows.
