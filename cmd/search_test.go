package cmd

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestSearchCmdConfiguration(t *testing.T) {
	// Test the command metadata
	assert.Equal(t, "search [query]", searchCmd.Use)
	assert.Equal(t, "Search notes", searchCmd.Short)
	assert.NotNil(t, searchCmd.Args)
}

func TestSearchAuthValidation(t *testing.T) {
	// Create a temporary directory for test config files
	tempDir, err := os.MkdirTemp("", "evernote-cli-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Save original configPath and restore after test
	originalConfigPath := configPath
	defer func() { configPath = originalConfigPath }()

	t.Run("checkAuth with no token", func(t *testing.T) {
		// Setup config without valid token
		configPath = filepath.Join(tempDir, "no-auth-config.json")
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

	t.Run("checkAuth with expired token", func(t *testing.T) {
		configPath = filepath.Join(tempDir, "expired-token-config.json")
		testConfig := Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			Token: &oauth2.Token{
				AccessToken: "expired-token",
				TokenType:   "Bearer",
				Expiry:      time.Now().Add(-time.Hour), // Expired
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
}

func TestSearchURLEncoding(t *testing.T) {
	// Test URL encoding functionality
	testCases := []struct {
		input    string
		expected string
	}{
		{"simple query", "simple+query"},
		{"query with spaces", "query+with+spaces"},
		{"query+with+plus", "query%2Bwith%2Bplus"},
		{"query&with&ampersand", "query%26with%26ampersand"},
		{"query=with=equals", "query%3Dwith%3Dequals"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			encoded := url.QueryEscape(tc.input)
			assert.Equal(t, tc.expected, encoded)
		})
	}
}

func TestSearchJSONFormatting(t *testing.T) {
	// Test JSON output formatting without making actual HTTP calls
	testData := map[string]interface{}{
		"notes": []map[string]interface{}{
			{"title": "Test Note", "id": "note-1"},
		},
		"total": 1,
	}

	t.Run("json encoding with indentation", func(t *testing.T) {
		var output bytes.Buffer

		// Test the JSON encoding logic used in search command
		enc := json.NewEncoder(&output)
		enc.SetIndent("", "  ")
		err := enc.Encode(testData)
		require.NoError(t, err)

		result := output.String()
		assert.Contains(t, result, "\"notes\"")
		assert.Contains(t, result, "\"Test Note\"")
		assert.Contains(t, result, "  ") // Should be indented
	})

	t.Run("json marshal with indentation", func(t *testing.T) {
		// Test the alternative JSON formatting used in search command
		bytes, err := json.MarshalIndent(testData, "", "  ")
		require.NoError(t, err)

		result := string(bytes)
		assert.Contains(t, result, "\"notes\"")
		assert.Contains(t, result, "\"Test Note\"")
		assert.Contains(t, result, "  ")                 // Should be indented
		assert.False(t, strings.HasSuffix(result, "\n")) // MarshalIndent doesn't add final newline
	})
}

func TestSearchHTTPRequest(t *testing.T) {
	// Test HTTP request construction (mocking the external API)
	tempDir, err := os.MkdirTemp("", "evernote-cli-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Save original configPath and restore after test
	originalConfigPath := configPath
	defer func() { configPath = originalConfigPath }()

	// Setup valid config
	configPath = filepath.Join(tempDir, "valid-config.json")
	testConfig := Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		Token: &oauth2.Token{
			AccessToken: "test-access-token",
			TokenType:   "Bearer",
			Expiry:      time.Now().Add(time.Hour),
		},
	}
	data, err := json.MarshalIndent(testConfig, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(configPath, data, 0600)
	require.NoError(t, err)

	t.Run("mock API response", func(t *testing.T) {
		// Create a mock server to test the HTTP request
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify the request format
			assert.Equal(t, "GET", r.Method)
			assert.Contains(t, r.URL.Path, "/v1/search")
			assert.Equal(t, "Bearer test-access-token", r.Header.Get("Authorization"))

			// Return mock response
			mockResponse := map[string]interface{}{
				"notes": []map[string]interface{}{
					{"title": "Mock Note", "id": "mock-1"},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(mockResponse)
		}))
		defer server.Close()

		// Test making a request to our mock server
		token, err := checkAuth()
		require.NoError(t, err)

		req, err := http.NewRequest("GET", server.URL+"/v1/search?query=test", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var data interface{}
		err = json.NewDecoder(resp.Body).Decode(&data)
		require.NoError(t, err)

		// Verify we got the expected mock data
		dataMap := data.(map[string]interface{})
		assert.Contains(t, dataMap, "notes")
	})
}

func TestSearchCmdIntegration(t *testing.T) {
	// Test that search command is properly registered
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "search" {
			found = true
			assert.Equal(t, "Search notes", cmd.Short)
			break
		}
	}
	assert.True(t, found, "search command should be registered with root command")
}
