package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/itchyny/gojq"
	"github.com/roboalchemist/desearch-cli/pkg/api"
)

func init() {
	// Respect the NO_COLOR environment variable (https://no-color.org/).
	// The formatter currently uses plain text (strings.Builder) with no ANSI color
	// codes, so NO_COLOR is trivially satisfied. This init() is here to ensure
	// correct behavior if color support is added in the future via fatih/color.
	// When that library is imported, set color.NoColor = true here.
	if os.Getenv("NO_COLOR") != "" {
		// color.NoColor = true  // uncomment when fatih/color is imported
	}
}

// Formatter defines an interface for formatting search responses.
type Formatter interface {
	Format(*api.SearchResponse) string
}

// OutputFlags holds the flags that control output formatting.
type OutputFlags struct {
	JSON        bool
	NoAI        bool
	Tool        string // empty means all tools
	Plaintext   bool
	FilterFields string // comma-separated top-level field names to include in JSON output
}

// NewFormatter returns the appropriate formatter based on flags.
func NewFormatter(flags OutputFlags) Formatter {
	if flags.JSON {
		return &JSONFormatter{FilterFields: flags.FilterFields}
	}
	if flags.Plaintext {
		return &PlaintextFormatter{NoAI: flags.NoAI, Tool: flags.Tool}
	}
	return &HumanFormatter{NoAI: flags.NoAI, Tool: flags.Tool}
}

// JSONFormatter outputs raw JSON with json.MarshalIndent.
type JSONFormatter struct {
	FilterFields string // comma-separated top-level field names to include
}

// Format returns the JSON representation of the search response,
// optionally filtered to only include the top-level fields specified.
func (f *JSONFormatter) Format(resp *api.SearchResponse) string {
	data, err := resp.MarshalJSON()
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to marshal response: %v"}`, err)
	}
	if f.FilterFields != "" {
		filtered, err := FilterJSONFields(data, f.FilterFields)
		if err != nil {
			return fmt.Sprintf(`{"error": "failed to filter fields: %v"}`, err)
		}
		return string(filtered)
	}
	return string(data)
}

// HumanFormatter pretty-prints each source section with headers.
type HumanFormatter struct {
	NoAI bool
	Tool string // empty means all tools
}

// Format returns a human-readable formatted string of the search response.
func (f *HumanFormatter) Format(resp *api.SearchResponse) string {
	var sb strings.Builder

	// Define source sections in order
	sources := []struct {
		key      string
		name     string
		results  interface{}
		canCheck bool
	}{
		{"search", "WEB", resp.Search, true},
		{"hacker_news_search", "HACKERNEWS", resp.HackerNewsSearch, true},
		{"reddit_search", "REDDIT", resp.RedditSearch, true},
		{"youtube_search", "YOUTUBE", resp.YoutubeSearch, true},
		{"tweets", "TWITTER", resp.Tweets, true},
		{"wikipedia_search", "WIKIPEDIA", resp.WikipediaSearch, true},
		{"arxiv_search", "ARXIV", resp.ArxivSearch, true},
	}

	for _, src := range sources {
		// Skip if a specific tool filter is set and doesn't match
		if f.Tool != "" && !strings.EqualFold(f.Tool, src.key) &&
			!strings.EqualFold(f.Tool, src.name) &&
			!strings.EqualFold(f.Tool, strings.TrimSuffix(src.name, "S")) {
			continue
		}

		if !src.canCheck {
			continue
		}

		switch r := src.results.(type) {
		case []api.WebResult:
			if len(r) > 0 {
				f.writeWebResults(&sb, src.name, r)
			}
		case []api.HackerNewsResult:
			if len(r) > 0 {
				f.writeHackerNewsResults(&sb, src.name, r)
			}
		case []api.RedditResult:
			if len(r) > 0 {
				f.writeRedditResults(&sb, src.name, r)
			}
		case []api.YoutubeResult:
			if len(r) > 0 {
				f.writeYoutubeResults(&sb, src.name, r)
			}
		case []api.TweetResult:
			if len(r) > 0 {
				f.writeTweetResults(&sb, src.name, r)
			}
		case []api.WikipediaResult:
			if len(r) > 0 {
				f.writeWikipediaResults(&sb, src.name, r)
			}
		case []api.ArxivResult:
			if len(r) > 0 {
				f.writeArxivResults(&sb, src.name, r)
			}
		}
	}

	// AI Summary section
	if !f.NoAI && resp.Completion != "" {
		sb.WriteString("=== AI SUMMARY ===\n")
		sb.WriteString(resp.Completion)
		sb.WriteString("\n")
	}

	return sb.String()
}

func (f *HumanFormatter) writeWebResults(sb *strings.Builder, header string, results []api.WebResult) {
	sb.WriteString(fmt.Sprintf("=== %s ===\n", header))
	for _, r := range results {
		sb.WriteString(fmt.Sprintf("[%s](%s)\n", r.Title, r.Link))
		if r.Snippet != "" {
			sb.WriteString(fmt.Sprintf("  %s\n", r.Snippet))
		}
		sb.WriteString("\n")
	}
}

func (f *HumanFormatter) writeHackerNewsResults(sb *strings.Builder, header string, results []api.HackerNewsResult) {
	sb.WriteString(fmt.Sprintf("=== %s ===\n", header))
	for _, r := range results {
		sb.WriteString(fmt.Sprintf("[%s](%s)\n", r.Title, r.Link))
		if r.Snippet != "" {
			sb.WriteString(fmt.Sprintf("  %s\n", r.Snippet))
		}
		sb.WriteString("\n")
	}
}

func (f *HumanFormatter) writeRedditResults(sb *strings.Builder, header string, results []api.RedditResult) {
	sb.WriteString(fmt.Sprintf("=== %s ===\n", header))
	for _, r := range results {
		sb.WriteString(fmt.Sprintf("[%s](%s)\n", r.Title, r.Link))
		if r.Snippet != "" {
			sb.WriteString(fmt.Sprintf("  %s\n", r.Snippet))
		}
		sb.WriteString("\n")
	}
}

func (f *HumanFormatter) writeYoutubeResults(sb *strings.Builder, header string, results []api.YoutubeResult) {
	sb.WriteString(fmt.Sprintf("=== %s ===\n", header))
	for _, r := range results {
		sb.WriteString(fmt.Sprintf("[%s](%s)\n", r.Title, r.Link))
		if r.Snippet != "" {
			sb.WriteString(fmt.Sprintf("  %s\n", r.Snippet))
		}
		sb.WriteString("\n")
	}
}

func (f *HumanFormatter) writeTweetResults(sb *strings.Builder, header string, results []api.TweetResult) {
	sb.WriteString(fmt.Sprintf("=== %s ===\n", header))
	for _, t := range results {
		if t.User.Username != "" {
			sb.WriteString(fmt.Sprintf("@%s\n", t.User.Username))
		}
		sb.WriteString(fmt.Sprintf("  %s\n", t.Text))
		if t.URL != "" {
			sb.WriteString(fmt.Sprintf("  Link: %s\n", t.URL))
		}
		// Engagement metrics
		metrics := []string{}
		if t.LikeCount > 0 {
			metrics = append(metrics, fmt.Sprintf("%d likes", t.LikeCount))
		}
		if t.RetweetCount > 0 {
			metrics = append(metrics, fmt.Sprintf("%d retweets", t.RetweetCount))
		}
		if t.ReplyCount > 0 {
			metrics = append(metrics, fmt.Sprintf("%d replies", t.ReplyCount))
		}
		if t.QuoteCount > 0 {
			metrics = append(metrics, fmt.Sprintf("%d quotes", t.QuoteCount))
		}
		if t.BookmarkCount > 0 {
			metrics = append(metrics, fmt.Sprintf("%d bookmarks", t.BookmarkCount))
		}
		if len(metrics) > 0 {
			sb.WriteString(fmt.Sprintf("  %s\n", strings.Join(metrics, ", ")))
		}
		sb.WriteString("\n")
	}
}

func (f *HumanFormatter) writeWikipediaResults(sb *strings.Builder, header string, results []api.WikipediaResult) {
	sb.WriteString(fmt.Sprintf("=== %s ===\n", header))
	for _, r := range results {
		sb.WriteString(fmt.Sprintf("[%s](%s)\n", r.Title, r.Link))
		if r.Snippet != "" {
			sb.WriteString(fmt.Sprintf("  %s\n", r.Snippet))
		}
		sb.WriteString("\n")
	}
}

func (f *HumanFormatter) writeArxivResults(sb *strings.Builder, header string, results []api.ArxivResult) {
	sb.WriteString(fmt.Sprintf("=== %s ===\n", header))
	for _, r := range results {
		sb.WriteString(fmt.Sprintf("[%s](%s)\n", r.Title, r.Link))
		if r.Snippet != "" {
			sb.WriteString(fmt.Sprintf("  %s\n", r.Snippet))
		}
		sb.WriteString("\n")
	}
}

// PlaintextFormatter outputs tab-separated values (title TAB url TAB snippet per line,
// section headers as lines starting with ===).
type PlaintextFormatter struct {
	NoAI bool
	Tool string // empty means all tools
}

// Format returns a tab-separated formatted string of the search response.
func (f *PlaintextFormatter) Format(resp *api.SearchResponse) string {
	var sb strings.Builder

	// Define source sections in order
	sources := []struct {
		key      string
		name     string
		results  interface{}
		canCheck bool
	}{
		{"search", "WEB", resp.Search, true},
		{"hacker_news_search", "HACKERNEWS", resp.HackerNewsSearch, true},
		{"reddit_search", "REDDIT", resp.RedditSearch, true},
		{"youtube_search", "YOUTUBE", resp.YoutubeSearch, true},
		{"tweets", "TWITTER", resp.Tweets, true},
		{"wikipedia_search", "WIKIPEDIA", resp.WikipediaSearch, true},
		{"arxiv_search", "ARXIV", resp.ArxivSearch, true},
	}

	for _, src := range sources {
		// Skip if a specific tool filter is set and doesn't match
		if f.Tool != "" && !strings.EqualFold(f.Tool, src.key) &&
			!strings.EqualFold(f.Tool, src.name) &&
			!strings.EqualFold(f.Tool, strings.TrimSuffix(src.name, "S")) {
			continue
		}

		if !src.canCheck {
			continue
		}

		switch r := src.results.(type) {
		case []api.WebResult:
			if len(r) > 0 {
				f.writeWebResults(&sb, src.name, r)
			}
		case []api.HackerNewsResult:
			if len(r) > 0 {
				f.writeHackerNewsResults(&sb, src.name, r)
			}
		case []api.RedditResult:
			if len(r) > 0 {
				f.writeRedditResults(&sb, src.name, r)
			}
		case []api.YoutubeResult:
			if len(r) > 0 {
				f.writeYoutubeResults(&sb, src.name, r)
			}
		case []api.TweetResult:
			if len(r) > 0 {
				f.writeTweetResults(&sb, src.name, r)
			}
		case []api.WikipediaResult:
			if len(r) > 0 {
				f.writeWikipediaResults(&sb, src.name, r)
			}
		case []api.ArxivResult:
			if len(r) > 0 {
				f.writeArxivResults(&sb, src.name, r)
			}
		}
	}

	// AI Summary section
	if !f.NoAI && resp.Completion != "" {
		sb.WriteString("=== AI SUMMARY ===\n")
		sb.WriteString(resp.Completion)
		sb.WriteString("\n")
	}

	return sb.String()
}

func (f *PlaintextFormatter) writeWebResults(sb *strings.Builder, header string, results []api.WebResult) {
	sb.WriteString(fmt.Sprintf("=== %s ===\n", header))
	for _, r := range results {
		sb.WriteString(fmt.Sprintf("%s\t%s\t%s\n", r.Title, r.Link, r.Snippet))
	}
}

func (f *PlaintextFormatter) writeHackerNewsResults(sb *strings.Builder, header string, results []api.HackerNewsResult) {
	sb.WriteString(fmt.Sprintf("=== %s ===\n", header))
	for _, r := range results {
		sb.WriteString(fmt.Sprintf("%s\t%s\t%s\n", r.Title, r.Link, r.Snippet))
	}
}

func (f *PlaintextFormatter) writeRedditResults(sb *strings.Builder, header string, results []api.RedditResult) {
	sb.WriteString(fmt.Sprintf("=== %s ===\n", header))
	for _, r := range results {
		sb.WriteString(fmt.Sprintf("%s\t%s\t%s\n", r.Title, r.Link, r.Snippet))
	}
}

func (f *PlaintextFormatter) writeYoutubeResults(sb *strings.Builder, header string, results []api.YoutubeResult) {
	sb.WriteString(fmt.Sprintf("=== %s ===\n", header))
	for _, r := range results {
		sb.WriteString(fmt.Sprintf("%s\t%s\t%s\n", r.Title, r.Link, r.Snippet))
	}
}

func (f *PlaintextFormatter) writeTweetResults(sb *strings.Builder, header string, results []api.TweetResult) {
	sb.WriteString(fmt.Sprintf("=== %s ===\n", header))
	for _, t := range results {
		sb.WriteString(fmt.Sprintf("%s\t%s\t%s\n", t.User.Username, t.URL, t.Text))
	}
}

func (f *PlaintextFormatter) writeWikipediaResults(sb *strings.Builder, header string, results []api.WikipediaResult) {
	sb.WriteString(fmt.Sprintf("=== %s ===\n", header))
	for _, r := range results {
		sb.WriteString(fmt.Sprintf("%s\t%s\t%s\n", r.Title, r.Link, r.Snippet))
	}
}

func (f *PlaintextFormatter) writeArxivResults(sb *strings.Builder, header string, results []api.ArxivResult) {
	sb.WriteString(fmt.Sprintf("=== %s ===\n", header))
	for _, r := range results {
		sb.WriteString(fmt.Sprintf("%s\t%s\t%s\n", r.Title, r.Link, r.Snippet))
	}
}

// StreamingFormatter writes streamed completion chunks directly to stdout.
// Unlike other formatters it does not buffer output, since chunks must be
// printed immediately for a responsive streaming experience.
type StreamingFormatter struct {
	JSON bool // if true, outputs accumulated text as a JSON object at the end
}

// WriteChunk prints a completion token to stdout immediately, with no extra
// whitespace. It returns the text appended to the internal buffer.
func (f *StreamingFormatter) WriteChunk(chunk string) {
	fmt.Print(chunk)
	os.Stdout.Sync()
}

// Finalize outputs the accumulated text as a final newline (plaintext) or as
// a JSON object (JSON mode). It returns the string written so callers can
// include it in combined output if needed.
func (f *StreamingFormatter) Finalize(query, text string) string {
	if f.JSON {
		output := map[string]string{
			"query":      query,
			"completion": text,
		}
		data, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			// Fall back to plaintext on marshal error
			fmt.Println()
			return ""
		}
		fmt.Println(string(data))
		return string(data)
	}
	fmt.Println()
	return ""
}

// FilterJSONFields filters a JSON object to only include the specified top-level fields.
// fields is a comma-separated list of field names (JSON keys).
// Returns the filtered JSON with the same indentation as the input.
func FilterJSONFields(data []byte, fields string) ([]byte, error) {
	var v map[string]interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Parse the comma-separated field list
	requested := strings.Split(fields, ",")
	allowed := make(map[string]bool, len(requested))
	for _, f := range requested {
		allowed[strings.TrimSpace(f)] = true
	}

	// Build a new map with only the requested fields, preserving original keys
	filtered := make(map[string]interface{})
	for k, v := range v {
		if allowed[k] {
			filtered[k] = v
		}
	}

	return json.MarshalIndent(filtered, "", "  ")
}

// EvaluateJQ runs a jq expression on the given JSON bytes and returns the filtered result.
// If the expression is empty, returns the original JSON.
func EvaluateJQ(data []byte, expression string) ([]byte, error) {
	if expression == "" {
		return data, nil
	}

	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	query, err := gojq.Parse(expression)
	if err != nil {
		return nil, fmt.Errorf("failed to parse jq expression: %w", err)
	}

	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "  ")

	iter := query.Run(v)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return nil, fmt.Errorf("jq evaluation error: %w", err)
		}
		if err := encoder.Encode(v); err != nil {
			return nil, fmt.Errorf("failed to encode jq result: %w", err)
		}
	}

	return buf.Bytes(), nil
}
