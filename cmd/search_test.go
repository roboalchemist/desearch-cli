package cmd

import (
	"bytes"
	"encoding/json"
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

func TestRunSearchNormal_Integration(t *testing.T) {
	// This test requires a mock server to be set up
	// Skip if we can't run integration tests
	if os.Getenv("SKIP_INTEGRATION") != "" {
		t.Skip("Skipping integration test")
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
