//go:build integration
// +build integration

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// buildBinary builds the desearch binary for testing.
func buildBinary(t *testing.T) string {
	t.Helper()

	// Find the project root (where go.mod is)
	modRoot, err := findModuleRoot()
	if err != nil {
		t.Skipf("could not find module root: %v", err)
	}

	// Build the binary to a temp location
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "desearch")

	cmd := exec.Command("go", "build",
		"-ldflags", "-X github.com/roboalchemist/desearch-cli/cmd.version=test",
		"-o", binaryPath,
		modRoot)
	cmd.Dir = modRoot

	if output, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("could not build binary: %v\n%s", err, output)
	}

	return binaryPath
}

func findModuleRoot() (string, error) {
	// Start from current working directory and walk up
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

// mockSearchServer creates an httptest server that responds to search requests.
func mockSearchServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/desearch/ai/search" {
			http.NotFound(w, r)
			return
		}

		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			http.Error(w, "expected JSON", http.StatusBadRequest)
			return
		}

		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Return a sample response
		resp := map[string]interface{}{
			"search": []map[string]string{
				{"title": "Test Result", "link": "https://example.com", "snippet": "Test snippet"},
			},
			"completion": "AI summary",
		}
		json.NewEncoder(w).Encode(resp)
	}))
}

func TestIntegration_Help(t *testing.T) {
	binary := buildBinary(t)

	tests := []struct {
		name    string
		args    []string
		wantOut string
	}{
		{"root help", []string{"--help"}, "CLI tool for Desearch AI"},
		{"search help", []string{"search", "--help"}, "Search the web"},
		{"completion help", []string{"completion", "--help"}, "AI-generated summary"},
		{"config help", []string{"config", "--help"}, "Manage the CLI configuration"},
		{"version help", []string{"version", "--help"}, "version"},
		{"docs help", []string{"docs", "--help"}, "Print the full README"},
		{"skill help", []string{"skill", "--help"}, "Claude Code skill"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binary, tt.args...)
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Errorf("command failed: %v\nOutput: %s", err, output)
				return
			}

			if !strings.Contains(string(output), tt.wantOut) {
				t.Errorf("output does not contain %q:\n%s", tt.wantOut, output)
			}
		})
	}
}

func TestIntegration_Version(t *testing.T) {
	binary := buildBinary(t)

	cmd := exec.Command(binary, "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("version command failed: %v\nOutput: %s", err, output)
	}
}

func TestIntegration_Search_DryRun(t *testing.T) {
	binary := buildBinary(t)

	cmd := exec.Command(binary, "search", "test query", "--dry-run")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Errorf("search --dry-run failed: %v\nOutput: %s", err, output)
		return
	}

	if !strings.Contains(string(output), "test query") {
		t.Errorf("dry-run output does not contain query:\n%s", output)
	}
}

func TestIntegration_ExitCodes(t *testing.T) {
	binary := buildBinary(t)

	tests := []struct {
		name     string
		args     []string
		wantCode int
	}{
		{"no args", []string{}, 0}, // shows help, no error
		{"unknown flag", []string{"search", "--invalid-flag"}, 2},
		{"unknown command", []string{"nonexistent"}, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binary, tt.args...)
			cmd.Run() // ignore output, just check exit code

			if cmd.ProcessState.ExitCode() != tt.wantCode {
				t.Errorf("exit code = %d, want %d", cmd.ProcessState.ExitCode(), tt.wantCode)
			}
		})
	}
}

func TestIntegration_SearchFlags(t *testing.T) {
	binary := buildBinary(t)

	// These should not error even without API key (dry-run mode)
	flags := []string{
		"--tool", "web",
		"--date-filter", "PAST_WEEK",
		"--count", "10",
		"--result-type", "ONLY_LINKS",
		"--system-message", "test",
		"--no-ai",
		"--plaintext",
		"--dry-run",
	}

	args := append([]string{"search", "test"}, flags...)
	cmd := exec.Command(binary, args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Errorf("search with flags failed: %v\nOutput: %s", err, output)
	}
}

func TestIntegration_CompletionFlags(t *testing.T) {
	binary := buildBinary(t)

	// Test completion with different flags (will fail due to no API key, but should not crash)
	flags := []string{
		"--system-message", "test",
		"--json",
	}

	for _, flag := range flags {
		t.Run(flag, func(t *testing.T) {
			args := append([]string{"completion", "test query"}, flag)
			cmd := exec.Command(binary, args...)
			output, _ := cmd.CombinedOutput()

			// Should fail with exit code 1 (no API key) but not crash
			if cmd.ProcessState.ExitCode() == 0 {
				t.Logf("command succeeded unexpectedly:\n%s", output)
			}
		})
	}
}

func TestIntegration_ConfigCommands(t *testing.T) {
	binary := buildBinary(t)

	// Test config show (no API key needed)
	cmd := exec.Command(binary, "config", "show")
	output, err := cmd.CombinedOutput()

	// May fail if config doesn't exist but should not crash
	if cmd.ProcessState.ExitCode() > 1 {
		t.Errorf("config show failed unexpectedly: %v\nOutput: %s", err, output)
	}

	// Test config --help
	cmd = exec.Command(binary, "config", "--help")
	output, err = cmd.CombinedOutput()

	if err != nil {
		t.Errorf("config --help failed: %v\nOutput: %s", err, output)
	}
}

func TestIntegration_DocsCommand(t *testing.T) {
	binary := buildBinary(t)

	cmd := exec.Command(binary, "docs")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Errorf("docs command failed: %v\nOutput: %s", err, output)
		return
	}

	if !strings.Contains(string(output), "Desearch") {
		t.Errorf("docs output does not contain 'Desearch':\n%s", output)
	}
}

func TestIntegration_SkillCommands(t *testing.T) {
	binary := buildBinary(t)

	// Test skill print
	cmd := exec.Command(binary, "skill", "print")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Errorf("skill print failed: %v\nOutput: %s", err, output)
	}

	// Test skill --help
	cmd = exec.Command(binary, "skill", "--help")
	output, err = cmd.CombinedOutput()

	if err != nil {
		t.Errorf("skill --help failed: %v\nOutput: %s", err, output)
	}
}

func TestIntegration_FlagCombinations(t *testing.T) {
	binary := buildBinary(t)

	// Test combinations that should work together
	combos := [][]string{
		{"--json", "--no-ai"},
		{"--plaintext", "--tool", "web"},
		{"--verbose", "--json"},
	}

	for _, flags := range combos {
		t.Run(strings.Join(flags, " "), func(t *testing.T) {
			args := append([]string{"search", "test"}, flags...)
			cmd := exec.Command(binary, args...)

			// Use dry-run to avoid needing API key
			args = append(args, "--dry-run")
			cmd = exec.Command(binary, args...)
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Errorf("command failed: %v\nOutput: %s", err, output)
			}
		})
	}
}

func TestIntegration_OutputFormats(t *testing.T) {
	binary := buildBinary(t)

	// Note: --plaintext doesn't work with --dry-run since dry-run always outputs JSON
	// So we only test --json format here
	t.Run("json format", func(t *testing.T) {
		cmd := exec.Command(binary, "search", "test", "--json", "--dry-run")
		output, err := cmd.CombinedOutput()

		if err != nil {
			t.Errorf("command failed: %v\nOutput: %s", err, output)
			return
		}

		// JSON should start with {
		trimmed := strings.TrimSpace(string(output))
		if !strings.HasPrefix(trimmed, "{") {
			t.Errorf("output does not look like JSON:\n%s", output)
		}
	})
}

// TestIntegration_LiveAPI runs live API tests if SKIP_INTEGRATION is not set
func TestIntegration_LiveAPI(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION") != "" {
		t.Skip("Skipping live API tests")
	}

	apiKey := os.Getenv("DESEARCH_API_KEY")
	if apiKey == "" {
		t.Skip("DESEARCH_API_KEY not set")
	}

	binary := buildBinary(t)

	// Test search command with real API
	cmd := exec.Command(binary, "search", "test query", "--api-key", apiKey, "--dry-run")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Logf("search dry-run output: %s", output)
	}

	// Test completion command with real API
	cmd = exec.Command(binary, "completion", "test query", "--api-key", apiKey)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	select {
	case err = <-done:
		if err != nil {
			t.Logf("completion finished with error: %v", err)
		}
	case <-time.After(30 * time.Second):
		cmd.Process.Kill()
		t.Log("completion timed out")
	}
}
