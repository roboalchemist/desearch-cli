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
						ID:            "123",
						Text:          "This is a tweet",
						URL:           "https://x.com/user/status/123",
						User:          api.TweetUser{Username: "testuser", Name: "Test User"},
						LikeCount:     10,
						RetweetCount:  5,
						ReplyCount:    2,
						QuoteCount:    1,
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
		name      string
		response  *api.SearchResponse
		noAI      bool
		tool      string
		checkFunc func(string) bool
		wantEmpty bool
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
						ID:           "123",
						Text:         "This is a tweet",
						URL:          "https://x.com/user/status/123",
						User:         api.TweetUser{Username: "testuser", Name: "Test User"},
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
						ID:            "123",
						Text:          "Test tweet",
						User:          api.TweetUser{Username: "testuser", Name: "Test User"},
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
		name       string
		flags      OutputFlags
		wantJSON   bool
		wantFields string
	}{
		{
			name:       "json flag returns JSONFormatter",
			flags:      OutputFlags{JSON: true},
			wantJSON:   true,
			wantFields: "",
		},
		{
			name:       "json flag with FilterFields returns JSONFormatter with fields",
			flags:      OutputFlags{JSON: true, FilterFields: "completion,search"},
			wantJSON:   true,
			wantFields: "completion,search",
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
			jf, isJSON := formatter.(*JSONFormatter)
			if isJSON != tt.wantJSON {
				t.Errorf("NewFormatter() isJSON = %v, want %v", isJSON, tt.wantJSON)
			}
			if isJSON && jf.FilterFields != tt.wantFields {
				t.Errorf("NewFormatter().FilterFields = %q, want %q", jf.FilterFields, tt.wantFields)
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
	if flags.FilterFields != "" {
		t.Errorf("OutputFlags.FilterFields = %q, want empty string", flags.FilterFields)
	}
}

func TestPlaintextFormatter_Format(t *testing.T) {
	tests := []struct {
		name      string
		response  *api.SearchResponse
		noAI      bool
		tool      string
		checkFunc func(string) bool
		wantEmpty bool
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
						ID:           "123",
						Text:         "This is a tweet",
						URL:          "https://x.com/user/status/123",
						User:         api.TweetUser{Username: "testuser", Name: "Test User"},
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
					strings.Contains(output, "Web Page\thttps://example.com\tWeb snippet") &&
					strings.Contains(output, "testuser")
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
			},
			tool: "web",
			checkFunc: func(output string) bool {
				return strings.Contains(output, "=== WEB ===") &&
					!strings.Contains(output, "=== HACKERNEWS ===")
			},
		},
		{
			name:      "empty response",
			response:  &api.SearchResponse{},
			wantEmpty: true,
		},
		{
			name: "web results format",
			response: &api.SearchResponse{
				Search: []api.WebResult{
					{Title: "Test", Link: "https://example.com", Snippet: "Snippet"},
				},
			},
			noAI: true,
			checkFunc: func(output string) bool {
				return strings.Contains(output, "Test\thttps://example.com\tSnippet")
			},
		},
		{
			name: "hackernews results format",
			response: &api.SearchResponse{
				HackerNewsSearch: []api.HackerNewsResult{
					{Title: "HN Test", Link: "https://news.ycombinator.com/item?id=1", Snippet: "HN Snippet"},
				},
			},
			noAI: true,
			checkFunc: func(output string) bool {
				return strings.Contains(output, "HN Test\thttps://news.ycombinator.com/item?id=1\tHN Snippet")
			},
		},
		{
			name: "reddit results format",
			response: &api.SearchResponse{
				RedditSearch: []api.RedditResult{
					{Title: "Reddit Test", Link: "https://reddit.com/r/test", Snippet: "Reddit Snippet"},
				},
			},
			noAI: true,
			checkFunc: func(output string) bool {
				return strings.Contains(output, "Reddit Test\thttps://reddit.com/r/test\tReddit Snippet")
			},
		},
		{
			name: "youtube results format",
			response: &api.SearchResponse{
				YoutubeSearch: []api.YoutubeResult{
					{Title: "YT Test", Link: "https://youtube.com/watch?v=1", Snippet: "YT Snippet"},
				},
			},
			noAI: true,
			checkFunc: func(output string) bool {
				return strings.Contains(output, "YT Test\thttps://youtube.com/watch?v=1\tYT Snippet")
			},
		},
		{
			name: "tweet results format",
			response: &api.SearchResponse{
				Tweets: []api.TweetResult{
					{
						ID:   "123",
						Text: "Tweet text",
						User: api.TweetUser{Username: "tweetuser", Name: "Tweet User"},
					},
				},
			},
			noAI: true,
			checkFunc: func(output string) bool {
				return strings.Contains(output, "tweetuser")
			},
		},
		{
			name: "wikipedia results format",
			response: &api.SearchResponse{
				WikipediaSearch: []api.WikipediaResult{
					{Title: "Wiki Test", Link: "https://en.wikipedia.org/wiki/Test", Snippet: "Wiki Snippet"},
				},
			},
			noAI: true,
			checkFunc: func(output string) bool {
				return strings.Contains(output, "Wiki Test\thttps://en.wikipedia.org/wiki/Test\tWiki Snippet")
			},
		},
		{
			name: "arxiv results format",
			response: &api.SearchResponse{
				ArxivSearch: []api.ArxivResult{
					{Title: "ArXiv Test", Link: "https://arxiv.org/abs/1234", Snippet: "ArXiv Snippet"},
				},
			},
			noAI: true,
			checkFunc: func(output string) bool {
				return strings.Contains(output, "ArXiv Test\thttps://arxiv.org/abs/1234\tArXiv Snippet")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := &PlaintextFormatter{NoAI: tt.noAI, Tool: tt.tool}
			output := formatter.Format(tt.response)

			if tt.wantEmpty {
				if output != "" {
					t.Errorf("PlaintextFormatter.Format() = %q, want empty string", output)
				}
				return
			}

			if tt.checkFunc != nil && !tt.checkFunc(output) {
				t.Errorf("PlaintextFormatter.Format() = %q, failed check function", output)
			}
		})
	}
}

func TestEvaluateJQ(t *testing.T) {
	tests := []struct {
		name        string
		data        string
		expression  string
		wantErr     bool
		wantContain string
	}{
		{
			name:        "empty expression returns original",
			data:        `{"key": "value"}`,
			expression:  "",
			wantContain: "key",
		},
		{
			name:        "simple key access",
			data:        `{"key": "value"}`,
			expression:  ".key",
			wantContain: "value",
		},
		{
			name:        "nested key access",
			data:        `{"outer": {"inner": "nested value"}}`,
			expression:  ".outer.inner",
			wantContain: "nested value",
		},
		{
			name:        "array access",
			data:        `{"items": ["a", "b", "c"]}`,
			expression:  ".items[]",
			wantContain: `"a"`,
		},
		{
			name:        "nonexistent key returns null",
			data:        `{"key": "value"}`,
			expression:  ".nonexistent",
			wantContain: "null",
		},
		{
			name:       "invalid jq syntax",
			data:       `{"key": "value"}`,
			expression: "[invalid",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := EvaluateJQ([]byte(tt.data), tt.expression)

			if tt.wantErr {
				if err == nil {
					t.Errorf("EvaluateJQ() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("EvaluateJQ() unexpected error: %v", err)
				return
			}

			if tt.wantContain != "" && !strings.Contains(string(result), tt.wantContain) {
				t.Errorf("EvaluateJQ() = %q, want containing %q", string(result), tt.wantContain)
			}
		})
	}
}

func TestEvaluateJQ_InvalidJSON(t *testing.T) {
	_, err := EvaluateJQ([]byte("not json"), ".key")
	if err == nil {
		t.Error("EvaluateJQ() expected error for invalid JSON, got nil")
	}
}

func TestFilterJSONFields(t *testing.T) {
	tests := []struct {
		name        string
		data        string
		fields      string
		wantErr     bool
		wantKeys    []string
		wantMissing []string
	}{
		{
			name:     "single field",
			data:     `{"search": [{"title": "Test"}], "completion": "summary", "reddit_search": []}`,
			fields:   "completion",
			wantKeys: []string{"completion"},
		},
		{
			name:        "multiple fields",
			data:        `{"search": [{"title": "Test"}], "completion": "summary", "reddit_search": []}`,
			fields:      "completion,search",
			wantKeys:    []string{"completion", "search"},
			wantMissing: []string{"reddit_search"},
		},
		{
			name:        "fields with spaces trimmed",
			data:        `{"search": [], "completion": "text", "hacker_news_search": []}`,
			fields:      " completion , search ",
			wantKeys:    []string{"completion", "search"},
			wantMissing: []string{"hacker_news_search"},
		},
		{
			name:     "unknown field returns empty object",
			data:     `{"search": []}`,
			fields:   "nonexistent",
			wantKeys: []string{},
		},
		{
			name:    "invalid JSON",
			data:    "not json",
			fields:  "completion",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := FilterJSONFields([]byte(tt.data), tt.fields)

			if tt.wantErr {
				if err == nil {
					t.Errorf("FilterJSONFields() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("FilterJSONFields() unexpected error: %v", err)
				return
			}

			var parsed map[string]interface{}
			if err := json.Unmarshal(result, &parsed); err != nil {
				t.Errorf("FilterJSONFields() produced invalid JSON: %v", err)
				return
			}

			for _, k := range tt.wantKeys {
				if _, ok := parsed[k]; !ok {
					t.Errorf("FilterJSONFields() missing expected key %q in output", k)
				}
			}

			for _, k := range tt.wantMissing {
				if _, ok := parsed[k]; ok {
					t.Errorf("FilterJSONFields() should not contain key %q but got it", k)
				}
			}
		})
	}
}

func TestJSONFormatter_Format_FilterFields(t *testing.T) {
	resp := &api.SearchResponse{
		Search: []api.WebResult{
			{Title: "Test Page", Link: "https://example.com", Snippet: "Test snippet"},
		},
		RedditSearch: []api.RedditResult{
			{Title: "Reddit Post", Link: "https://reddit.com/r/test", Snippet: "Reddit snippet"},
		},
		Completion: "AI summary text",
	}

	t.Run("filters to single field", func(t *testing.T) {
		formatter := &JSONFormatter{FilterFields: "completion"}
		output := formatter.Format(resp)

		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(output), &parsed); err != nil {
			t.Fatalf("JSONFormatter.Format() produced invalid JSON: %v", err)
		}

		if _, ok := parsed["completion"]; !ok {
			t.Errorf("JSONFormatter.Format() missing 'completion' key")
		}
		if _, ok := parsed["search"]; ok {
			t.Errorf("JSONFormatter.Format() should not contain 'search' key")
		}
	})

	t.Run("filters to multiple fields", func(t *testing.T) {
		formatter := &JSONFormatter{FilterFields: "search,completion"}
		output := formatter.Format(resp)

		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(output), &parsed); err != nil {
			t.Fatalf("JSONFormatter.Format() produced invalid JSON: %v", err)
		}

		if _, ok := parsed["search"]; !ok {
			t.Errorf("JSONFormatter.Format() missing 'search' key")
		}
		if _, ok := parsed["completion"]; !ok {
			t.Errorf("JSONFormatter.Format() missing 'completion' key")
		}
		if _, ok := parsed["reddit_search"]; ok {
			t.Errorf("JSONFormatter.Format() should not contain 'reddit_search' key")
		}
	})

	t.Run("empty FilterFields returns all fields", func(t *testing.T) {
		formatter := &JSONFormatter{FilterFields: ""}
		output := formatter.Format(resp)

		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(output), &parsed); err != nil {
			t.Fatalf("JSONFormatter.Format() produced invalid JSON: %v", err)
		}

		// All fields should be present
		if _, ok := parsed["search"]; !ok {
			t.Errorf("JSONFormatter.Format() missing 'search' key")
		}
		if _, ok := parsed["reddit_search"]; !ok {
			t.Errorf("JSONFormatter.Format() missing 'reddit_search' key")
		}
		if _, ok := parsed["completion"]; !ok {
			t.Errorf("JSONFormatter.Format() missing 'completion' key")
		}
	})
}
