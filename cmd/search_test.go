package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/roboalchemist/desearch-cli/pkg/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	// Initialize viper for config
	viper.SetConfigType("toml")
}

func resetFlags() {
	flagTool = nil
	flagDateFilter = ""
	flagStartDate = ""
	flagEndDate = ""
	flagStreaming = false
	flagResultType = ""
	flagCount = 0
	flagSystemMsg = ""
	flagNoAI = false
	flagPlaintext = false
	flagDryRun = false
	flagJQ = ""
	flagFields = ""
}

func TestBuildSearchRequest(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		tools          []string
		dateFilter     string
		startDate      string
		endDate        string
		streaming      bool
		resultType     string
		count          int
		systemMessage  string
		noAI           bool
		wantPrompt     string
		wantTools      []string
		wantDateFilter *string
		wantResultType *string
		wantCount      *int
	}{
		{
			name:           "basic query",
			query:          "test query",
			wantPrompt:     "test query",
			wantTools:      nil,
			wantDateFilter: nil,
			wantResultType: ptrString("LINKS_WITH_FINAL_SUMMARY"),
			wantCount:      nil,
		},
		{
			name:           "query with tools",
			query:          "test",
			tools:          []string{"web", "hackernews"},
			wantPrompt:     "test",
			wantTools:      []string{"web", "hackernews"},
			wantResultType: ptrString("LINKS_WITH_FINAL_SUMMARY"),
		},
		{
			name:           "query with date filter",
			query:          "test",
			dateFilter:     "PAST_WEEK",
			wantPrompt:     "test",
			wantDateFilter: ptrString("PAST_WEEK"),
			wantResultType: ptrString("LINKS_WITH_FINAL_SUMMARY"),
		},
		{
			name:           "query with start and end date",
			query:          "test",
			startDate:      "2024-01-01",
			endDate:        "2024-01-31",
			wantPrompt:     "test",
			wantDateFilter: nil,
			wantResultType: ptrString("LINKS_WITH_FINAL_SUMMARY"),
		},
		{
			name:           "query with result type",
			query:          "test",
			resultType:     "ONLY_LINKS",
			wantPrompt:     "test",
			wantResultType: ptrString("ONLY_LINKS"),
		},
		{
			name:           "query with count",
			query:          "test",
			count:          20,
			wantPrompt:     "test",
			wantCount:      ptrInt(20),
			wantResultType: ptrString("LINKS_WITH_FINAL_SUMMARY"),
		},
		{
			name:           "query with system message",
			query:          "test",
			systemMessage:  "Be concise",
			wantPrompt:     "test",
			wantResultType: ptrString("LINKS_WITH_FINAL_SUMMARY"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags before each sub-test
			resetFlags()

			// Set flags
			if tt.tools != nil {
				flagTool = tt.tools
			}
			if tt.dateFilter != "" {
				flagDateFilter = tt.dateFilter
			}
			if tt.startDate != "" {
				flagStartDate = tt.startDate
			}
			if tt.endDate != "" {
				flagEndDate = tt.endDate
			}
			if tt.streaming {
				flagStreaming = true
			}
			if tt.resultType != "" {
				flagResultType = tt.resultType
			}
			if tt.count != 0 {
				flagCount = tt.count
			}
			if tt.systemMessage != "" {
				flagSystemMsg = tt.systemMessage
			}
			if tt.noAI {
				flagNoAI = true
			}

			req := buildSearchRequest(tt.query)

			if req.Prompt != tt.wantPrompt {
				t.Errorf("Prompt = %q, want %q", req.Prompt, tt.wantPrompt)
			}
			if len(req.Tools) != len(tt.wantTools) {
				t.Errorf("Tools = %v, want %v", req.Tools, tt.wantTools)
			}
			if tt.wantDateFilter != nil && (req.DateFilter == nil || *req.DateFilter != *tt.wantDateFilter) {
				t.Errorf("DateFilter = %v, want %v", req.DateFilter, tt.wantDateFilter)
			}
			if tt.wantResultType != nil && (req.ResultType == nil || *req.ResultType != *tt.wantResultType) {
				t.Errorf("ResultType = %v, want %v", req.ResultType, tt.wantResultType)
			}
			if tt.wantCount != nil && (req.Count == nil || *req.Count != *tt.wantCount) {
				t.Errorf("Count = %v, want %v", req.Count, tt.wantCount)
			}
		})
	}
}

func TestGetAPIKey(t *testing.T) {
	// Save original and restore after test
	origAPIKey := apiKey
	t.Cleanup(func() {
		apiKey = origAPIKey
	})

	t.Run("returns flag API key when set", func(t *testing.T) {
		apiKey = "flag-key-123"
		if got := getAPIKey(); got != "flag-key-123" {
			t.Errorf("getAPIKey() = %q, want %q", got, "flag-key-123")
		}
	})
}

func TestRunSearch_NoAPIKey(t *testing.T) {
	// Save original and restore
	origAPIKey := apiKey
	t.Cleanup(func() {
		apiKey = origAPIKey
	})

	apiKey = ""
	resetFlags()

	cmd := &cobra.Command{}
	err := runSearch(cmd, []string{"test query"})

	// Should fail because there's no API key
	if err == nil {
		t.Error("expected error for empty API key")
	}
	if !strings.Contains(err.Error(), "no API key") {
		t.Errorf("error should mention 'no API key', got: %v", err)
	}
}

func TestRunSearch_FieldsWithoutJSON(t *testing.T) {
	// Save original and restore
	origAPIKey := apiKey
	t.Cleanup(func() {
		apiKey = origAPIKey
	})

	apiKey = ""
	resetFlags()
	flagFields = "completion"
	// jsonOut is false by default (zero value)

	cmd := &cobra.Command{}
	err := runSearch(cmd, []string{"test query"})

	if err == nil {
		t.Error("expected error for --fields without --json")
	}
	if !strings.Contains(err.Error(), "--fields requires --json") {
		t.Errorf("error should mention '--fields requires --json', got: %v", err)
	}
}

func TestRunSearch_FieldsWithDryRun(t *testing.T) {
	resetFlags()
	flagFields = "completion"
	flagDryRun = true

	cmd := &cobra.Command{}
	err := runSearch(cmd, []string{"test query"})

	if err == nil {
		t.Error("expected error for --fields with --dry-run")
	}
	if !strings.Contains(err.Error(), "--fields cannot be used with --dry-run") {
		t.Errorf("error should mention '--fields cannot be used with --dry-run', got: %v", err)
	}
}

func TestRunSearch_FieldsWithJSON_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(api.SearchResponse{
			Search: []api.WebResult{
				{Title: "Test Result", Link: "https://example.com", Snippet: "Test snippet"},
			},
			Completion: "AI summary",
		}); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	client := api.NewClient("test-key")
	client.BaseURL = server.URL
	client.HTTPClient = server.Client()

	resetFlags()
	flagFields = "completion"
	jsonOut = true

	cmd := &cobra.Command{}
	req := &api.SearchRequest{Prompt: "test query"}

	err := runSearchNormal(cmd, client, req)
	if err != nil {
		t.Fatalf("runSearchNormal with --fields failed: %v", err)
	}
}

func TestRunSearchNormal_WithMockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/desearch/ai/search" {
			t.Errorf("expected path /desearch/ai/search, got %s", r.URL.Path)
		}

		var req api.SearchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(api.SearchResponse{
			Search: []api.WebResult{
				{Title: "Test Result", Link: "https://example.com", Snippet: "Test snippet"},
			},
			Completion: "AI summary",
		}); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	// Create a custom client that uses the mock server
	client := api.NewClient("test-key")
	// Override BaseURL to point to mock server
	client.BaseURL = server.URL
	client.HTTPClient = server.Client()

	resetFlags()
	flagNoAI = false

	cmd := &cobra.Command{}
	req := &api.SearchRequest{Prompt: "test query"}

	err := runSearchNormal(cmd, client, req)
	if err != nil {
		t.Fatalf("runSearchNormal failed: %v", err)
	}
}

func TestRunSearchStream_WithMockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Write streaming response
		if _, err := w.Write([]byte(`{"completion": "Part 1"}` + "\n")); err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(`{"completion": "Part 2"}` + "\n")); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	client := api.NewClient("test-key")
	client.BaseURL = server.URL
	client.HTTPClient = server.Client()

	resetFlags()

	cmd := &cobra.Command{}
	req := &api.SearchRequest{Prompt: "test query"}

	err := runSearchStream(cmd, client, req)
	if err != nil {
		t.Fatalf("runSearchStream failed: %v", err)
	}
}

func TestRunSearch_DryRun(t *testing.T) {
	origAPIKey := apiKey
	t.Cleanup(func() {
		apiKey = origAPIKey
	})

	apiKey = ""
	resetFlags()
	flagDryRun = true

	cmd := &cobra.Command{}

	r, w, _ := os.Pipe()
	oldStdout := os.Stdout
	os.Stdout = w

	err := runSearch(cmd, []string{"test query"})

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("runSearch with dry-run failed: %v", err)
	}

	// Capture output
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatal(err)
	}
	output := buf.String()

	// Dry-run should output JSON
	if !strings.Contains(output, "\"prompt\"") {
		t.Errorf("dry-run output should contain JSON with prompt, got: %s", output)
	}
}

func TestSearchCmd_FlagParsing(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantErr    bool
		errMessage string
	}{
		{
			name:    "no args shows help",
			args:    []string{},
			wantErr: false, // cobra.ExactArgs(1) will show help but not error in test
		},
		{
			name:    "single arg is valid",
			args:    []string{"test query"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := searchCmd
			cmd.SetArgs(tt.args)

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.Execute()
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errMessage)
				} else if !strings.Contains(err.Error(), tt.errMessage) {
					t.Errorf("error = %q, want containing %q", err.Error(), tt.errMessage)
				}
			}
		})
	}
}

func TestSearchRequest_JSON(t *testing.T) {
	streaming := false
	count := 10
	resultType := "ONLY_LINKS"
	dateFilter := "PAST_WEEK"
	startDate := "2024-01-01"
	endDate := "2024-01-31"
	systemMsg := "Test system message"

	req := &api.SearchRequest{
		Prompt:        "test query",
		Tools:         []string{"web", "hackernews"},
		DateFilter:    &dateFilter,
		StartDate:     &startDate,
		EndDate:       &endDate,
		Streaming:     &streaming,
		ResultType:    &resultType,
		SystemMessage: &systemMsg,
		Count:         &count,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var decoded api.SearchRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if decoded.Prompt != req.Prompt {
		t.Errorf("Prompt = %q, want %q", decoded.Prompt, req.Prompt)
	}
	if len(decoded.Tools) != len(req.Tools) {
		t.Errorf("Tools len = %d, want %d", len(decoded.Tools), len(req.Tools))
	}
	if *decoded.Count != *req.Count {
		t.Errorf("Count = %d, want %d", *decoded.Count, *req.Count)
	}
}

// Helper functions
func ptrString(s string) *string {
	return &s
}

func ptrInt(i int) *int {
	return &i
}

// TestSearchCmdHelp tests that help command works
func TestSearchCmdHelp(t *testing.T) {
	cmd := searchCmd
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("help command failed: %v", err)
	}
	// Help command should succeed - just verify no error
}

func TestAPIClient_Search_DecodeError(t *testing.T) {
	// Test that Search handles non-JSON response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Write invalid JSON
		if _, err := w.Write([]byte(`{invalid json`)); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	client := &api.Client{BaseURL: server.URL, APIKey: "test-key", HTTPClient: server.Client()}
	_, err := client.Search(context.Background(), &api.SearchRequest{Prompt: "test"})
	if err == nil {
		t.Error("expected error for invalid JSON response")
	}
}

func TestAPIClient_SearchStream_Non200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		if _, err := w.Write([]byte(`{"detail": "Bad request"}`)); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	client := &api.Client{BaseURL: server.URL, APIKey: "test-key", HTTPClient: server.Client()}
	_, err := client.SearchStream(context.Background(), &api.SearchRequest{Prompt: "test"})
	if err == nil {
		t.Error("expected error for non-200 response")
	}
}

func TestRunSearchStream_ClientError(t *testing.T) {
	// Test that runSearchStream handles client errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"completion": "test"}`)); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	client := api.NewClient("test-key")
	client.BaseURL = server.URL
	client.HTTPClient = server.Client()

	resetFlags()
	cmd := &cobra.Command{}
	req := &api.SearchRequest{Prompt: "test"}

	// This should succeed
	err := runSearchStream(cmd, client, req)
	if err != nil {
		t.Errorf("runSearchStream unexpected error: %v", err)
	}
}

func TestSearchCmd_WithAllFlags(t *testing.T) {
	resetFlags()
	flagTool = []string{"web", "hackernews"}
	flagDateFilter = "PAST_WEEK"
	flagCount = 20
	flagResultType = "ONLY_LINKS"
	flagSystemMsg = "Test system"
	flagNoAI = true
	flagPlaintext = true

	req := buildSearchRequest("test query")

	if req.Prompt != "test query" {
		t.Errorf("Prompt = %q, want %q", req.Prompt, "test query")
	}
	if len(req.Tools) != 2 {
		t.Errorf("Tools len = %d, want 2", len(req.Tools))
	}
	if req.DateFilter == nil || *req.DateFilter != "PAST_WEEK" {
		t.Errorf("DateFilter = %v, want PAST_WEEK", req.DateFilter)
	}
	if req.Count == nil || *req.Count != 20 {
		t.Errorf("Count = %v, want 20", req.Count)
	}
}

func TestRunSearchNormal_Plaintext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(api.SearchResponse{
			Search: []api.WebResult{
				{Title: "Test", Link: "https://example.com", Snippet: "Snippet"},
			},
		}); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	client := api.NewClient("test-key")
	client.BaseURL = server.URL
	client.HTTPClient = server.Client()

	resetFlags()
	flagPlaintext = true

	cmd := &cobra.Command{}
	req := &api.SearchRequest{Prompt: "test"}

	err := runSearchNormal(cmd, client, req)
	if err != nil {
		t.Fatalf("runSearchNormal failed: %v", err)
	}
}

func TestRunSearchNormal_NoAI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(api.SearchResponse{
			Search: []api.WebResult{
				{Title: "Test", Link: "https://example.com", Snippet: "Snippet"},
			},
			Completion: "AI Summary",
		}); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	client := api.NewClient("test-key")
	client.BaseURL = server.URL
	client.HTTPClient = server.Client()

	resetFlags()
	flagNoAI = true

	cmd := &cobra.Command{}
	req := &api.SearchRequest{Prompt: "test"}

	err := runSearchNormal(cmd, client, req)
	if err != nil {
		t.Fatalf("runSearchNormal failed: %v", err)
	}
}

func TestRunSearchNormal_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(api.SearchResponse{}); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	client := api.NewClient("test-key")
	client.BaseURL = server.URL
	client.HTTPClient = server.Client()

	resetFlags()

	cmd := &cobra.Command{}
	req := &api.SearchRequest{Prompt: "test"}

	err := runSearchNormal(cmd, client, req)
	if err != nil {
		t.Fatalf("runSearchNormal failed: %v", err)
	}
}

func TestRunSearch_ClientSearchError(t *testing.T) {
	// Test error handling when client.Search fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := api.NewClient("test-key")
	client.BaseURL = server.URL
	client.HTTPClient = server.Client()

	resetFlags()

	cmd := &cobra.Command{}
	req := &api.SearchRequest{Prompt: "test"}

	err := runSearchNormal(cmd, client, req)
	if err == nil {
		t.Error("expected error when API returns 500")
	}
}
