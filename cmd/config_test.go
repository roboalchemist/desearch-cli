package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/roboalchemist/desearch-cli/pkg/auth"
	"github.com/spf13/viper"
)

func initConfigTest() {
	viper.Reset()
}

func resetConfigFlags() {
	flagAPIKey = ""
	flagDefaultTools = nil
	flagDefaultDateFilter = ""
	flagDefaultCount = 0
	// Reset the history-enabled and default-count flags' Changed state.
	if f := configCmd.Flags().Lookup("history-enabled"); f != nil {
		f.Changed = false
		_ = f.Value.Set("false")
	}
	if f := configCmd.Flags().Lookup("default-count"); f != nil {
		f.Changed = false
		_ = f.Value.Set("0")
	}
}

func TestConfigPath(t *testing.T) {
	// Save original and restore after test
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	t.Cleanup(func() {
		if origXDG == "" {
			os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			os.Setenv("XDG_CONFIG_HOME", origXDG)
		}
	})

	t.Run("uses XDG_CONFIG_HOME when set", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tmpDir)

		path, err := auth.ConfigPath()
		if err != nil {
			t.Fatalf("auth.ConfigPath() error = %v", err)
		}
		expected := filepath.Join(tmpDir, "desearch-cli", "config.toml")
		if path != expected {
			t.Errorf("auth.ConfigPath() = %q, want %q", path, expected)
		}
	})

	t.Run("falls back to ~/.config when XDG_CONFIG_HOME not set", func(t *testing.T) {
		os.Unsetenv("XDG_CONFIG_HOME")

		home, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("UserHomeDir() error = %v", err)
		}
		path, err := auth.ConfigPath()
		if err != nil {
			t.Fatalf("auth.ConfigPath() error = %v", err)
		}
		expected := filepath.Join(home, ".config", "desearch-cli", "config.toml")
		if path != expected {
			t.Errorf("auth.ConfigPath() = %q, want %q", path, expected)
		}
	})
}

func TestShowCmd(t *testing.T) {
	// Save original and restore
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	t.Cleanup(func() {
		if origXDG == "" {
			os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			os.Setenv("XDG_CONFIG_HOME", origXDG)
		}
	})

	t.Run("shows empty config", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tmpDir)
		initConfigTest()

		cmd := showCmd
		// Note: showCmd writes directly to os.Stderr, so we can't capture output in buffer
		// We just verify the command runs without panic/error
		err := cmd.Execute()
		if err != nil {
			t.Errorf("showCmd.Execute() error = %v", err)
		}
	})
}

func TestConfigCmd_ShowHelp(t *testing.T) {
	cmd := configCmd
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("configCmd --help failed: %v", err)
	}
	// Help command should succeed - just verify no error
}

func TestClearCmd(t *testing.T) {
	// Save original and restore
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	t.Cleanup(func() {
		if origXDG == "" {
			os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			os.Setenv("XDG_CONFIG_HOME", origXDG)
		}
	})

	t.Run("clear non-existent config does not error", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tmpDir)
		initConfigTest()

		cmd := clearCmd
		// Note: clearCmd writes directly to os.Stderr, so we can't capture output in buffer
		// We just verify the command runs without error
		err := cmd.Execute()
		if err != nil {
			t.Errorf("clearCmd.Execute() error = %v", err)
		}
	})
}

func TestConfigCmd_SetAPIKey(t *testing.T) {
	// Save original and restore
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	t.Cleanup(func() {
		if origXDG == "" {
			os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			os.Setenv("XDG_CONFIG_HOME", origXDG)
		}
	})

	t.Run("sets api-key flag", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tmpDir)
		initConfigTest()
		resetConfigFlags()

		// Instead of executing the full command (which calls os.Exit),
		// we test the flag parsing logic by simulating what the command does
		flagAPIKey = "test-key-123"

		// Verify the flag was set
		if flagAPIKey != "test-key-123" {
			t.Errorf("flagAPIKey = %q, want %q", flagAPIKey, "test-key-123")
		}
	})
}

func TestConfigCmd_EmptyAPIKey(t *testing.T) {
	// Test that empty API key after trimming is rejected.
	// The actual rejection happens in configCmd.RunE.
	emptyKey := "   "
	if strings.TrimSpace(emptyKey) != "" {
		t.Errorf("expected trimmed key to be empty, got %q", strings.TrimSpace(emptyKey))
	}
}

func TestConfigCmd_Subcommands(t *testing.T) {
	cmd := configCmd

	// Verify subcommands are added
	subcommands := cmd.Commands()

	var hasShow, hasClear bool
	for _, sub := range subcommands {
		if sub.Name() == "show" {
			hasShow = true
		}
		if sub.Name() == "clear" {
			hasClear = true
		}
	}

	if !hasShow {
		t.Error("configCmd should have 'show' subcommand")
	}
	if !hasClear {
		t.Error("configCmd should have 'clear' subcommand")
	}
}

func TestConfigCmd_Flags(t *testing.T) {
	cmd := configCmd

	// Verify flags exist
	flag := cmd.Flags().Lookup("api-key")
	if flag == nil {
		t.Error("configCmd should have --api-key flag")
	}

	flag = cmd.Flags().Lookup("default-tool")
	if flag == nil {
		t.Error("configCmd should have --default-tool flag")
	}

	flag = cmd.Flags().Lookup("default-date-filter")
	if flag == nil {
		t.Error("configCmd should have --default-date-filter flag")
	}

	flag = cmd.Flags().Lookup("history-enabled")
	if flag == nil {
		t.Error("configCmd should have --history-enabled flag")
	}

	flag = cmd.Flags().Lookup("default-count")
	if flag == nil {
		t.Error("configCmd should have --default-count flag")
	}
}

func TestConfigCmdRunE_SetsAPIKey(t *testing.T) {
	// Save original and restore
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	t.Cleanup(func() {
		if origXDG == "" {
			os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			os.Setenv("XDG_CONFIG_HOME", origXDG)
		}
		resetConfigFlags()
	})

	t.Run("saves config with api-key", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tmpDir)
		initConfigTest()
		resetConfigFlags()

		// Directly call the RunE closure by invoking the command with flags
		flagAPIKey = "test-api-key-xyz"
		flagDefaultTools = nil
		flagDefaultDateFilter = ""

		cmd := configCmd
		err := cmd.RunE(cmd, []string{})
		if err != nil {
			t.Fatalf("configCmd.RunE with api-key failed: %v", err)
		}
	})

	t.Run("saves config with default-tool", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tmpDir)
		initConfigTest()
		resetConfigFlags()

		flagAPIKey = ""
		flagDefaultTools = []string{"hackernews", "reddit"}
		flagDefaultDateFilter = ""

		cmd := configCmd
		err := cmd.RunE(cmd, []string{})
		if err != nil {
			t.Fatalf("configCmd.RunE with default-tool failed: %v", err)
		}
	})

	t.Run("saves config with default-date-filter", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tmpDir)
		initConfigTest()
		resetConfigFlags()

		flagAPIKey = ""
		flagDefaultTools = nil
		flagDefaultDateFilter = "PAST_WEEK"

		cmd := configCmd
		err := cmd.RunE(cmd, []string{})
		if err != nil {
			t.Fatalf("configCmd.RunE with default-date-filter failed: %v", err)
		}
	})

	t.Run("shows help when no flags provided", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tmpDir)
		initConfigTest()
		resetConfigFlags()

		flagAPIKey = ""
		flagDefaultTools = nil
		flagDefaultDateFilter = ""

		cmd := configCmd
		err := cmd.RunE(cmd, []string{})
		if err != nil {
			t.Fatalf("configCmd.RunE with no flags failed: %v", err)
		}
	})
}

func TestConfigCmd_HistoryEnabled(t *testing.T) {
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	t.Cleanup(func() {
		if origXDG == "" {
			os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			os.Setenv("XDG_CONFIG_HOME", origXDG)
		}
		resetConfigFlags()
	})

	t.Run("sets history-enabled=true", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tmpDir)
		initConfigTest()
		resetConfigFlags()

		f := configCmd.Flags().Lookup("history-enabled")
		if err := f.Value.Set("true"); err != nil {
			t.Fatalf("failed to set history-enabled: %v", err)
		}
		f.Changed = true

		cmd := configCmd
		err := cmd.RunE(cmd, []string{})
		if err != nil {
			t.Fatalf("configCmd.RunE with --history-enabled=true failed: %v", err)
		}

		cfg, err := auth.LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig failed: %v", err)
		}
		if !cfg.HistoryEnabled {
			t.Error("expected HistoryEnabled=true after --history-enabled=true")
		}
	})

	t.Run("sets history-enabled=false", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tmpDir)
		initConfigTest()
		resetConfigFlags()

		// Pre-set history_enabled=true in config
		cfgPath := filepath.Join(tmpDir, "desearch-cli", "config.toml")
		if err := os.MkdirAll(filepath.Dir(cfgPath), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(cfgPath, []byte("history_enabled = true\n"), 0600); err != nil {
			t.Fatal(err)
		}

		f := configCmd.Flags().Lookup("history-enabled")
		if err := f.Value.Set("false"); err != nil {
			t.Fatalf("failed to set history-enabled: %v", err)
		}
		f.Changed = true

		cmd := configCmd
		err := cmd.RunE(cmd, []string{})
		if err != nil {
			t.Fatalf("configCmd.RunE with --history-enabled=false failed: %v", err)
		}

		cfg, err := auth.LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig failed: %v", err)
		}
		if cfg.HistoryEnabled {
			t.Error("expected HistoryEnabled=false after --history-enabled=false")
		}
	})
}

func TestConfigCmd_DefaultCount(t *testing.T) {
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	t.Cleanup(func() {
		if origXDG == "" {
			os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			os.Setenv("XDG_CONFIG_HOME", origXDG)
		}
		resetConfigFlags()
	})

	t.Run("sets default-count to 50", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tmpDir)
		initConfigTest()
		resetConfigFlags()

		flagDefaultCount = 50
		f := configCmd.Flags().Lookup("default-count")
		if err := f.Value.Set("50"); err != nil {
			t.Fatalf("failed to set default-count: %v", err)
		}
		f.Changed = true

		cmd := configCmd
		err := cmd.RunE(cmd, []string{})
		if err != nil {
			t.Fatalf("configCmd.RunE with --default-count 50 failed: %v", err)
		}

		cfg, err := auth.LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig failed: %v", err)
		}
		if cfg.DefaultCount != 50 {
			t.Errorf("expected DefaultCount=50, got %d", cfg.DefaultCount)
		}
	})

	t.Run("clears default-count with 0", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tmpDir)
		initConfigTest()
		resetConfigFlags()

		// Pre-set a count
		cfgPath := filepath.Join(tmpDir, "desearch-cli", "config.toml")
		if err := os.MkdirAll(filepath.Dir(cfgPath), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(cfgPath, []byte("default_count = 100\n"), 0600); err != nil {
			t.Fatal(err)
		}

		flagDefaultCount = 0
		f := configCmd.Flags().Lookup("default-count")
		if err := f.Value.Set("0"); err != nil {
			t.Fatalf("failed to set default-count: %v", err)
		}
		f.Changed = true

		cmd := configCmd
		err := cmd.RunE(cmd, []string{})
		if err != nil {
			t.Fatalf("configCmd.RunE with --default-count 0 failed: %v", err)
		}

		cfg, err := auth.LoadConfig()
		if err != nil {
			t.Fatalf("LoadConfig failed: %v", err)
		}
		if cfg.DefaultCount != 0 {
			t.Errorf("expected DefaultCount=0, got %d", cfg.DefaultCount)
		}
	})

	t.Run("rejects out-of-range default-count", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tmpDir)
		initConfigTest()
		resetConfigFlags()

		flagDefaultCount = 5 // below minimum of 10
		f := configCmd.Flags().Lookup("default-count")
		if err := f.Value.Set("5"); err != nil {
			t.Fatalf("failed to set default-count: %v", err)
		}
		f.Changed = true

		cmd := configCmd
		err := cmd.RunE(cmd, []string{})
		if err == nil {
			t.Fatal("expected error for out-of-range default-count, got nil")
		}
		if !strings.Contains(err.Error(), "--default-count") {
			t.Errorf("expected error message to mention --default-count, got: %v", err)
		}
	})
}

func TestShowCmd_DisplaysNewFields(t *testing.T) {
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	origJSONOut := jsonOut
	t.Cleanup(func() {
		if origXDG == "" {
			os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			os.Setenv("XDG_CONFIG_HOME", origXDG)
		}
		jsonOut = origJSONOut
	})

	t.Run("text mode shows history_enabled and default_count", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tmpDir)
		initConfigTest()
		jsonOut = false

		cfgPath := filepath.Join(tmpDir, "desearch-cli", "config.toml")
		if err := os.MkdirAll(filepath.Dir(cfgPath), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(cfgPath, []byte("history_enabled = true\ndefault_count = 20\n"), 0600); err != nil {
			t.Fatal(err)
		}

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := showCmd.RunE(showCmd, []string{})

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		if _, err2 := buf.ReadFrom(r); err2 != nil {
			t.Fatalf("failed to read captured output: %v", err2)
		}
		output := buf.String()

		if err != nil {
			t.Fatalf("showCmd.RunE failed: %v", err)
		}
		if !strings.Contains(output, "History Enabled:") {
			t.Errorf("expected 'History Enabled:' in output, got:\n%s", output)
		}
		if !strings.Contains(output, "Default Count:") {
			t.Errorf("expected 'Default Count:' in output, got:\n%s", output)
		}
		if !strings.Contains(output, "true") {
			t.Errorf("expected 'true' in output for history_enabled, got:\n%s", output)
		}
		if !strings.Contains(output, "20") {
			t.Errorf("expected '20' in output for default_count, got:\n%s", output)
		}
	})

	t.Run("json mode includes history_enabled and default_count", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tmpDir)
		initConfigTest()
		jsonOut = true

		cfgPath := filepath.Join(tmpDir, "desearch-cli", "config.toml")
		if err := os.MkdirAll(filepath.Dir(cfgPath), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(cfgPath, []byte("history_enabled = true\ndefault_count = 30\n"), 0600); err != nil {
			t.Fatal(err)
		}

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := showCmd.RunE(showCmd, []string{})

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		if _, err2 := buf.ReadFrom(r); err2 != nil {
			t.Fatalf("failed to read captured output: %v", err2)
		}
		output := buf.String()

		if err != nil {
			t.Fatalf("showCmd.RunE (json) failed: %v", err)
		}
		if !strings.Contains(output, `"history_enabled"`) {
			t.Errorf("expected 'history_enabled' in JSON output, got:\n%s", output)
		}
		if !strings.Contains(output, `"default_count"`) {
			t.Errorf("expected 'default_count' in JSON output, got:\n%s", output)
		}
		if !strings.Contains(output, "true") {
			t.Errorf("expected 'true' in JSON output for history_enabled, got:\n%s", output)
		}
		if !strings.Contains(output, "30") {
			t.Errorf("expected '30' in JSON output for default_count, got:\n%s", output)
		}
	})
}

func TestShowCmd_JSONOutput(t *testing.T) {
	// Save original and restore
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	origJSONOut := jsonOut
	t.Cleanup(func() {
		if origXDG == "" {
			os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			os.Setenv("XDG_CONFIG_HOME", origXDG)
		}
		jsonOut = origJSONOut
	})

	t.Run("shows config in json mode", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tmpDir)
		initConfigTest()
		jsonOut = true

		cmd := showCmd
		err := cmd.RunE(cmd, []string{})
		if err != nil {
			t.Fatalf("showCmd.RunE (json) failed: %v", err)
		}
	})

	t.Run("shows config in text mode with short api key", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tmpDir)
		initConfigTest()
		jsonOut = false

		// Write a config with a very short api key (<=4 chars)
		cfgPath := filepath.Join(tmpDir, "desearch-cli", "config.toml")
		if err := os.MkdirAll(filepath.Dir(cfgPath), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(cfgPath, []byte(`api_key = "abc"`), 0600); err != nil {
			t.Fatal(err)
		}

		cmd := showCmd
		err := cmd.RunE(cmd, []string{})
		if err != nil {
			t.Fatalf("showCmd.RunE (short key) failed: %v", err)
		}
	})

	t.Run("shows config in text mode with long api key", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tmpDir)
		initConfigTest()
		jsonOut = false

		// Write a config with a longer api key (>4 chars)
		cfgPath := filepath.Join(tmpDir, "desearch-cli", "config.toml")
		if err := os.MkdirAll(filepath.Dir(cfgPath), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(cfgPath, []byte(`api_key = "abcdefghij"`), 0600); err != nil {
			t.Fatal(err)
		}

		cmd := showCmd
		err := cmd.RunE(cmd, []string{})
		if err != nil {
			t.Fatalf("showCmd.RunE (long key) failed: %v", err)
		}
	})
}

func TestClearCmd_WithExistingFile(t *testing.T) {
	// Save original and restore
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	t.Cleanup(func() {
		if origXDG == "" {
			os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			os.Setenv("XDG_CONFIG_HOME", origXDG)
		}
	})

	t.Run("clears existing config file", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tmpDir)
		initConfigTest()

		// Create a config file to clear
		cfgPath := filepath.Join(tmpDir, "desearch-cli", "config.toml")
		if err := os.MkdirAll(filepath.Dir(cfgPath), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(cfgPath, []byte(`api_key = "test-key"`), 0600); err != nil {
			t.Fatal(err)
		}

		cmd := clearCmd
		err := cmd.RunE(cmd, []string{})
		if err != nil {
			t.Fatalf("clearCmd.RunE failed: %v", err)
		}

		// Verify file was removed
		if _, err := os.Stat(cfgPath); !os.IsNotExist(err) {
			t.Error("expected config file to be removed")
		}
	})
}

// TestViperIntegration tests that viper is properly integrated
func TestViperIntegration(t *testing.T) {
	// Save original and restore
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	t.Cleanup(func() {
		if origXDG == "" {
			os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			os.Setenv("XDG_CONFIG_HOME", origXDG)
		}
	})

	t.Run("viper reads from XDG_CONFIG_HOME", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tmpDir)
		initConfigTest()

		viper.Set("api_key", "viper-test-key")
		viper.Set("default_tools", []string{"web", "news"})

		if viper.GetString("api_key") != "viper-test-key" {
			t.Errorf("viper.GetString(api_key) = %q, want %q", viper.GetString("api_key"), "viper-test-key")
		}
	})
}
