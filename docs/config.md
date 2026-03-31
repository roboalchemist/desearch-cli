# Configuration

Desearch CLI uses a config file to store your API key and default search preferences.

## Config File Location

```
~/.config/desearch-cli/config.toml
```

The config directory follows the [XDG Base Directory Specification](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html).

## File Permissions

The config file contains sensitive credentials. Ensure it has mode `0600`:

```bash
chmod 0600 ~/.config/desearch-cli/config.toml
```

## Environment Variable Override

The `DESEARCH_API_KEY` environment variable always takes precedence over the config file:

```bash
export DESEARCH_API_KEY="sk-xxxx"
desearch "query"
```

## Schema (TOML)

```toml
# Required: your Desearch API key
# Get one at https://console.desearch.ai
api_key = "sk-xxxx"

# Optional: sources to query by default
# Valid values: "web", "hackernews", "reddit", "wikipedia", "youtube", "twitter", "arxiv"
default_tools = ["web"]

# Optional: default date range filter
# Valid values: "PAST_24_HOURS", "PAST_2_DAYS", "PAST_WEEK", "PAST_2_WEEKS",
#               "PAST_MONTH", "PAST_2_MONTHS", "PAST_YEAR", "PAST_2_YEARS"
default_date_filter = "PAST_24_HOURS"

# Optional: default number of results per source (10-200)
default_count = 10
```

## JSONC Schema (with comments)

```jsonc
{
  // REQUIRED: Your Desearch API key
  // Get one at https://console.desearch.ai
  "api_key": "sk-xxxx",

  // OPTIONAL: Sources to query by default
  // Type: array of strings
  // Valid values:
  //   "web"        - Web search
  //   "hackernews" - Hacker News
  //   "reddit"     - Reddit
  //   "wikipedia"  - Wikipedia
  //   "youtube"    - YouTube
  //   "twitter"    - Twitter/X
  //   "arxiv"      - arXiv
  // Default: all sources
  "default_tools": ["web", "hackernews", "reddit", "wikipedia", "youtube", "twitter", "arxiv"],

  // OPTIONAL: Default date range filter
  // Type: string
  // Valid values:
  //   "PAST_24_HOURS"   - Last 24 hours
  //   "PAST_2_DAYS"     - Last 2 days
  //   "PAST_WEEK"       - Last 7 days
  //   "PAST_2_WEEKS"    - Last 14 days
  //   "PAST_MONTH"      - Last 30 days
  //   "PAST_2_MONTHS"   - Last 60 days
  //   "PAST_YEAR"       - Last 365 days
  //   "PAST_2_YEARS"    - Last 730 days
  // Default: "PAST_24_HOURS"
  "default_date_filter": "PAST_24_HOURS",

  // OPTIONAL: Default number of results per source
  // Type: integer (10-200)
  // Default: 10
  "default_count": 10
}
```

## Defaults Table

| Field | Default | Valid Range |
|-------|---------|-------------|
| `api_key` | (none - required) | Any valid Desearch API key |
| `default_tools` | All sources | Array of: web, hackernews, reddit, wikipedia, youtube, twitter, arxiv |
| `default_date_filter` | `PAST_24_HOURS` | PAST_24_HOURS, PAST_2_DAYS, PAST_WEEK, PAST_2_WEEKS, PAST_MONTH, PAST_2_MONTHS, PAST_YEAR, PAST_2_YEARS |
| `default_count` | `10` | 10-200 |

## Managing Config

Use the `config` command to manage settings:

```bash
# Set API key
desearch config --api-key sk-xxxx

# Set defaults
desearch config --default-tool web --default-date-filter PAST_WEEK --default-count 20

# View current config
desearch config show

# Clear config (reset to defaults)
desearch config clear
```

## Validation

All fields are optional at the config level. Any unset fields use the API server defaults. The `api_key` must be set either in the config file or via the `DESEARCH_API_KEY` environment variable for any search to work.
