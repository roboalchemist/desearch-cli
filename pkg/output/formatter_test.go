package output

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/roboalchemist/desearch-cli/pkg/api"
)

func TestJSONFormatter_Format(t *testing.T) {
	tests := []struct {
		name     string
		response *api.SearchResponse
		wantJSON bool
	}{
		{
			name: "full response",
			response: &api.SearchResponse{
				Search: []api.WebResult{
					{Title: "Test Page", Link: "https://example.com", Snippet: "Test snippet"},
				},
				HackerNewsSearch: []api.HackerNewsResult{
					{Title: "HN Post", Link: "https://news.ycombinator.com/item?id=1", Snippet: "HN snippet"},
				},
				RedditSearch: []api.RedditResult{
					{Title: "Reddit Post", Link: "https://reddit.com/r/test", Snippet: "Reddit snippet"},
				},
				YoutubeSearch: []api.YoutubeResult{
					{Title: "YouTube Video", Link: "https://youtube.com/watch?v=1", Snippet: "Video snippet"},
				},
				Tweets: []api.TweetResult{
					{
						ID: "123",
						Text: "This is a tweet",
						URL: "https://x.com/user/status/123",
						User: api.TweetUser{Username: "testuser", Name: "Test User"},
						LikeCount:    10,
						RetweetCount: 5,
						ReplyCount:   2,
						QuoteCount:   1,
						BookmarkCount: 3,
					},
				},
				WikipediaSearch: []api.WikipediaResult{
					{Title: "Wikipedia Article", Link: "https://en.wikipedia.org/wiki/Test", Snippet: "Wikipedia snippet"},
				},
				ArxivSearch: []api.ArxivResult{
					{Title: "ArXiv Paper", Link: "https://arxiv.org/abs/2101.00001", Snippet: "ArXiv snippet"},
				},
				Completion: "AI summary text",
			},
			wantJSON: true,
		},
		{
			name: "empty response",
			response: &api.SearchResponse{
				Search: []api.WebResult{},
			},
			wantJSON: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := &JSONFormatter{}
			output := formatter.Format(tt.response)

			if tt.wantJSON {
				// Verify it's valid JSON
				var decoded api.SearchResponse
				if err := json.Unmarshal([]byte(output), &decoded); err != nil {
					t.Errorf("JSONFormatter.Format() produced invalid JSON: %v", err)
				}
			}
		})
	}
}

func TestHumanFormatter_Format(t *testing.T) {
	tests := []struct {
		name       string
		response   *api.SearchResponse
		noAI       bool
		tool       string
		checkFunc  func(string) bool
		wantEmpty  bool
	}{
		{
			name: "full response with all sections",
			response: &api.SearchResponse{
				Search: []api.WebResult{
					{Title: "Web Page", Link: "https://example.com", Snippet: "Web snippet"},
				},
				HackerNewsSearch: []api.HackerNewsResult{
					{Title: "HN Post", Link: "https://news.ycombinator.com/item?id=1", Snippet: "HN snippet"},
				},
				RedditSearch: []api.RedditResult{
					{Title: "Reddit Post", Link: "https://reddit.com/r/test", Snippet: "Reddit snippet"},
				},
				YoutubeSearch: []api.YoutubeResult{
					{Title: "YouTube Video", Link: "https://youtube.com/watch?v=1", Snippet: "Video snippet"},
				},
				Tweets: []api.TweetResult{
					{
						ID: "123",
						Text: "This is a tweet",
						URL: "https://x.com/user/status/123",
						User: api.TweetUser{Username: "testuser", Name: "Test User"},
						LikeCount:    10,
						RetweetCount: 5,
						ReplyCount:   2,
					},
				},
				WikipediaSearch: []api.WikipediaResult{
					{Title: "Wikipedia Article", Link: "https://en.wikipedia.org/wiki/Test", Snippet: "Wikipedia snippet"},
				},
				ArxivSearch: []api.ArxivResult{
					{Title: "ArXiv Paper", Link: "https://arxiv.org/abs/2101.00001", Snippet: "ArXiv snippet"},
				},
				Completion: "AI summary text",
			},
			noAI: false,
			checkFunc: func(output string) bool {
				return strings.Contains(output, "=== WEB ===") &&
					strings.Contains(output, "=== HACKERNEWS ===") &&
					strings.Contains(output, "=== REDDIT ===") &&
					strings.Contains(output, "=== YOUTUBE ===") &&
					strings.Contains(output, "=== TWITTER ===") &&
					strings.Contains(output, "=== WIKIPEDIA ===") &&
					strings.Contains(output, "=== ARXIV ===") &&
					strings.Contains(output, "=== AI SUMMARY ===") &&
					strings.Contains(output, "[Web Page](https://example.com)") &&
					strings.Contains(output, "@testuser") &&
					strings.Contains(output, "10 likes, 5 retweets, 2 replies")
			},
		},
		{
			name: "no-ai flag hides AI summary",
			response: &api.SearchResponse{
				Search: []api.WebResult{
					{Title: "Web Page", Link: "https://example.com", Snippet: "Web snippet"},
				},
				Completion: "AI summary text",
			},
			noAI: true,
			checkFunc: func(output string) bool {
				return strings.Contains(output, "=== WEB ===") &&
					!strings.Contains(output, "=== AI SUMMARY ===")
			},
		},
		{
			name: "tool filter shows only web",
			response: &api.SearchResponse{
				Search: []api.WebResult{
					{Title: "Web Page", Link: "https://example.com", Snippet: "Web snippet"},
				},
				HackerNewsSearch: []api.HackerNewsResult{
					{Title: "HN Post", Link: "https://news.ycombinator.com/item?id=1", Snippet: "HN snippet"},
				},
				RedditSearch: []api.RedditResult{
					{Title: "Reddit Post", Link: "https://reddit.com/r/test", Snippet: "Reddit snippet"},
				},
			},
			tool: "web",
			checkFunc: func(output string) bool {
				return strings.Contains(output, "=== WEB ===") &&
					!strings.Contains(output, "=== HACKERNEWS ===") &&
					!strings.Contains(output, "=== REDDIT ===")
			},
		},
		{
			name: "tool filter case insensitive",
			response: &api.SearchResponse{
				Search: []api.WebResult{
					{Title: "Web Page", Link: "https://example.com", Snippet: "Web snippet"},
				},
			},
			tool: "WEB",
			checkFunc: func(output string) bool {
				return strings.Contains(output, "=== WEB ===")
			},
		},
		{
			name: "tool filter with singular form",
			response: &api.SearchResponse{
				HackerNewsSearch: []api.HackerNewsResult{
					{Title: "HN Post", Link: "https://news.ycombinator.com/item?id=1", Snippet: "HN snippet"},
				},
			},
			tool: "hackernews",
			checkFunc: func(output string) bool {
				return strings.Contains(output, "=== HACKERNEWS ===")
			},
		},
		{
			name: "empty response",
			response: &api.SearchResponse{
				Search: []api.WebResult{},
			},
			wantEmpty: true,
		},
		{
			name: "tweet with all metrics",
			response: &api.SearchResponse{
				Tweets: []api.TweetResult{
					{
						ID: "123",
						Text: "Test tweet",
						User: api.TweetUser{Username: "testuser", Name: "Test User"},
						LikeCount:     100,
						RetweetCount:  50,
						ReplyCount:    25,
						QuoteCount:    10,
						BookmarkCount: 5,
					},
				},
			},
			noAI: true,
			checkFunc: func(output string) bool {
				return strings.Contains(output, "100 likes") &&
					strings.Contains(output, "50 retweets") &&
					strings.Contains(output, "25 replies") &&
					strings.Contains(output, "10 quotes") &&
					strings.Contains(output, "5 bookmarks")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := &HumanFormatter{NoAI: tt.noAI, Tool: tt.tool}
			output := formatter.Format(tt.response)

			if tt.wantEmpty {
				if output != "" {
					t.Errorf("HumanFormatter.Format() = %q, want empty string", output)
				}
				return
			}

			if tt.checkFunc != nil && !tt.checkFunc(output) {
				t.Errorf("HumanFormatter.Format() = %q, failed check function", output)
			}
		})
	}
}

func TestNewFormatter(t *testing.T) {
	tests := []struct {
		name     string
		flags    OutputFlags
		wantJSON bool
	}{
		{
			name:     "json flag returns JSONFormatter",
			flags:    OutputFlags{JSON: true},
			wantJSON: true,
		},
		{
			name:     "no json flag returns HumanFormatter",
			flags:    OutputFlags{JSON: false},
			wantJSON: false,
		},
		{
			name:     "default flags returns HumanFormatter",
			flags:    OutputFlags{},
			wantJSON: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewFormatter(tt.flags)
			_, isJSON := formatter.(*JSONFormatter)
			if isJSON != tt.wantJSON {
				t.Errorf("NewFormatter() isJSON = %v, want %v", isJSON, tt.wantJSON)
			}
		})
	}
}

func TestOutputFlags_Defaults(t *testing.T) {
	flags := OutputFlags{}
	if flags.JSON != false {
		t.Errorf("OutputFlags.JSON = %v, want false", flags.JSON)
	}
	if flags.NoAI != false {
		t.Errorf("OutputFlags.NoAI = %v, want false", flags.NoAI)
	}
	if flags.Tool != "" {
		t.Errorf("OutputFlags.Tool = %q, want empty string", flags.Tool)
	}
}
