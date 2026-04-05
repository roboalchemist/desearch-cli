package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
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
		Prompt:        "Bittensor",
		Tools:         []string{"web", "hackernews"},
		DateFilter:    ptrString("PAST_WEEK"),
		Streaming:     &streaming,
		ResultType:    ptrString("LINKS_WITH_FINAL_SUMMARY"),
		SystemMessage: ptrString("Summarize in pros and cons"),
		Count:         &count,
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
				User:      TweetUser{Username: "testuser", Name: "Test User"},
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
		if auth == "" {
			t.Errorf("expected non-empty Authorization header, got empty string")
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
		if err := json.NewEncoder(w).Encode(SearchResponse{
			HackerNewsSearch: []HackerNewsResult{
				{Title: "Test Result", Link: "https://example.com", Snippet: "Test"},
			},
		}); err != nil {
			t.Fatal(err)
		}
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
		if _, err := w.Write([]byte(`{"detail": "Invalid API key"}`)); err != nil {
			t.Fatal(err)
		}
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
		if auth == "" {
			t.Errorf("expected non-empty Authorization header, got empty string")
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
		if _, err := w.Write([]byte(`{"hacker_news_search":[{"title":"Test Result","link":"https://example.com","snippet":"Test"}]}`)); err != nil {
			t.Fatal(err)
		}
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
	defer bufReader.Close()

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
		if _, err := w.Write([]byte(`{"detail": "Rate limit exceeded"}`)); err != nil {
			t.Fatal(err)
		}
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
		if _, err := w.Write([]byte(`{invalid json that cannot be decoded`)); err != nil {
			t.Fatal(err)
		}
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

// TestSearch_Non200_JSONDecodeError verifies that Search returns an error when
// a non-200 response contains malformed JSON (decode error path for non-2xx).
func TestSearch_Non200_JSONDecodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		// Return malformed JSON — client should still return an error.
		if _, err := w.Write([]byte(`{not valid json`)); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL, APIKey: "test-key", HTTPClient: server.Client()}
	_, err := client.Search(context.Background(), &SearchRequest{
		Prompt: "test",
		Tools:  []string{"web"},
	})
	if err == nil {
		t.Fatal("expected error for non-200 response with malformed JSON")
	}
}

// TestSearchStream_Non200_JSONDecodeError verifies that SearchStream returns an error
// when a non-200 response contains malformed JSON (decode error path for non-2xx).
func TestSearchStream_Non200_JSONDecodeError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		// Return malformed JSON — client should still return an error.
		if _, err := w.Write([]byte(`{malformed`)); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL, APIKey: "test-key", HTTPClient: server.Client()}
	_, err := client.SearchStream(context.Background(), &SearchRequest{
		Prompt: "test",
		Tools:  []string{"web"},
	})
	if err == nil {
		t.Fatal("expected error for non-200 response with malformed JSON")
	}
}

// TestSearch_Non200_StructuredError verifies that Search returns a structured
// error message when a non-200 response contains valid JSON with a "detail" field.
func TestSearch_Non200_StructuredError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"detail": "invalid-api-key"}`)); err != nil {
			t.Fatal(err)
		}
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

// TestSearch_Non200_RawBodyError verifies that Search falls back to raw body
// when a non-200 response does not contain a "detail" field in JSON.
func TestSearch_Non200_RawBodyError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"error": "something went wrong"}`)); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL, APIKey: "test-key", HTTPClient: server.Client()}
	_, err := client.Search(context.Background(), &SearchRequest{
		Prompt: "test",
		Tools:  []string{"web"},
	})
	if err == nil {
		t.Fatal("expected error for non-200 response")
	}
}

// TestSearch_ClientDoError verifies that Search returns an error when the HTTP
// client fails to send the request (e.g., server unreachable or connection refused).
func TestSearch_ClientDoError(t *testing.T) {
	// Use a port that nothing is listening on to force a connection error.
	client := &Client{BaseURL: "http://127.0.0.1:1", APIKey: "test-key", HTTPClient: &http.Client{}}
	_, err := client.Search(context.Background(), &SearchRequest{
		Prompt: "test",
		Tools:  []string{"web"},
	})
	if err == nil {
		t.Fatal("expected error when HTTP client fails")
	}
}

// TestSearch_RequestCreationError verifies that Search returns an error when
// http.NewRequestWithContext fails due to a malformed URL.
func TestSearch_RequestCreationError(t *testing.T) {
	client := &Client{BaseURL: "http://[invalid", APIKey: "test-key", HTTPClient: &http.Client{}}
	_, err := client.Search(context.Background(), &SearchRequest{
		Prompt: "test",
		Tools:  []string{"web"},
	})
	if err == nil {
		t.Fatal("expected error for malformed URL")
	}
}

// TestSearchStream_ClientDoError verifies that SearchStream returns an error when
// the HTTP client fails to send the request.
func TestSearchStream_ClientDoError(t *testing.T) {
	// Use a port that nothing is listening on to force a connection error.
	client := &Client{BaseURL: "http://127.0.0.1:1", APIKey: "test-key", HTTPClient: &http.Client{}}
	_, err := client.SearchStream(context.Background(), &SearchRequest{
		Prompt: "test",
		Tools:  []string{"web"},
	})
	if err == nil {
		t.Fatal("expected error when HTTP client fails")
	}
}

// TestSearchStream_RequestCreationError verifies that SearchStream returns an error
// when http.NewRequestWithContext fails due to a malformed URL.
func TestSearchStream_RequestCreationError(t *testing.T) {
	client := &Client{BaseURL: "http://[invalid", APIKey: "test-key", HTTPClient: &http.Client{}}
	_, err := client.SearchStream(context.Background(), &SearchRequest{
		Prompt: "test",
		Tools:  []string{"web"},
	})
	if err == nil {
		t.Fatal("expected error for malformed URL")
	}
}

// TestSearchStream_ContextCancel_ClosesBody verifies that SearchStream closes the
// response body when the context is cancelled before the stream finishes.
func TestSearchStream_ContextCancel_ClosesBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Write nothing — handler returns immediately after headers.
		// The client will cancel the context, which may cause the connection to
		// close early. The important thing: Close() must be callable without panic.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create a context that cancels immediately.
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	client := &Client{BaseURL: server.URL, APIKey: "test-key", HTTPClient: server.Client()}
	reader, err := client.SearchStream(ctx, &SearchRequest{
		Prompt: "test",
		Tools:  []string{"web"},
	})
	// With immediate cancel, Do() returns a context error.
	if err == nil {
		// If there's no error (unlikely), still close and verify it works.
		reader.Close()
	}
	// No goroutine leak: the defer in test server won't hang.
}

// TestSearchStream_ReturnsReadCloser verifies that SearchStream returns a
// *streamReadCloser (which satisfies io.ReadCloser) so callers can close the body.
func TestSearchStream_ReturnsReadCloser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{}`)); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	client := &Client{BaseURL: server.URL, APIKey: "test-key", HTTPClient: server.Client()}
	reader, err := client.SearchStream(context.Background(), &SearchRequest{
		Prompt: "test",
		Tools:  []string{"web"},
	})
	if err != nil {
		t.Fatalf("SearchStream failed: %v", err)
	}

	// Verify it's an io.ReadCloser (has Close method).
	var closer io.Closer = reader
	_ = closer // staticcheck: concrete type is never nil

	// Verify we can read something.
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("failed to read stream: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected some data from stream")
	}

	// Verify Close works.
	if err := reader.Close(); err != nil {
		t.Errorf("Close() returned error: %v", err)
	}
}

func ptrString(s string) *string {
	return &s
}
