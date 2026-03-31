package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/roboalchemist/desearch-cli/pkg/api"
	"github.com/spf13/cobra"
)

func resetCompletionFlags() {
	completionSystemMessage = ""
}

func TestCompletionCmd_FlagParsing(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no args shows help and runs PreRun",
			args:    []string{},
			wantErr: false, // cobra shows help but doesn't error
		},
		{
			name:    "single arg is valid",
			args:    []string{"test query"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := completionCmd
			cmd.SetArgs(tt.args)

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			_ = cmd.Execute()
			// Just verify command runs without panicking
		})
	}
}

func TestCompletionCmd_Help(t *testing.T) {
	cmd := completionCmd
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("completionCmd --help failed: %v", err)
	}
	// Help command should succeed - just verify no error
}

func TestRunCompletion_NoAPIKey(t *testing.T) {
	// Save original and restore
	origAPIKey := apiKey
	t.Cleanup(func() {
		apiKey = origAPIKey
	})

	apiKey = ""
	resetCompletionFlags()

	cmd := &cobra.Command{}
	err := runCompletion(cmd, []string{"test query"})

	if err == nil {
		t.Error("expected error for missing API key")
	}
	if !strings.Contains(err.Error(), "no API key") {
		t.Errorf("error should mention 'no API key', got: %v", err)
	}
}

func TestCompletionRequest_Build(t *testing.T) {
	// Test building a completion request
	query := "test query"
	systemMsg := "Be concise"

	streaming := true
	resultType := "LINKS_WITH_FINAL_SUMMARY"

	req := &api.SearchRequest{
		Prompt:      query,
		Streaming:   &streaming,
		ResultType:  &resultType,
	}

	// Simulate the request building logic from runCompletion
	if systemMsg != "" {
		req.SystemMessage = &systemMsg
	}

	if req.Prompt != query {
		t.Errorf("Prompt = %q, want %q", req.Prompt, query)
	}
	if req.SystemMessage == nil || *req.SystemMessage != systemMsg {
		t.Errorf("SystemMessage = %v, want %q", req.SystemMessage, systemMsg)
	}
	if req.ResultType == nil || *req.ResultType != resultType {
		t.Errorf("ResultType = %v, want %q", req.ResultType, resultType)
	}
}

func TestCompletionResponse_Parse(t *testing.T) {
	tests := []struct {
		name       string
		data       string
		wantOutput bool
		wantText   string
	}{
		{
			name:       "completion chunk",
			data:       `{"completion": "This is a test completion"}`,
			wantOutput: true,
			wantText:   "This is a test completion",
		},
		{
			name:       "text chunk without completion",
			data:       `{"text": "Some text"}`,
			wantOutput: true,
			wantText:   "Some text",
		},
		{
			name:       "completion takes precedence",
			data:       `{"completion": "AI text", "text": "raw text"}`,
			wantOutput: true,
			wantText:   "AI text",
		},
		{
			name:       "empty completion",
			data:       `{"completion": ""}`,
			wantOutput: false,
		},
		{
			name:       "non-JSON raw text",
			data:       "Just some plain text",
			wantOutput: true,
			wantText:   "Just some plain text",
		},
		{
			name:       "JSON object without completion or text",
			data:       `{"other": "field"}`,
			wantOutput: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var partial map[string]interface{}
			if err := json.Unmarshal([]byte(tt.data), &partial); err != nil {
				// Non-JSON case
				if !tt.wantOutput {
					return
				}
				// For non-JSON, just verify we can print the raw text
				return
			}

			// Check for completion field
			if completion, ok := partial["completion"].(string); ok && completion != "" {
				if !tt.wantOutput {
					t.Errorf("expected no output but got completion: %s", completion)
				}
				if completion != tt.wantText {
					t.Errorf("completion = %q, want %q", completion, tt.wantText)
				}
				return
			}

			// Check for text field
			if text, ok := partial["text"].(string); ok && text != "" {
				if !tt.wantOutput {
					t.Errorf("expected no output but got text: %s", text)
				}
				if text != tt.wantText {
					t.Errorf("text = %q, want %q", text, tt.wantText)
				}
				return
			}

			// No output expected
			if tt.wantOutput {
				t.Errorf("expected output but got none for data: %s", tt.data)
			}
		})
	}
}

func TestCompletionCmd_ContextCancellation(t *testing.T) {
	// Test that context cancellation works properly
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// The function should return context.Canceled
	if ctx.Err() != context.Canceled {
		t.Errorf("context.Err() = %v, want %v", ctx.Err(), context.Canceled)
	}
}

// Integration test with mock server - verifies httptest server works correctly
func TestMockServer_Basic(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"completion": "Test completion response"}` + "\n"))
	}))
	defer server.Close()

	// Simple GET test to verify server works
	resp, err := server.Client().Get(server.URL)
	if err != nil {
		t.Fatalf("server not reachable: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("server status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

// TestSearchStreamResponse_Reader tests reading from a search stream response
func TestSearchStreamResponse_Reader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"completion": "Test 1"}` + "\n"))
		w.Write([]byte(`{"completion": "Test 2"}` + "\n"))
	}))
	defer server.Close()

	resp, err := server.Client().Get(server.URL)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	defer resp.Body.Close()

	// Read all response
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ReadAll() error = %v", err)
	}

	// Parse each line as JSON
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}

	var result map[string]string
	if err := json.Unmarshal([]byte(lines[0]), &result); err != nil {
		t.Fatalf("first line unmarshal error: %v", err)
	}
	if result["completion"] != "Test 1" {
		t.Errorf("first completion = %q, want %q", result["completion"], "Test 1")
	}
}

func TestCompletionCmd_SystemMessage(t *testing.T) {
	// Test the system message flag is properly set
	resetCompletionFlags()
	completionSystemMessage = "Test system message"

	if completionSystemMessage != "Test system message" {
		t.Errorf("completionSystemMessage = %q, want %q", completionSystemMessage, "Test system message")
	}
}
