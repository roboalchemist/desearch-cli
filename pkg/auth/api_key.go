package auth

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

// Config holds the application configuration.
type Config struct {
	APIKey            string   `toml:"api_key"`
	DefaultTools      []string `toml:"default_tools"`
	DefaultDateFilter string   `toml:"default_date_filter"`
	DefaultCount      int      `toml:"default_count"`
}

// ConfigPath returns the XDG-compliant config file path.
func ConfigPath() (string, error) {
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("could not determine home directory: %w", err)
		}
		xdgConfigHome = filepath.Join(home, ".config")
	}
	return filepath.Join(xdgConfigHome, "desearch-cli", "config.toml"), nil
}

// LoadConfig reads the configuration from the XDG config path.
// If the config file does not exist, it returns an empty Config and no error.
func LoadConfig() (*Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	return &cfg, nil
}

// SaveConfig writes the configuration to the XDG config path.
// It creates the directory structure if it does not exist and sets permissions to 0600.
func SaveConfig(c *Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := toml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// GetAPIKey returns the API key from the environment or config file.
// Environment variable DESEARCH_API_KEY takes precedence over config file.
// Returns an empty string if the key cannot be loaded.
func GetAPIKey() string {
	if key := os.Getenv("DESEARCH_API_KEY"); key != "" {
		return key
	}
	cfg, err := LoadConfig()
	if err != nil {
		return ""
	}
	return cfg.APIKey
}
