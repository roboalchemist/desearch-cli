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
	t.Cleanup(func() {
		if origXDG == "" {
			os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			os.Setenv("XDG_CONFIG_HOME", origXDG)
		}
	})

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
