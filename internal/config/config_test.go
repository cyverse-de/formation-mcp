package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, "info", cfg.LogLevel)
	assert.False(t, cfg.LogJSON)
	assert.Empty(t, cfg.MetricsAddr)
	assert.Equal(t, 5, cfg.PollInterval)
}

func TestFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected *Config
	}{
		{
			name:     "no environment variables",
			envVars:  map[string]string{},
			expected: nil,
		},
		{
			name: "base URL only",
			envVars: map[string]string{
				"FORMATION_BASE_URL": "https://example.com",
			},
			expected: &Config{
				BaseURL:      "https://example.com",
				PollInterval: 5,
			},
		},
		{
			name: "base URL with trailing slash",
			envVars: map[string]string{
				"FORMATION_BASE_URL": "https://example.com/",
			},
			expected: &Config{
				BaseURL:      "https://example.com",
				PollInterval: 5,
			},
		},
		{
			name: "with token",
			envVars: map[string]string{
				"FORMATION_BASE_URL": "https://example.com",
				"FORMATION_TOKEN":    "test-token",
			},
			expected: &Config{
				BaseURL:      "https://example.com",
				Token:        "test-token",
				PollInterval: 5,
			},
		},
		{
			name: "with credentials",
			envVars: map[string]string{
				"FORMATION_BASE_URL":  "https://example.com",
				"FORMATION_USERNAME": "testuser",
				"FORMATION_PASSWORD": "testpass",
			},
			expected: &Config{
				BaseURL:      "https://example.com",
				Username:     "testuser",
				Password:     "testpass",
				PollInterval: 5,
			},
		},
		{
			name: "with log settings",
			envVars: map[string]string{
				"FORMATION_BASE_URL": "https://example.com",
				"FORMATION_TOKEN":    "test-token",
				"LOG_LEVEL":          "debug",
				"LOG_JSON":           "true",
			},
			expected: &Config{
				BaseURL:      "https://example.com",
				Token:        "test-token",
				LogLevel:     "debug",
				LogJSON:      true,
				PollInterval: 5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set test environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			cfg := FromEnv()
			assert.Equal(t, tt.expected, cfg)
		})
	}
}

func TestFromFile(t *testing.T) {
	tests := []struct {
		name        string
		fileContent string
		expected    *Config
		expectError bool
	}{
		{
			name:     "empty path",
			expected: nil,
		},
		{
			name: "valid YAML",
			fileContent: `
base_url: https://example.com
token: test-token
log_level: debug
log_json: true
metrics_addr: :9090
poll_interval: 10
`,
			expected: &Config{
				BaseURL:      "https://example.com",
				Token:        "test-token",
				LogLevel:     "debug",
				LogJSON:      true,
				MetricsAddr:  ":9090",
				PollInterval: 10,
			},
		},
		{
			name: "with credentials",
			fileContent: `
base_url: https://example.com
username: testuser
password: testpass
`,
			expected: &Config{
				BaseURL:      "https://example.com",
				Username:     "testuser",
				Password:     "testpass",
				LogLevel:     "info",
				LogJSON:      false,
				MetricsAddr:  "",
				PollInterval: 5,
			},
		},
		{
			name:        "invalid YAML",
			fileContent: "invalid: yaml: content:",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var path string
			if tt.fileContent != "" {
				// Create temporary file
				tmpfile, err := os.CreateTemp("", "config-*.yaml")
				require.NoError(t, err)
				defer os.Remove(tmpfile.Name())

				_, err = tmpfile.WriteString(tt.fileContent)
				require.NoError(t, err)
				tmpfile.Close()

				path = tmpfile.Name()
			}

			cfg, err := FromFile(path)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, cfg)
			}
		})
	}
}

func TestFromFileNonExistent(t *testing.T) {
	cfg, err := FromFile("/nonexistent/path/config.yaml")
	assert.NoError(t, err)
	assert.Nil(t, cfg)
}

func TestFromFileHomeExpansion(t *testing.T) {
	// Skip if HOME is not set (e.g., in some CI environments)
	if os.Getenv("HOME") == "" {
		t.Skip("Skipping test: HOME environment variable not set")
	}

	// Create temp file in home directory
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	tmpfile, err := os.CreateTemp(home, "test-config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpfile.Name())

	content := `base_url: https://example.com
token: test-token`
	_, err = tmpfile.WriteString(content)
	require.NoError(t, err)
	tmpfile.Close()

	// Use relative path with ~
	relPath := "~/" + filepath.Base(tmpfile.Name())
	cfg, err := FromFile(relPath)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "https://example.com", cfg.BaseURL)
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid with token",
			config: &Config{
				BaseURL:  "https://example.com",
				Token:    "test-token",
				LogLevel: "info",
			},
			expectError: false,
		},
		{
			name: "valid with credentials",
			config: &Config{
				BaseURL:  "https://example.com",
				Username: "testuser",
				Password: "testpass",
				LogLevel: "info",
			},
			expectError: false,
		},
		{
			name: "missing base URL",
			config: &Config{
				Token:    "test-token",
				LogLevel: "info",
			},
			expectError: true,
			errorMsg:    "FORMATION_BASE_URL is required",
		},
		{
			name: "missing auth",
			config: &Config{
				BaseURL:  "https://example.com",
				LogLevel: "info",
			},
			expectError: true,
			errorMsg:    "either FORMATION_TOKEN or FORMATION_USERNAME+FORMATION_PASSWORD must be provided",
		},
		{
			name: "missing password",
			config: &Config{
				BaseURL:  "https://example.com",
				Username: "testuser",
				LogLevel: "info",
			},
			expectError: true,
			errorMsg:    "either FORMATION_TOKEN or FORMATION_USERNAME+FORMATION_PASSWORD must be provided",
		},
		{
			name: "invalid log level",
			config: &Config{
				BaseURL:  "https://example.com",
				Token:    "test-token",
				LogLevel: "invalid",
			},
			expectError: true,
			errorMsg:    "invalid log level",
		},
		{
			name: "valid debug level",
			config: &Config{
				BaseURL:  "https://example.com",
				Token:    "test-token",
				LogLevel: "DEBUG",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		envVars     map[string]string
		fileContent string
		cliConfig   *Config
		expected    *Config
	}{
		{
			name: "CLI overrides all",
			envVars: map[string]string{
				"FORMATION_BASE_URL": "https://env.com",
				"FORMATION_TOKEN":    "env-token",
			},
			fileContent: `base_url: https://file.com
token: file-token`,
			cliConfig: &Config{
				BaseURL: "https://cli.com",
				Token:   "cli-token",
			},
			expected: &Config{
				BaseURL:      "https://cli.com",
				Token:        "cli-token",
				LogLevel:     "info",
				LogJSON:      false,
				MetricsAddr:  "",
				PollInterval: 5,
			},
		},
		{
			name: "env overrides file",
			envVars: map[string]string{
				"FORMATION_BASE_URL": "https://env.com",
				"FORMATION_TOKEN":    "env-token",
			},
			fileContent: `base_url: https://file.com
token: file-token`,
			cliConfig: &Config{},
			expected: &Config{
				BaseURL:      "https://env.com",
				Token:        "env-token",
				LogLevel:     "info",
				LogJSON:      false,
				MetricsAddr:  "",
				PollInterval: 5,
			},
		},
		{
			name:    "file only",
			envVars: map[string]string{},
			fileContent: `base_url: https://file.com
token: file-token
log_level: debug`,
			cliConfig: &Config{},
			expected: &Config{
				BaseURL:      "https://file.com",
				Token:        "file-token",
				LogLevel:     "debug",
				LogJSON:      false,
				MetricsAddr:  "",
				PollInterval: 5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set test environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Create config file if provided
			if tt.fileContent != "" {
				tmpfile, err := os.CreateTemp("", "config-*.yaml")
				require.NoError(t, err)
				defer os.Remove(tmpfile.Name())

				_, err = tmpfile.WriteString(tt.fileContent)
				require.NoError(t, err)
				tmpfile.Close()

				tt.cliConfig.ConfigFile = tmpfile.Name()
			}

			cfg, err := Load(tt.cliConfig)
			require.NoError(t, err)
			// Don't compare ConfigFile field as it contains temp path
			cfg.ConfigFile = ""
			assert.Equal(t, tt.expected, cfg)
		})
	}
}

func TestMergeConfigs(t *testing.T) {
	base := &Config{
		BaseURL:      "https://base.com",
		Token:        "base-token",
		LogLevel:     "info",
		PollInterval: 5,
	}

	override := &Config{
		BaseURL:  "https://override.com",
		LogLevel: "debug",
	}

	result := mergeConfigs(base, override)

	assert.Equal(t, "https://override.com", result.BaseURL)
	assert.Equal(t, "base-token", result.Token)
	assert.Equal(t, "debug", result.LogLevel)
	assert.Equal(t, 5, result.PollInterval)
}
