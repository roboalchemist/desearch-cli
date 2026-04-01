# Commands Reference

## `desearch search <query>`

Search the web using DeSearch AI's contextual search engine.

**Flags:**
- `--tool` - Sources to query (web, hackernews, reddit, wikipedia, youtube, twitter, arxiv)
- `--date-filter` - Predefined date range (PAST_24_HOURS, PAST_2_DAYS, PAST_WEEK, PAST_2_WEEKS, PAST_MONTH, PAST_2_MONTHS, PAST_YEAR, PAST_2_YEARS)
- `--start-date` - Start date in ISO8601 UTC format
- `--end-date` - End date in ISO8601 UTC format
- `--streaming` - Stream results as they arrive
- `--result-type` - Result type: ONLY_LINKS or LINKS_WITH_FINAL_SUMMARY
- `--count` - Number of results per source (10-200)
- `--system-message` - System message to influence AI behavior
- `--no-ai` - Skip AI completion/summary

**Examples:**
```bash
desearch "golang best practices"
desearch "rust vs go" --tool web --count 20
desearch "AI news" --date-filter PAST_2_DAYS --streaming
```

## `desearch completion <query>`

Get an AI-generated summary without per-source results.

**Flags:**
- `--system-message` - Optional system message to override the default
- `--json` - Output raw JSON response

**Examples:**
```bash
desearch completion "what is bittensor"
desearch completion "explain transformers" --system-message "Summarize in simple terms"
```

## `desearch config`

Manage the CLI configuration including API key and default search settings.

**Subcommands:**
- `desearch config --show` - Display current configuration
- `desearch config --clear` - Reset configuration to defaults

**Flags:**
- `--api-key` - Set API key
- `--default-tool` - Set default sources (can be specified multiple times)
- `--default-date-filter` - Set date filter (e.g., PAST_24_HOURS, PAST_WEEK, PAST_MONTH)

**Examples:**
```bash
desearch config --api-key sk-xxx
desearch config --show
desearch config --default-tool web --default-tool hackernews
```

## `desearch version`

Print the version number of desearch.

**Example:**
```bash
desearch version
```

## `desearch skill`

Manage Claude Code skill for desearch-cli.

**Subcommands:**
- `desearch skill print` - Print SKILL.md to stdout
- `desearch skill add` - Install skill to ~/.claude/skills/desearch-cli/

**Examples:**
```bash
desearch skill print
desearch skill add
```

## `desearch ai [query]`

Streaming AI completion without per-source search results — streams the AI-generated response as it is generated.

**Flags:**
- `--system-message` - Optional system message to override the default
- `--json` - Output raw JSON response

**Examples:**
```bash
desearch ai "what is bittensor"
desearch ai "explain transformers" --system-message "Summarize in simple terms"
```

## `desearch docs`

Print the embedded README documentation to stdout. Useful for offline reference.

**Example:**
```bash
desearch docs
```

## Global Flags

| Flag | Description |
|------|-------------|
| `--api-key KEY` | API key for authentication (overrides config file) |
| `--config PATH` | Config file path (default ~/.config/desearch-cli/config.toml) |
| `--json` | Output in JSON format |
| `--help`, `-h` | Show help |
| `--version`, `-V` | Show version |
