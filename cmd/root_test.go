package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary directory for test config files
	tempDir, err := os.MkdirTemp("", "evernote-cli-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Save original configPath and restore after test
	originalConfigPath := configPath
	defer func() { configPath = originalConfigPath }()

	t.Run("successful load", func(t *testing.T) {
		configPath = filepath.Join(tempDir, "test-config.json")

		// Create a test config
		testConfig := Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			Token: &oauth2.Token{
				AccessToken:  "test-access-token",
				TokenType:    "Bearer",
				RefreshToken: "test-refresh-token",
				Expiry:       time.Now().Add(time.Hour),
			},
		}

		// Write the test config to file
		data, err := json.MarshalIndent(testConfig, "", "  ")
		require.NoError(t, err)
		err = os.WriteFile(configPath, data, 0600)
		require.NoError(t, err)

		// Test loading the config
		config, err := loadConfig()
		require.NoError(t, err)
		assert.Equal(t, "test-client-id", config.ClientID)
		assert.Equal(t, "test-client-secret", config.ClientSecret)
		assert.NotNil(t, config.Token)
		assert.Equal(t, "test-access-token", config.Token.AccessToken)
		assert.Equal(t, "Bearer", config.Token.TokenType)
		assert.Equal(t, "test-refresh-token", config.Token.RefreshToken)
	})

	t.Run("file does not exist", func(t *testing.T) {
		configPath = filepath.Join(tempDir, "nonexistent-config.json")

		config, err := loadConfig()
		assert.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "no such file or directory")
	})

	t.Run("invalid JSON", func(t *testing.T) {
		configPath = filepath.Join(tempDir, "invalid-config.json")

		// Write invalid JSON
		err = os.WriteFile(configPath, []byte("invalid json content"), 0600)
		require.NoError(t, err)

		config, err := loadConfig()
		assert.Error(t, err)
		assert.Nil(t, config)
	})

	t.Run("empty file", func(t *testing.T) {
		configPath = filepath.Join(tempDir, "empty-config.json")

		// Write empty file
		err = os.WriteFile(configPath, []byte(""), 0600)
		require.NoError(t, err)

		config, err := loadConfig()
		assert.Error(t, err)
		assert.Nil(t, config)
	})
}

func TestSaveConfig(t *testing.T) {
	// Create a temporary directory for test config files
	tempDir, err := os.MkdirTemp("", "evernote-cli-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Save original configPath and restore after test
	originalConfigPath := configPath
	defer func() { configPath = originalConfigPath }()

	t.Run("successful save", func(t *testing.T) {
		configPath = filepath.Join(tempDir, "subdir", "test-config.json")

		testConfig := &Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			Token: &oauth2.Token{
				AccessToken:  "test-access-token",
				TokenType:    "Bearer",
				RefreshToken: "test-refresh-token",
				Expiry:       time.Now().Add(time.Hour),
			},
		}

		err := saveConfig(testConfig)
		require.NoError(t, err)

		// Verify the file was created
		assert.FileExists(t, configPath)

		// Verify the content is correct
		data, err := os.ReadFile(configPath)
		require.NoError(t, err)

		var savedConfig Config
		err = json.Unmarshal(data, &savedConfig)
		require.NoError(t, err)

		assert.Equal(t, testConfig.ClientID, savedConfig.ClientID)
		assert.Equal(t, testConfig.ClientSecret, savedConfig.ClientSecret)
		assert.Equal(t, testConfig.Token.AccessToken, savedConfig.Token.AccessToken)
	})

	t.Run("save config without token", func(t *testing.T) {
		configPath = filepath.Join(tempDir, "no-token-config.json")

		testConfig := &Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			Token:        nil,
		}

		err := saveConfig(testConfig)
		require.NoError(t, err)

		// Verify the file was created and token is omitted
		data, err := os.ReadFile(configPath)
		require.NoError(t, err)

		var savedConfig Config
		err = json.Unmarshal(data, &savedConfig)
		require.NoError(t, err)

		assert.Equal(t, testConfig.ClientID, savedConfig.ClientID)
		assert.Equal(t, testConfig.ClientSecret, savedConfig.ClientSecret)
		assert.Nil(t, savedConfig.Token)
	})

	t.Run("save to protected directory", func(t *testing.T) {
		// This test might fail on systems where the directory is writable (for example when running as root).
		// Skip to avoid false failures in such cases.
		if os.Geteuid() == 0 {
			t.Skip("skipping permissions test when running as root")
		}
		configPath = "/root/protected/test-config.json"

		testConfig := &Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
		}

		err := saveConfig(testConfig)
		// We expect this to fail due to permission issues
		assert.Error(t, err)
	})
}

func TestCheckAuth(t *testing.T) {
	// Create a temporary directory for test config files
	tempDir, err := os.MkdirTemp("", "evernote-cli-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Save original configPath and restore after test
	originalConfigPath := configPath
	defer func() { configPath = originalConfigPath }()

	t.Run("valid token", func(t *testing.T) {
		configPath = filepath.Join(tempDir, "valid-token-config.json")

		// Create config with valid token
		testConfig := Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			Token: &oauth2.Token{
				AccessToken: "test-access-token",
				TokenType:   "Bearer",
				Expiry:      time.Now().Add(time.Hour), // Valid for 1 hour
			},
		}

		data, err := json.MarshalIndent(testConfig, "", "  ")
		require.NoError(t, err)
		err = os.WriteFile(configPath, data, 0600)
		require.NoError(t, err)

		token, err := checkAuth()
		require.NoError(t, err)
		assert.Equal(t, "test-access-token", token)
	})

	t.Run("expired token", func(t *testing.T) {
		configPath = filepath.Join(tempDir, "expired-token-config.json")

		// Create config with expired token
		testConfig := Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			Token: &oauth2.Token{
				AccessToken: "expired-access-token",
				TokenType:   "Bearer",
				Expiry:      time.Now().Add(-time.Hour), // Expired 1 hour ago
			},
		}

		data, err := json.MarshalIndent(testConfig, "", "  ")
		require.NoError(t, err)
		err = os.WriteFile(configPath, data, 0600)
		require.NoError(t, err)

		token, err := checkAuth()
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Contains(t, err.Error(), "no valid token found")
	})

	t.Run("nil token", func(t *testing.T) {
		configPath = filepath.Join(tempDir, "nil-token-config.json")

		// Create config without token
		testConfig := Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			Token:        nil,
		}

		data, err := json.MarshalIndent(testConfig, "", "  ")
		require.NoError(t, err)
		err = os.WriteFile(configPath, data, 0600)
		require.NoError(t, err)

		token, err := checkAuth()
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Contains(t, err.Error(), "no valid token found")
	})

	t.Run("config file does not exist", func(t *testing.T) {
		configPath = filepath.Join(tempDir, "nonexistent-config.json")

		token, err := checkAuth()
		assert.Error(t, err)
		assert.Empty(t, token)
		assert.Contains(t, err.Error(), "could not read config")
	})
}

func TestConfig_Struct(t *testing.T) {
	t.Run("json marshaling and unmarshaling", func(t *testing.T) {
		original := Config{
			ClientID:     "test-id",
			ClientSecret: "test-secret",
			Token: &oauth2.Token{
				AccessToken:  "access-token",
				RefreshToken: "refresh-token",
				TokenType:    "Bearer",
				Expiry:       time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC),
			},
		}

		// Marshal to JSON
		data, err := json.Marshal(original)
		require.NoError(t, err)

		// Unmarshal back
		var unmarshaled Config
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		// Verify all fields
		assert.Equal(t, original.ClientID, unmarshaled.ClientID)
		assert.Equal(t, original.ClientSecret, unmarshaled.ClientSecret)
		assert.NotNil(t, unmarshaled.Token)
		assert.Equal(t, original.Token.AccessToken, unmarshaled.Token.AccessToken)
		assert.Equal(t, original.Token.RefreshToken, unmarshaled.Token.RefreshToken)
		assert.Equal(t, original.Token.TokenType, unmarshaled.Token.TokenType)
		// Note: Time comparison might need to account for precision differences
		assert.True(t, original.Token.Expiry.Equal(unmarshaled.Token.Expiry))
	})

	t.Run("json omitempty for nil token", func(t *testing.T) {
		config := Config{
			ClientID:     "test-id",
			ClientSecret: "test-secret",
			Token:        nil,
		}

		data, err := json.Marshal(config)
		require.NoError(t, err)

		// Should not contain "token" field when nil
		jsonStr := string(data)
		assert.NotContains(t, jsonStr, "token")
		assert.Contains(t, jsonStr, "client_id")
		assert.Contains(t, jsonStr, "client_secret")
	})
}
