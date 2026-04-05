package output

import (
	"encoding/json"
)

// ParseSSEEvent parses a single SSE data segment (after stripping the "data: " prefix)
// and returns the content string if the event is a "type":"text" event, or "" otherwise.
//
// Empty string is returned for: empty segments, [DONE] sentinel, non-JSON garbage,
// and non-"text" event types.
func ParseSSEEvent(segment []byte) string {
	segment = bytesTrimSpace(segment)
	if len(segment) == 0 {
		return ""
	}
	// Skip the SSE stream-end sentinel
	if string(segment) == "[DONE]" {
		return ""
	}
	var partial map[string]interface{}
	if err := json.Unmarshal(segment, &partial); err != nil {
		// Not valid JSON — skip silently (could be partial/garbled data)
		return ""
	}
	// Only emit content for "type": "text" events; skip metadata and done signals silently.
	eventType, _ := partial["type"].(string)
	if eventType != "text" {
		return ""
	}
	content, _ := partial["content"].(string)
	return content
}

// bytesTrimSpace is a copy of bytes.TrimSpace to avoid importing bytes in this package.
// It trims leading and trailing white space as defined by Unicode.
func bytesTrimSpace(b []byte) []byte {
	start := 0
	end := len(b)
	for start < end && isSpace(b[start]) {
		start++
	}
	for end > start && isSpace(b[end-1]) {
		end--
	}
	return b[start:end]
}

// isSpace reports whether b is a whitespace character.
func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}
