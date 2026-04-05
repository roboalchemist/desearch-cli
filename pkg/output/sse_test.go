package output

import (
	"testing"
)

func TestParseSSEEvent(t *testing.T) {
	tests := []struct {
		name    string
		segment []byte
		want    string
	}{
		{
			name:    "empty segment",
			segment: []byte(""),
			want:    "",
		},
		{
			name:    "whitespace only",
			segment: []byte("   \t\n  "),
			want:    "",
		},
		{
			name:    "done sentinel",
			segment: []byte("[DONE]"),
			want:    "",
		},
		{
			name:    "done sentinel with whitespace",
			segment: []byte("  [DONE]  "),
			want:    "",
		},
		{
			name:    "non-JSON garbage",
			segment: []byte("not valid json at all"),
			want:    "",
		},
		{
			name:    "non-text event type",
			segment: []byte(`{"type":"metadata","content":"some metadata"}`),
			want:    "",
		},
		{
			name:    "text event with content",
			segment: []byte(`{"type":"text","content":"Hello world"}`),
			want:    "Hello world",
		},
		{
			name:    "text event with whitespace",
			segment: []byte(`  {"type":"text","content":"Hello"}  `),
			want:    "Hello",
		},
		{
			name:    "text event with extra fields",
			segment: []byte(`{"type":"text","role":"summary","content":"Summary text","index":0}`),
			want:    "Summary text",
		},
		{
			name:    "text event empty content",
			segment: []byte(`{"type":"text","content":""}`),
			want:    "",
		},
		{
			name:    "text event missing content field",
			segment: []byte(`{"type":"text"}`),
			want:    "",
		},
		{
			name:    "text event nil content",
			segment: []byte(`{"type":"text","content":null}`),
			want:    "",
		},
		{
			name:    "text event non-string content",
			segment: []byte(`{"type":"text","content":123}`),
			want:    "",
		},
		{
			name:    "text event content is integer zero",
			segment: []byte(`{"type":"text","content":0}`),
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseSSEEvent(tt.segment)
			if got != tt.want {
				t.Errorf("ParseSSEEvent(%q) = %q, want %q", string(tt.segment), got, tt.want)
			}
		})
	}
}
