package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

func TestTagsCommand(t *testing.T) {
	// Setup test config directory
	tempDir := t.TempDir()
	originalConfigPath := configPath
	configPath = filepath.Join(tempDir, "auth.json")
	defer func() { configPath = originalConfigPath }()

	// Create a test config with a valid token
	testConfig := &Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		Token: &oauth2.Token{
			AccessToken: "test-token",
			TokenType:   "Bearer",
		},
	}
	// Make token appear valid by setting it as non-expired
	testConfig.Token.Expiry = time.Now().Add(time.Hour)

	// Save test config
	if err := saveConfig(testConfig); err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}

	tests := []struct {
		name           string
		jsonFlag       bool
		responseBody   string
		statusCode     int
		expectedOutput string
		expectError    bool
	}{
		{
			name:       "successful tags list with formatting",
			jsonFlag:   false,
			statusCode: http.StatusOK,
			responseBody: `{
				"tags": [
					{
						"name": "Work",
						"guid": "tag-12345"
					},
					{
						"name": "Personal",
						"guid": "tag-67890"
					}
				]
			}`,
			expectedOutput: "Found 2 tag(s):\n\n1. Work\n   GUID: tag-12345\n\n2. Personal\n   GUID: tag-67890\n\n",
			expectError:    false,
		},
		{
			name:       "successful tags list with JSON output",
			jsonFlag:   true,
			statusCode: http.StatusOK,
			responseBody: `{
				"tags": [
					{
						"name": "Work",
						"guid": "tag-12345"
					}
				]
			}`,
			expectedOutput: `{
  "tags": [
    {
      "guid": "tag-12345",
      "name": "Work"
    }
  ]
}`,
			expectError: false,
		},
		{
			name:           "empty tags list",
			jsonFlag:       false,
			statusCode:     http.StatusOK,
			responseBody:   `{"tags": []}`,
			expectedOutput: "No tags found.\n",
			expectError:    false,
		},
		{
			name:        "API error response",
			jsonFlag:    false,
			statusCode:  http.StatusUnauthorized,
			expectError: true,
		},
		{
			name:       "malformed response falls back to JSON",
			jsonFlag:   false,
			statusCode: http.StatusOK,
			responseBody: `{
				"invalid": "response"
			}`,
			expectedOutput: `{
  "invalid": "response"
}`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify the request
				if r.Method != "GET" {
					t.Errorf("Expected GET request, got %s", r.Method)
				}
				if r.URL.Path != "/v1/tags" {
					t.Errorf("Expected path /v1/tags, got %s", r.URL.Path)
				}
				if r.Header.Get("Authorization") != "Bearer test-token" {
					t.Errorf("Expected Authorization header 'Bearer test-token', got %s", r.Header.Get("Authorization"))
				}

				w.WriteHeader(tt.statusCode)
				if tt.responseBody != "" {
					w.Write([]byte(tt.responseBody))
				}
			}))
			defer server.Close()

			// Create a test command with modified URL
			cmd := &cobra.Command{
				Use:   "tags",
				Short: "List all tags",
				RunE: func(cmd *cobra.Command, args []string) error {
					token, err := checkAuth()
					if err != nil {
						return err
					}

					req, err := http.NewRequest("GET", server.URL+"/v1/tags", nil)
					if err != nil {
						return err
					}
					req.Header.Set("Authorization", "Bearer "+token)

					resp, err := http.DefaultClient.Do(req)
					if err != nil {
						return err
					}
					defer resp.Body.Close()
					if resp.StatusCode != http.StatusOK {
						return fmt.Errorf("unexpected status: %s", resp.Status)
					}

					var data interface{}
					if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
						return err
					}

					if tt.jsonFlag {
						jsonFlag = true
						enc := json.NewEncoder(cmd.OutOrStdout())
						enc.SetIndent("", "  ")
						return enc.Encode(data)
					} else {
						jsonFlag = false
						return formatTags(cmd, data)
					}
				},
			}

			// Capture output
			var buf bytes.Buffer
			cmd.SetOut(&buf)

			// Execute command
			err := cmd.Execute()

			// Check results
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				output := buf.String()
				if !strings.Contains(output, strings.TrimSpace(tt.expectedOutput)) {
					t.Errorf("Expected output to contain:\n%s\nGot:\n%s", tt.expectedOutput, output)
				}
			}
		})
	}
}

func TestFormatTags(t *testing.T) {
	tests := []struct {
		name           string
		input          interface{}
		expectedOutput string
	}{
		{
			name: "valid tags response",
			input: map[string]interface{}{
				"tags": []interface{}{
					map[string]interface{}{
						"name": "Important",
						"guid": "tag-123",
					},
					map[string]interface{}{
						"name": "Project",
						"guid": "tag-456",
					},
				},
			},
			expectedOutput: "Found 2 tag(s):\n\n1. Important\n   GUID: tag-123\n\n2. Project\n   GUID: tag-456\n\n",
		},
		{
			name:           "invalid data type",
			input:          "not a map",
			expectedOutput: "\"not a map\"\n",
		},
		{
			name: "missing tags array",
			input: map[string]interface{}{
				"other": "data",
			},
			expectedOutput: "{\n  \"other\": \"data\"\n}\n",
		},
		{
			name: "empty tags array",
			input: map[string]interface{}{
				"tags": []interface{}{},
			},
			expectedOutput: "No tags found.\n",
		},
		{
			name: "tags without GUID",
			input: map[string]interface{}{
				"tags": []interface{}{
					map[string]interface{}{
						"name": "No GUID Tag",
					},
				},
			},
			expectedOutput: "Found 1 tag(s):\n\n1. No GUID Tag\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			cmd := &cobra.Command{}
			cmd.SetOut(&buf)

			err := formatTags(cmd, tt.input)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			output := buf.String()
			if output != tt.expectedOutput {
				t.Errorf("Expected:\n%s\nGot:\n%s", tt.expectedOutput, output)
			}
		})
	}
}

func TestTagsCmdConfiguration(t *testing.T) {
	// Test the command metadata
	if tagsCmd.Use != "tags" {
		t.Errorf("Expected Use to be 'tags', got %s", tagsCmd.Use)
	}
	if tagsCmd.Short != "List all tags" {
		t.Errorf("Expected Short to be 'List all tags', got %s", tagsCmd.Short)
	}
	if tagsCmd.RunE == nil {
		t.Error("Expected RunE to be set")
	}
}

func TestTagsCommandRegistration(t *testing.T) {
	// Test that tags command is properly registered with root command
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "tags" {
			found = true
			if cmd.Short != "List all tags" {
				t.Errorf("Expected Short to be 'List all tags', got %s", cmd.Short)
			}
			break
		}
	}
	if !found {
		t.Error("tags command should be registered with root command")
	}
}
