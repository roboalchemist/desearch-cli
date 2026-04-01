package auth

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
		path, err := ConfigPath()
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(tmpDir, "desearch-cli", "config.toml"), path)
	})

	t.Run("falls back to ~/.config when XDG_CONFIG_HOME not set", func(t *testing.T) {
		os.Unsetenv("XDG_CONFIG_HOME")
		home, err := os.UserHomeDir()
		require.NoError(t, err)
		path, err := ConfigPath()
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(home, ".config", "desearch-cli", "config.toml"), path)
	})
}

func TestLoadConfig(t *testing.T) {
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	t.Cleanup(func() {
		if origXDG == "" {
			os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			os.Setenv("XDG_CONFIG_HOME", origXDG)
		}
	})

	t.Run("returns empty config when file does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tmpDir)
		cfg, err := LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, "", cfg.APIKey)
		assert.Equal(t, []string(nil), cfg.DefaultTools)
		assert.Equal(t, "", cfg.DefaultDateFilter)
		assert.Equal(t, 0, cfg.DefaultCount)
	})

	t.Run("loads existing config correctly", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tmpDir)

		cfg := &Config{
			APIKey:            "test-api-key-123",
			DefaultTools:      []string{"web", "news"},
			DefaultDateFilter: "7d",
			DefaultCount:      20,
		}
		data, err := toml.Marshal(cfg)
		require.NoError(t, err)

		configDir := filepath.Join(tmpDir, "desearch-cli")
		err = os.MkdirAll(configDir, 0700)
		require.NoError(t, err)

		configFile := filepath.Join(configDir, "config.toml")
		err = os.WriteFile(configFile, data, 0600)
		require.NoError(t, err)

		loaded, err := LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, "test-api-key-123", loaded.APIKey)
		assert.Equal(t, []string{"web", "news"}, loaded.DefaultTools)
		assert.Equal(t, "7d", loaded.DefaultDateFilter)
		assert.Equal(t, 20, loaded.DefaultCount)
	})
}

func TestSaveConfig(t *testing.T) {
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	t.Cleanup(func() {
		if origXDG == "" {
			os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			os.Setenv("XDG_CONFIG_HOME", origXDG)
		}
	})

	t.Run("creates directory and file with 0600 permissions", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tmpDir)

		cfg := &Config{
			APIKey:            "my-secret-key",
			DefaultTools:      []string{"web"},
			DefaultDateFilter: "30d",
			DefaultCount:      50,
		}
		err := SaveConfig(cfg)
		require.NoError(t, err)

		configFile := filepath.Join(tmpDir, "desearch-cli", "config.toml")
		info, err := os.Stat(configFile)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0600), info.Mode().Perm())

		data, err := os.ReadFile(configFile)
		require.NoError(t, err)

		var loaded Config
		err = toml.Unmarshal(data, &loaded)
		require.NoError(t, err)
		assert.Equal(t, "my-secret-key", loaded.APIKey)
		assert.Equal(t, []string{"web"}, loaded.DefaultTools)
		assert.Equal(t, "30d", loaded.DefaultDateFilter)
		assert.Equal(t, 50, loaded.DefaultCount)
	})

	t.Run("returns error when directory cannot be created", func(t *testing.T) {
		// Use a path that exists as a read-only file, not a directory,
		// so os.MkdirAll fails.
		tmpDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tmpDir)

		// Create a file at the config directory path so MkdirAll fails
		readonlyFile := filepath.Join(tmpDir, "desearch-cli")
		f, err := os.Create(readonlyFile)
		require.NoError(t, err)
		f.Close()

		// Make the parent unreadable so MkdirAll can't traverse
		if err := os.Chmod(tmpDir, 0000); err != nil {
			t.Fatal(err)
		}

		cfg := &Config{APIKey: "some-key"}
		err = SaveConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "creating config directory")

		// Restore so TempDir cleanup can proceed
		if err := os.Chmod(tmpDir, 0700); err != nil {
			t.Fatal(err)
		}
		os.Remove(readonlyFile)
	})

	t.Run("returns error when ConfigPath fails", func(t *testing.T) {
		// Clear both XDG_CONFIG_HOME and the HOME env var so os.UserHomeDir() fails.
		origXDG := os.Getenv("XDG_CONFIG_HOME")
		origHOME := os.Getenv("HOME")
		t.Cleanup(func() {
			if origXDG == "" {
				os.Unsetenv("XDG_CONFIG_HOME")
			} else {
				os.Setenv("XDG_CONFIG_HOME", origXDG)
			}
			if origHOME == "" {
				os.Unsetenv("HOME")
			} else {
				os.Setenv("HOME", origHOME)
			}
		})

		os.Unsetenv("XDG_CONFIG_HOME")
		os.Unsetenv("HOME")

		cfg := &Config{APIKey: "some-key"}
		err := SaveConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "could not determine home directory")
	})

	t.Run("returns error when file cannot be written", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tmpDir)

		// Pre-create the directory with no write permission
		configDir := filepath.Join(tmpDir, "desearch-cli")
		err := os.MkdirAll(configDir, 0555) // read-only directory
		require.NoError(t, err)

		cfg := &Config{APIKey: "some-key"}
		err = SaveConfig(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "writing config file")

		// Restore permissions so TempDir cleanup can proceed
		if err := os.Chmod(configDir, 0700); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("overwrites existing config", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tmpDir)

		cfg1 := &Config{APIKey: "key-1"}
		err := SaveConfig(cfg1)
		require.NoError(t, err)

		cfg2 := &Config{APIKey: "key-2", DefaultCount: 100}
		err = SaveConfig(cfg2)
		require.NoError(t, err)

		loaded, err := LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, "key-2", loaded.APIKey)
		assert.Equal(t, 100, loaded.DefaultCount)
	})
}

func TestGetAPIKey(t *testing.T) {
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	origEnvKey := os.Getenv("DESEARCH_API_KEY")
	t.Cleanup(func() {
		if origXDG == "" {
			os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			os.Setenv("XDG_CONFIG_HOME", origXDG)
		}
		if origEnvKey == "" {
			os.Unsetenv("DESEARCH_API_KEY")
		} else {
			os.Setenv("DESEARCH_API_KEY", origEnvKey)
		}
	})
	// Unset env var so tests that check config-file behavior are not polluted.
	os.Unsetenv("DESEARCH_API_KEY")

	t.Run("returns empty string when config does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tmpDir)
		assert.Equal(t, "", GetAPIKey())
	})

	t.Run("returns API key from config", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tmpDir)

		cfg := &Config{APIKey: "my-api-key"}
		err := SaveConfig(cfg)
		require.NoError(t, err)

		assert.Equal(t, "my-api-key", GetAPIKey())
	})

	t.Run("env var takes precedence over config file", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tmpDir)

		// Write config file with one API key
		cfg := &Config{APIKey: "config-file-key"}
		err := SaveConfig(cfg)
		require.NoError(t, err)

		// Set env var with different key
		os.Setenv("DESEARCH_API_KEY", "env-override-key")

		// Env var should win
		assert.Equal(t, "env-override-key", GetAPIKey())

		os.Unsetenv("DESEARCH_API_KEY")
	})

	t.Run("returns empty string when LoadConfig errors", func(t *testing.T) {
		// XDG_CONFIG_HOME points to a non-existent path on a read-only root
		// so LoadConfig returns an error. GetAPIKey should gracefully return "".
		tmpDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tmpDir)

		// Remove all permissions from tmpDir so os.ReadFile fails with permission denied
		if err := os.Chmod(tmpDir, 0000); err != nil {
			t.Fatal(err)
		}

		result := GetAPIKey()
		assert.Equal(t, "", result)

		// Restore permissions so TempDir cleanup can proceed
		if err := os.Chmod(tmpDir, 0700); err != nil {
			t.Fatal(err)
		}
	})
}

func TestLoadConfig_Persistence(t *testing.T) {
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	t.Cleanup(func() {
		if origXDG == "" {
			os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			os.Setenv("XDG_CONFIG_HOME", origXDG)
		}
	})

	t.Run("config persists across invocations", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tmpDir)

		cfg := &Config{
			APIKey:            "persist-key",
			DefaultTools:      []string{"web", "news", "social"},
			DefaultDateFilter: "90d",
			DefaultCount:      25,
		}
		err := SaveConfig(cfg)
		require.NoError(t, err)

		// Simulate new invocation by loading config
		loaded, err := LoadConfig()
		require.NoError(t, err)
		assert.Equal(t, "persist-key", loaded.APIKey)
		assert.Equal(t, []string{"web", "news", "social"}, loaded.DefaultTools)
		assert.Equal(t, "90d", loaded.DefaultDateFilter)
		assert.Equal(t, 25, loaded.DefaultCount)
	})
}

func TestLoadConfig_InvalidTOML(t *testing.T) {
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	t.Cleanup(func() {
		if origXDG == "" {
			os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			os.Setenv("XDG_CONFIG_HOME", origXDG)
		}
	})

	t.Run("returns error for invalid TOML content", func(t *testing.T) {
		tmpDir := t.TempDir()
		os.Setenv("XDG_CONFIG_HOME", tmpDir)

		// Write invalid TOML content
		configFile := filepath.Join(tmpDir, "desearch-cli", "config.toml")
		err := os.MkdirAll(filepath.Dir(configFile), 0700)
		require.NoError(t, err)

		err = os.WriteFile(configFile, []byte("invalid toml {[["), 0600)
		require.NoError(t, err)

		_, err = LoadConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parsing config file")
	})
}
