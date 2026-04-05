package output

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/roboalchemist/desearch-cli/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// makeTestResponse returns a minimal SearchResponse for use in tests.
func makeTestResponse() *api.SearchResponse {
	return &api.SearchResponse{
		Text: "test result",
	}
}

// TestWriteHistory_HappyPath verifies that WriteHistory creates a file at the
// expected path with the correct JSON envelope structure.
func TestWriteHistory_HappyPath(t *testing.T) {
	tmpDir := t.TempDir()
	params := map[string]interface{}{
		"prompt": "what is the best programming language",
		"tools":  []string{"web"},
	}
	resp := makeTestResponse()

	err := WriteHistory(tmpDir, "search", params, resp, 1234, true)
	require.NoError(t, err)

	// Find the written file — it's nested under history/search/<Y>/<M>/<D>/
	var files []string
	err = filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".json") {
			files = append(files, path)
		}
		return nil
	})
	require.NoError(t, err)
	require.Len(t, files, 1, "expected exactly one history file")

	filePath := files[0]

	// Verify path structure: <configDir>/history/search/<YYYY>/<MM>/<DD>/<filename>.json
	relPath, err := filepath.Rel(tmpDir, filePath)
	require.NoError(t, err)
	parts := strings.Split(relPath, string(filepath.Separator))
	// parts[0] = "history", parts[1] = "search", parts[2] = year, parts[3] = month, parts[4] = day, parts[5] = filename
	require.GreaterOrEqual(t, len(parts), 6, "unexpected path depth: %s", relPath)
	assert.Equal(t, "history", parts[0])
	assert.Equal(t, "search", parts[1])

	now := time.Now().UTC()
	assert.Equal(t, now.Format("2006"), parts[2], "year component")
	assert.Equal(t, now.Format("01"), parts[3], "month component")
	assert.Equal(t, now.Format("02"), parts[4], "day component")

	// Verify file content.
	data, err := os.ReadFile(filePath)
	require.NoError(t, err)

	var envelope map[string]interface{}
	err = json.Unmarshal(data, &envelope)
	require.NoError(t, err, "history file must be valid JSON")

	meta, ok := envelope["meta"].(map[string]interface{})
	require.True(t, ok, "envelope must have a 'meta' object")

	assert.Equal(t, "search", meta["command"])
	assert.NotEmpty(t, meta["timestamp"])
	assert.NotEmpty(t, meta["hostname"])
	assert.EqualValues(t, float64(1234), meta["latency_ms"])

	// Params should be present but must NOT contain api_key.
	metaParams, ok := meta["params"].(map[string]interface{})
	require.True(t, ok, "'params' must be a JSON object")
	assert.Equal(t, "what is the best programming language", metaParams["prompt"])
	assert.NotContains(t, metaParams, "api_key")

	// Response field must be present.
	_, ok = envelope["response"]
	assert.True(t, ok, "envelope must have a 'response' key")

	// File permissions must be 0600 (skip on Windows).
	if runtime.GOOS != "windows" {
		info, err := os.Stat(filePath)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0600), info.Mode().Perm(), "file must be mode 0600")
	}
}

// TestWriteHistory_NoOp verifies that WriteHistory is a no-op when
// historyEnabled is false.
func TestWriteHistory_NoOp(t *testing.T) {
	tmpDir := t.TempDir()
	params := map[string]interface{}{"prompt": "hello"}
	resp := makeTestResponse()

	err := WriteHistory(tmpDir, "search", params, resp, 500, false)
	require.NoError(t, err)

	// No files should have been created.
	entries, err := os.ReadDir(tmpDir)
	require.NoError(t, err)
	assert.Empty(t, entries, "no files should be created when historyEnabled=false")
}

// TestWriteHistory_AICommand verifies that cmd="ai" produces the correct
// directory structure.
func TestWriteHistory_AICommand(t *testing.T) {
	tmpDir := t.TempDir()
	params := map[string]interface{}{"prompt": "explain recursion"}
	resp := makeTestResponse()

	err := WriteHistory(tmpDir, "ai", params, resp, 800, true)
	require.NoError(t, err)

	var files []string
	err = filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	require.NoError(t, err)
	require.Len(t, files, 1)

	// The path must contain "ai" as the command segment.
	assert.Contains(t, files[0], filepath.Join("history", "ai"))
}

// TestWriteHistory_APIKeyStripped verifies that api_key is removed from params.
func TestWriteHistory_APIKeyStripped(t *testing.T) {
	tmpDir := t.TempDir()
	params := map[string]interface{}{
		"prompt":  "hello",
		"api_key": "super-secret-key",
	}
	resp := makeTestResponse()

	err := WriteHistory(tmpDir, "search", params, resp, 100, true)
	require.NoError(t, err)

	var files []string
	_ = filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	require.Len(t, files, 1)

	data, err := os.ReadFile(files[0])
	require.NoError(t, err)

	// api_key must not appear anywhere in the file content.
	assert.NotContains(t, string(data), "api_key",
		"api_key must be stripped from history file")
	assert.NotContains(t, string(data), "super-secret-key",
		"api_key value must not appear in history file")
}

// TestMakeSlug verifies slug sanitization behaviour.
func TestMakeSlug(t *testing.T) {
	tests := []struct {
		name   string
		params map[string]interface{}
		want   string
	}{
		{
			name:   "normal query",
			params: map[string]interface{}{"prompt": "what is golang"},
			want:   "what-is-golang",
		},
		{
			name:   "special characters stripped",
			params: map[string]interface{}{"prompt": "hello! world? foo/bar"},
			want:   "hello-world-foo-bar",
		},
		{
			name:   "long query truncated at 40 chars",
			params: map[string]interface{}{"prompt": "abcdefghijklmnopqrstuvwxyz0123456789ABCDE_extra"},
			// alphanumeric only: "abcdefghijklmnopqrstuvwxyz0123456789ABCDE" = 41 chars; truncated to 40
			want: "abcdefghijklmnopqrstuvwxyz0123456789ABCD",
		},
		{
			name:   "empty prompt",
			params: map[string]interface{}{"prompt": ""},
			want:   "",
		},
		{
			name:   "no prompt key",
			params: map[string]interface{}{},
			want:   "",
		},
		{
			name:   "leading and trailing separators stripped",
			params: map[string]interface{}{"prompt": "!!! hello !!!"},
			want:   "hello",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := makeSlug(tc.params)
			assert.Equal(t, tc.want, got)
		})
	}
}

// TestWriteHistory_UnwritableDir verifies that WriteHistory returns an error
// when the target directory is not writable (skipped on Windows and when
// running as root, where permission checks behave differently).
func TestWriteHistory_UnwritableDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission model differs on Windows")
	}
	if os.Getuid() == 0 {
		t.Skip("root bypasses permission checks")
	}

	tmpDir := t.TempDir()
	// Make the config dir itself read-only so MkdirAll fails.
	err := os.Chmod(tmpDir, 0500)
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chmod(tmpDir, 0700) })

	params := map[string]interface{}{"prompt": "test"}
	resp := makeTestResponse()

	writeErr := WriteHistory(tmpDir, "search", params, resp, 100, true)
	assert.Error(t, writeErr, "expected an error when directory is not writable")
}
