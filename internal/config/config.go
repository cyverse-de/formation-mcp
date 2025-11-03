// Package config provides configuration management for the Formation MCP server.
// It supports loading configuration from environment variables, YAML config files,
// and command-line flags with proper precedence handling.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds the configuration for the Formation MCP server.
type Config struct {
	// BaseURL is the base URL of the Formation service (required)
	BaseURL string `yaml:"base_url"`

	// Token is a pre-obtained JWT token for authentication
	Token string `yaml:"token"`

	// Username for username/password authentication
	Username string `yaml:"username"`

	// Password for username/password authentication
	Password string `yaml:"password"`

	// LogLevel controls the logging verbosity (debug, info, warn, error)
	LogLevel string `yaml:"log_level"`

	// LogJSON enables JSON-formatted logging output
	LogJSON bool `yaml:"log_json"`

	// MetricsAddr is the address for the metrics endpoint (empty = disabled)
	MetricsAddr string `yaml:"metrics_addr"`

	// ConfigFile is the path to the YAML config file
	ConfigFile string `yaml:"-"`

	// PollInterval is the interval for polling analysis status (in seconds)
	PollInterval int `yaml:"poll_interval"`
}

// DefaultConfig returns a Config with default values.
func DefaultConfig() *Config {
	return &Config{
		LogLevel:     "info",
		LogJSON:      false,
		MetricsAddr:  "",
		PollInterval: 5,
	}
}

// FromEnv loads configuration from environment variables.
// Returns nil if no environment variables are set.
func FromEnv() *Config {
	baseURL := os.Getenv("FORMATION_BASE_URL")
	if baseURL == "" {
		return nil
	}

	cfg := &Config{
		BaseURL:      strings.TrimSuffix(baseURL, "/"),
		Token:        os.Getenv("FORMATION_TOKEN"),
		Username:     os.Getenv("FORMATION_USERNAME"),
		Password:     os.Getenv("FORMATION_PASSWORD"),
		LogLevel:     os.Getenv("LOG_LEVEL"),
		MetricsAddr:  os.Getenv("METRICS_ADDR"),
		PollInterval: 5, // default
	}

	// Handle LOG_JSON env var
	if logJSON := os.Getenv("LOG_JSON"); logJSON == "true" || logJSON == "1" {
		cfg.LogJSON = true
	}

	return cfg
}

// FromFile loads configuration from a YAML file.
func FromFile(path string) (*Config, error) {
	if path == "" {
		return nil, nil
	}

	// Expand home directory
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(home, path[2:])
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // File doesn't exist, not an error
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Normalize BaseURL
	if cfg.BaseURL != "" {
		cfg.BaseURL = strings.TrimSuffix(cfg.BaseURL, "/")
	}

	return cfg, nil
}

// Load loads configuration with proper precedence:
// CLI flags (via cfg parameter) > environment variables > config file > defaults
func Load(cfg *Config) (*Config, error) {
	// Start with defaults
	result := DefaultConfig()

	// Try to load from config file if specified
	if cfg != nil && cfg.ConfigFile != "" {
		fileCfg, err := FromFile(cfg.ConfigFile)
		if err != nil {
			return nil, err
		}
		if fileCfg != nil {
			result = mergeConfigs(result, fileCfg)
		}
	} else {
		// Try default config locations
		for _, defaultPath := range []string{
			"~/.formation-mcp.yaml",
			"~/.config/formation-mcp/config.yaml",
		} {
			fileCfg, err := FromFile(defaultPath)
			if err != nil {
				return nil, err
			}
			if fileCfg != nil {
				result = mergeConfigs(result, fileCfg)
				break
			}
		}
	}

	// Load from environment variables
	envCfg := FromEnv()
	if envCfg != nil {
		result = mergeConfigs(result, envCfg)
	}

	// Apply CLI flags (highest precedence)
	if cfg != nil {
		result = mergeConfigs(result, cfg)
	}

	// Validate the final configuration
	if err := result.Validate(); err != nil {
		return nil, err
	}

	return result, nil
}

// Validate checks that the configuration is valid.
func (c *Config) Validate() error {
	if c.BaseURL == "" {
		return errors.New("FORMATION_BASE_URL is required")
	}

	// Must have either a token or username+password
	hasToken := c.Token != ""
	hasCredentials := c.Username != "" && c.Password != ""

	if !hasToken && !hasCredentials {
		return errors.New("either FORMATION_TOKEN or FORMATION_USERNAME+FORMATION_PASSWORD must be provided")
	}

	// Validate log level
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLevels[strings.ToLower(c.LogLevel)] {
		return fmt.Errorf("invalid log level: %s (must be debug, info, warn, or error)", c.LogLevel)
	}

	// Normalize log level to lowercase
	c.LogLevel = strings.ToLower(c.LogLevel)

	return nil
}

// mergeConfigs merges two configs, with values from 'override' taking precedence.
// Only non-zero values from 'override' are used.
func mergeConfigs(base, override *Config) *Config {
	result := *base

	if override.BaseURL != "" {
		result.BaseURL = override.BaseURL
	}
	if override.Token != "" {
		result.Token = override.Token
	}
	if override.Username != "" {
		result.Username = override.Username
	}
	if override.Password != "" {
		result.Password = override.Password
	}
	if override.LogLevel != "" {
		result.LogLevel = override.LogLevel
	}
	if override.LogJSON {
		result.LogJSON = override.LogJSON
	}
	if override.MetricsAddr != "" {
		result.MetricsAddr = override.MetricsAddr
	}
	if override.ConfigFile != "" {
		result.ConfigFile = override.ConfigFile
	}
	if override.PollInterval > 0 {
		result.PollInterval = override.PollInterval
	}

	return &result
}
