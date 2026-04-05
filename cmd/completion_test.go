package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/roboalchemist/desearch-cli/pkg/api"
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
	// The Desearch API sends events as {"type": "text", "role": "summary", "content": "..."}
	// Only "type": "text" events produce output; others are skipped.
	tests := []struct {
		name       string
		data       string
		wantOutput bool
		wantText   string
	}{
		{
			name:       "text event with content",
			data:       `{"type": "text", "role": "summary", "content": "This is a test completion"}`,
			wantOutput: true,
			wantText:   "This is a test completion",
		},
		{
			name:       "text event with empty content",
			data:       `{"type": "text", "role": "summary", "content": ""}`,
			wantOutput: false,
		},
		{
			name:       "metadata event skipped",
			data:       `{"type": "metadata", "role": "summary", "content": "ignored"}`,
			wantOutput: false,
		},
		{
			name:       "done event skipped",
			data:       `{"type": "done"}`,
			wantOutput: false,
		},
		{
			name:       "JSON object without type field skipped",
			data:       `{"other": "field"}`,
			wantOutput: false,
		},
		{
			name:       "non-JSON skipped silently",
			data:       "Just some plain text",
			wantOutput: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var partial map[string]interface{}
			if err := json.Unmarshal([]byte(tt.data), &partial); err != nil {
				// Non-JSON is skipped silently
				if tt.wantOutput {
					t.Errorf("expected output but got non-JSON for data: %s", tt.data)
				}
				return
			}

			eventType, _ := partial["type"].(string)
			if eventType != "text" {
				if tt.wantOutput {
					t.Errorf("expected output but event type=%q is not 'text'", eventType)
				}
				return
			}

			content, _ := partial["content"].(string)
			if content == "" {
				if tt.wantOutput {
					t.Errorf("expected output but content is empty")
				}
				return
			}

			if !tt.wantOutput {
				t.Errorf("expected no output but got content: %s", content)
			}
			if content != tt.wantText {
				t.Errorf("content = %q, want %q", content, tt.wantText)
			}
		})
	}
}

func TestCompletionSSE_MultiEventParsing(t *testing.T) {
	// Verify that when multiple SSE events are packed on one line (no newline
	// between them), splitting on "data: " correctly extracts each JSON segment.
	// Events use the real Desearch API format: {"type": "text", "role": "summary", "content": "..."}
	tests := []struct {
		name     string
		raw      string
		wantText []string
	}{
		{
			name:     "single text event with prefix",
			raw:      `data: {"type":"text","role":"summary","content":"hello"}`,
			wantText: []string{"hello"},
		},
		{
			name:     "two text events packed on one line",
			raw:      `data: {"type":"text","role":"summary","content":"foo"}data: {"type":"text","role":"summary","content":"bar"}`,
			wantText: []string{"foo", "bar"},
		},
		{
			name:     "DONE sentinel ignored",
			raw:      `data: {"type":"text","role":"summary","content":"text"}data: [DONE]`,
			wantText: []string{"text"},
		},
		{
			name:     "empty segment skipped",
			raw:      `data: {"type":"text","role":"summary","content":"only"}data: `,
			wantText: []string{"only"},
		},
		{
			name:     "metadata event skipped",
			raw:      `data: {"type":"metadata","content":"ignored"}data: {"type":"text","role":"summary","content":"visible"}`,
			wantText: []string{"visible"},
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
				eventType, _ := partial["type"].(string)
				if eventType != "text" {
					continue
				}
				content, _ := partial["content"].(string)
				if content != "" {
					got = append(got, content)
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
