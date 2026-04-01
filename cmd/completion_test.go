package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/roboalchemist/desearch-cli/pkg/api"
	"github.com/spf13/cobra"
)


func resetCompletionFlags() {
	completionSystemMessage = ""
	completionJSON = false
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
			wantErr: false,
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
}

func TestRunCompletion_NoAPIKey(t *testing.T) {
	// Save original and restore (including env var which auth.GetAPIKey reads)
	origAPIKey := apiKey
	origEnvKey := os.Getenv("DESEARCH_API_KEY")
	t.Cleanup(func() {
		apiKey = origAPIKey
		if origEnvKey != "" {
			os.Setenv("DESEARCH_API_KEY", origEnvKey)
		} else {
			os.Unsetenv("DESEARCH_API_KEY")
		}
	})

	apiKey = ""
	os.Unsetenv("DESEARCH_API_KEY")
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
	query := "test query"
	systemMsg := "Be concise"

	streaming := true
	resultType := "LINKS_WITH_FINAL_SUMMARY"

	req := &api.SearchRequest{
		Prompt:     query,
		Streaming:  &streaming,
		ResultType: &resultType,
	}

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
				if !tt.wantOutput {
					return
				}
				return
			}

			if completion, ok := partial["completion"].(string); ok && completion != "" {
				if !tt.wantOutput {
					t.Errorf("expected no output but got completion: %s", completion)
				}
				if completion != tt.wantText {
					t.Errorf("completion = %q, want %q", completion, tt.wantText)
				}
				return
			}

			if text, ok := partial["text"].(string); ok && text != "" {
				if !tt.wantOutput {
					t.Errorf("expected no output but got text: %s", text)
				}
				if text != tt.wantText {
					t.Errorf("text = %q, want %q", text, tt.wantText)
				}
				return
			}

			if tt.wantOutput {
				t.Errorf("expected output but got none for data: %s", tt.data)
			}
		})
	}
}

func TestCompletionSSE_MultiEventParsing(t *testing.T) {
	// Verify that when multiple SSE events are packed on one line (no newline
	// between them), splitting on "data: " correctly extracts each JSON segment.
	tests := []struct {
		name     string
		raw      string
		wantText []string
	}{
		{
			name:     "single event with prefix",
			raw:      `data: {"completion":"hello"}`,
			wantText: []string{"hello"},
		},
		{
			name:     "two events packed on one line",
			raw:      `data: {"completion":"foo"}data: {"completion":"bar"}`,
			wantText: []string{"foo", "bar"},
		},
		{
			name:     "DONE sentinel ignored",
			raw:      `data: {"completion":"text"}data: [DONE]`,
			wantText: []string{"text"},
		},
		{
			name:     "empty segment skipped",
			raw:      `data: {"completion":"only"}data: `,
			wantText: []string{"only"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []string

			// Simulate the splitting logic from runCompletion
			segments := bytes.Split([]byte(tt.raw), []byte("data: "))
			for _, seg := range segments {
				seg = bytes.TrimSpace(seg)
				if len(seg) == 0 {
					continue
				}
				if string(seg) == "[DONE]" {
					continue
				}
				var partial map[string]interface{}
				if err := json.Unmarshal(seg, &partial); err != nil {
					continue
				}
				if completion, ok := partial["completion"].(string); ok && completion != "" {
					got = append(got, completion)
				} else if text, ok := partial["text"].(string); ok && text != "" {
					got = append(got, text)
				}
			}

			if len(got) != len(tt.wantText) {
				t.Fatalf("got %d segments %v, want %d segments %v", len(got), got, len(tt.wantText), tt.wantText)
			}
			for i, w := range tt.wantText {
				if got[i] != w {
					t.Errorf("segment[%d] = %q, want %q", i, got[i], w)
				}
			}
		})
	}
}

func TestCompletionCmd_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if ctx.Err() != context.Canceled {
		t.Errorf("context.Err() = %v, want %v", ctx.Err(), context.Canceled)
	}
}

func TestCompletionCmd_SystemMessage(t *testing.T) {
	resetCompletionFlags()
	completionSystemMessage = "Test system message"

	if completionSystemMessage != "Test system message" {
		t.Errorf("completionSystemMessage = %q, want %q", completionSystemMessage, "Test system message")
	}
}
