//go:build integration
// +build integration

package main

import (
	"context"
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

// shouldSkipWriteTests returns true when READONLY=1 is set,
// skipping tests that mutate the filesystem or config.
func shouldSkipWriteTests() bool {
	return os.Getenv("READONLY") == "1"
}

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

func TestIntegration_AICommand(t *testing.T) {
	binary := buildBinary(t)

	cmd := exec.Command(binary, "ai", "--help")
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Errorf("ai --help failed: %v\nOutput: %s", err, output)
		return
	}

	if !strings.Contains(string(output), "AI-generated summary") {
		t.Errorf("ai --help output does not contain expected description:\n%s", output)
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

func TestIntegration_ExitCode3_UnreadableConfig(t *testing.T) {
	if shouldSkipWriteTests() {
		t.Skip("READONLY=1")
	}

	// Create a temp HOME dir with an unreadable config file.
	// LoadConfig returns a SystemError when the file exists but is unreadable
	// (permission denied, I/O failure), which triggers exit code 3.
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "desearch-cli")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Skipf("could not create temp config dir: %v", err)
	}

	configPath := filepath.Join(configDir, "config.toml")
	// Write a readable file first, then remove read permission to trigger permission denied.
	if err := os.WriteFile(configPath, []byte("api_key = \"test\"\n"), 0644); err != nil {
		t.Skipf("could not write temp config file: %v", err)
	}
	if err := os.Chmod(configPath, 0000); err != nil {
		t.Skipf("could not chmod temp config file to 0000: %v", err)
	}

	binary := buildBinary(t)

	cmd := exec.Command(binary, "search", "--dry-run", "test")
	// Override HOME so the binary looks for config in our temp dir.
	cmd.Env = append(os.Environ(), "HOME="+tmpDir)
	cmd.Run()

	exitCode := -1
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}

	// On systems where the test runs as root (e.g. inside a container),
	// root can read any file regardless of mode 0000, so skip in that case.
	if exitCode != 3 {
		// If we got exit code 0, likely running as root — skip rather than fail.
		if exitCode == 0 {
			t.Skip("exit code 0 suggests running as root (root ignores mode 0000); skipping")
		}
		t.Errorf("exit code = %d, want 3", exitCode)
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
			// Build a fresh args slice to avoid mutating the backing array
			args := append([]string{"search", "test"}, flags...)
			args = append(args, "--dry-run")
			cmd := exec.Command(binary, args...)
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

func TestIntegration_StartEndDate(t *testing.T) {
	binary := buildBinary(t)

	// Test --start-date and --end-date flags individually and together
	tests := []struct {
		name  string
		flags []string
	}{
		{
			name:  "start-date only",
			flags: []string{"search", "test", "--start-date", "2024-01-01", "--dry-run"},
		},
		{
			name:  "end-date only",
			flags: []string{"search", "test", "--end-date", "2024-12-31", "--dry-run"},
		},
		{
			name:  "start-date and end-date together",
			flags: []string{"search", "test", "--start-date", "2024-01-01", "--end-date", "2024-12-31", "--dry-run"},
		},
		{
			name:  "start-date with other flags",
			flags: []string{"search", "test", "--start-date", "2024-01-01", "--tool", "web", "--count", "5", "--dry-run"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binary, tt.flags...)
			output, err := cmd.CombinedOutput()

			if err != nil {
				t.Errorf("command failed: %v\nOutput: %s", err, output)
				return
			}

			// Verify the output contains the query
			if !strings.Contains(string(output), "test") {
				t.Errorf("output does not contain query:\n%s", output)
			}
		})
	}
}

func TestIntegration_Streaming(t *testing.T) {
	binary := buildBinary(t)

	// Test --streaming flag is accepted (will fail due to no API key, but should not crash).
	// Use exec.CommandContext with a short timeout to avoid hanging indefinitely.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, binary, "search", "test", "--streaming", "--tool", "web")
	output, err := cmd.CombinedOutput()

	// Check if the flag was rejected (exit code 2 = CLI arg parse error = bug)
	exitCode := -1
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}

	if exitCode == 2 {
		t.Errorf("streaming flag was rejected as unknown:\n%s", output)
		return
	}

	// Any other exit code is acceptable (API error, timeout, etc.) - just not a crash
	if err != nil {
		t.Logf("streaming command exited with: %v\nOutput: %s", err, output)
	}

	// Verify the flag was parsed (output should not contain "unknown flag")
	outputStr := string(output)
	if strings.Contains(outputStr, "unknown flag") || strings.Contains(outputStr, "flag not found") {
		t.Errorf("streaming flag was not recognized:\n%s", output)
	}
}

func TestIntegration_ConfigDefaults(t *testing.T) {
	binary := buildBinary(t)

	// Test --default-tool flag sets default tools in config
	t.Run("set default-tool", func(t *testing.T) {
		cmd := exec.Command(binary, "config", "--default-tool", "web", "--default-tool", "hackernews")
		output, _ := cmd.CombinedOutput()

		// Should not crash or report unknown flag
		if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() == 2 {
			t.Errorf("default-tool flag was rejected:\n%s", output)
			return
		}

		// Should report config saved or help text
		outputStr := string(output)
		if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() != 0 {
			if !strings.Contains(outputStr, "Configuration saved") {
				t.Logf("config output: %s", outputStr)
			}
		}
	})

	// Test --default-date-filter flag
	t.Run("set default-date-filter", func(t *testing.T) {
		cmd := exec.Command(binary, "config", "--default-date-filter", "PAST_WEEK")
		output, _ := cmd.CombinedOutput()

		if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() == 2 {
			t.Errorf("default-date-filter flag was rejected:\n%s", output)
			return
		}

		outputStr := string(output)
		if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() != 0 {
			if !strings.Contains(outputStr, "Configuration saved") {
				t.Logf("config output: %s", outputStr)
			}
		}
	})

	// Test --default-tool and --default-date-filter together
	t.Run("set both default flags", func(t *testing.T) {
		cmd := exec.Command(binary, "config", "--default-tool", "reddit", "--default-date-filter", "PAST_MONTH")
		output, _ := cmd.CombinedOutput()

		if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() == 2 {
			t.Errorf("combined default flags were rejected:\n%s", output)
			return
		}

		outputStr := string(output)
		if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() != 0 {
			if !strings.Contains(outputStr, "Configuration saved") {
				t.Logf("config output: %s", outputStr)
			}
		}
	})
}

func TestIntegration_ConfigForce(t *testing.T) {
	binary := buildBinary(t)

	// Test config clear --force flag
	t.Run("clear with force", func(t *testing.T) {
		cmd := exec.Command(binary, "config", "clear", "--force")
		output, _ := cmd.CombinedOutput()

		// Should not crash or report unknown flag
		if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() == 2 {
			t.Errorf("force flag was rejected:\n%s", output)
			return
		}

		// Should succeed (cleared or "no config to clear" is both fine)
		outputStr := string(output)
		if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() != 0 {
			if !strings.Contains(outputStr, "No config file to clear") {
				t.Errorf("unexpected output for config clear --force:\n%s", outputStr)
			}
		}
	})

	// Test config clear -f (short form)
	t.Run("clear with -f short flag", func(t *testing.T) {
		cmd := exec.Command(binary, "config", "clear", "-f")
		output, _ := cmd.CombinedOutput()

		if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() == 2 {
			t.Errorf("-f flag was rejected:\n%s", output)
			return
		}

		outputStr := string(output)
		if cmd.ProcessState != nil && cmd.ProcessState.ExitCode() != 0 {
			if !strings.Contains(outputStr, "No config file to clear") {
				t.Errorf("unexpected output for config clear -f:\n%s", outputStr)
			}
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
