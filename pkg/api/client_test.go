package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	client := NewClient("test-api-key")
	if client == nil {
		t.Fatal("NewClient returned nil")
	}
	if client.BaseURL != BaseURL {
		t.Errorf("expected BaseURL %q, got %q", BaseURL, client.BaseURL)
	}
	if client.APIKey != "test-api-key" {
		t.Errorf("expected APIKey %q, got %q", "test-api-key", client.APIKey)
	}
	if client.HTTPClient == nil {
		t.Error("expected non-nil HTTPClient")
	}
}

func TestSearchRequest_Marshal(t *testing.T) {
	streaming := true
	count := 50
	req := SearchRequest{
		Prompt:      "Bittensor",
		Tools:       []string{"web", "hackernews"},
		DateFilter:  ptrString("PAST_WEEK"),
		Streaming:   &streaming,
		ResultType:  ptrString("LINKS_WITH_FINAL_SUMMARY"),
		SystemMessage: ptrString("Summarize in pros and cons"),
		Count:       &count,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded SearchRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Prompt != req.Prompt {
		t.Errorf("prompt: expected %q, got %q", req.Prompt, decoded.Prompt)
	}
	if len(decoded.Tools) != len(req.Tools) {
		t.Errorf("tools length: expected %d, got %d", len(req.Tools), len(decoded.Tools))
	}
	if *decoded.Count != *req.Count {
		t.Errorf("count: expected %d, got %d", *req.Count, *decoded.Count)
	}
}

func TestSearchResponse_RoundTrip(t *testing.T) {
	resp := SearchResponse{
		HackerNewsSearch: []HackerNewsResult{
			{Title: "Bittensor on HN", Link: "https://news.ycombinator.com/item?id=1", Snippet: "Test snippet"},
		},
		RedditSearch: []RedditResult{
			{Title: "Bittensor on Reddit", Link: "https://reddit.com/r/bittensor", Snippet: "Test snippet"},
		},
		Search: []WebResult{
			{Title: "What is Bittensor?", Link: "https://example.com/bittensor", Snippet: "Test snippet"},
		},
		YoutubeSearch: []YoutubeResult{
			{Title: "Bittensor Video", Link: "https://youtube.com/watch?v=1", Snippet: "Test snippet"},
		},
		Tweets: []TweetResult{
			{
				ID: "123", Text: "Excited about Bittensor!",
				User: TweetUser{Username: "testuser", Name: "Test User"},
				LikeCount: 10, RetweetCount: 2,
			},
		},
		WikipediaSearch: []WikipediaResult{
			{Title: "Bittensor", Link: "https://en.wikipedia.org/wiki/Bittensor", Snippet: "Test snippet"},
		},
		ArxivSearch: []ArxivResult{
			{Title: "Bittensor Paper", Link: "https://arxiv.org/abs/1234", Snippet: "Test snippet"},
		},
		Text:            "Some additional text.",
		MinerLinkScores: map[string]string{"https://example.com": "HIGH"},
		Completion:      "Bittensor is a decentralized AI network.",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded SearchResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(decoded.HackerNewsSearch) != 1 {
		t.Errorf("hacker_news_search length: expected 1, got %d", len(decoded.HackerNewsSearch))
	}
	if decoded.Completion != resp.Completion {
		t.Errorf("completion: expected %q, got %q", resp.Completion, decoded.Completion)
	}
	if decoded.Text != resp.Text {
		t.Errorf("text: expected %q, got %q", resp.Text, decoded.Text)
	}
}

func TestSearch_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/desearch/ai/search" {
			t.Errorf("expected path /desearch/ai/search, got %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			t.Errorf("expected Authorization header starting with 'Bearer ', got %s", auth)
		}

		var req SearchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
		if req.Prompt != "test query" {
			t.Errorf("expected prompt 'test query', got %q", req.Prompt)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(SearchResponse{
			HackerNewsSearch: []HackerNewsResult{
				{Title: "Test Result", Link: "https://example.com", Snippet: "Test"},
			},
		})
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL, APIKey: "test-key", HTTPClient: server.Client()}
	streaming := false
	resp, err := client.Search(context.Background(), &SearchRequest{
		Prompt:    "test query",
		Tools:     []string{"web"},
		Streaming: &streaming,
	})
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(resp.HackerNewsSearch) != 1 {
		t.Errorf("expected 1 result, got %d", len(resp.HackerNewsSearch))
	}
}

func TestSearch_Non200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"detail": "Invalid API key"}`))
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL, APIKey: "bad-key", HTTPClient: server.Client()}
	_, err := client.Search(context.Background(), &SearchRequest{
		Prompt: "test",
		Tools:  []string{"web"},
	})
	if err == nil {
		t.Fatal("expected error for non-200 response")
	}
}

func TestSearchStream_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/desearch/ai/search" {
			t.Errorf("expected path /desearch/ai/search, got %s", r.URL.Path)
		}
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			t.Errorf("expected Authorization header starting with 'Bearer ', got %s", auth)
		}

		var req SearchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
		if req.Streaming == nil || !*req.Streaming {
			t.Error("expected streaming to be true in SearchStream request")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"hacker_news_search":[{"title":"Test Result","link":"https://example.com","snippet":"Test"}]}`))
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL, APIKey: "test-key", HTTPClient: server.Client()}
	bufReader, err := client.SearchStream(context.Background(), &SearchRequest{
		Prompt: "test stream",
		Tools:  []string{"web"},
	})
	if err != nil {
		t.Fatalf("SearchStream failed: %v", err)
	}

	data, err := io.ReadAll(bufReader)
	if err != nil {
		t.Fatalf("failed to read stream: %v", err)
	}

	var resp SearchResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("failed to unmarshal stream data: %v", err)
	}
	if len(resp.HackerNewsSearch) != 1 {
		t.Errorf("expected 1 result, got %d", len(resp.HackerNewsSearch))
	}
}

func TestSearchStream_Non200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"detail": "Rate limit exceeded"}`))
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL, APIKey: "test-key", HTTPClient: server.Client()}
	_, err := client.SearchStream(context.Background(), &SearchRequest{
		Prompt: "test",
		Tools:  []string{"web"},
	})
	if err == nil {
		t.Fatal("expected error for non-200 response")
	}
}

func TestSearch_DecodeError(t *testing.T) {
	// Test that Search handles non-JSON response from server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Write invalid JSON - this will cause json.NewDecoder to fail
		w.Write([]byte(`{invalid json that cannot be decoded`))
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL, APIKey: "test-key", HTTPClient: server.Client()}
	_, err := client.Search(context.Background(), &SearchRequest{
		Prompt: "test",
		Tools:  []string{"web"},
	})
	if err == nil {
		t.Fatal("expected error for invalid JSON response")
	}
}

func ptrString(s string) *string {
	return &s
}
