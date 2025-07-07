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

func TestNotebooksCommand(t *testing.T) {
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
			name:       "successful notebooks list with formatting",
			jsonFlag:   false,
			statusCode: http.StatusOK,
			responseBody: `{
				"notebooks": [
					{
						"name": "Default Notebook",
						"guid": "12345",
						"defaultNotebook": true
					},
					{
						"name": "Work Notes",
						"guid": "67890",
						"defaultNotebook": false
					}
				]
			}`,
			expectedOutput: "Found 2 notebook(s):\n\n1. Default Notebook (default)\n   GUID: 12345\n\n2. Work Notes\n   GUID: 67890\n\n",
			expectError:    false,
		},
		{
			name:       "successful notebooks list with JSON output",
			jsonFlag:   true,
			statusCode: http.StatusOK,
			responseBody: `{
				"notebooks": [
					{
						"name": "Default Notebook",
						"guid": "12345",
						"defaultNotebook": true
					}
				]
			}`,
			expectedOutput: `{
  "notebooks": [
    {
      "defaultNotebook": true,
      "guid": "12345",
      "name": "Default Notebook"
    }
  ]
}`,
			expectError: false,
		},
		{
			name:           "empty notebooks list",
			jsonFlag:       false,
			statusCode:     http.StatusOK,
			responseBody:   `{"notebooks": []}`,
			expectedOutput: "No notebooks found.\n",
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
				if r.URL.Path != "/v1/notebooks" {
					t.Errorf("Expected path /v1/notebooks, got %s", r.URL.Path)
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
				Use:   "notebooks",
				Short: "List all notebooks",
				RunE: func(cmd *cobra.Command, args []string) error {
					token, err := checkAuth()
					if err != nil {
						return err
					}

					// Use test server URL instead of real API
					req, err := http.NewRequest("GET", server.URL+"/v1/notebooks", nil)
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
					}

					jsonFlag = false
					return formatNotebooks(cmd, data)
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

func TestFormatNotebooks(t *testing.T) {
	tests := []struct {
		name           string
		input          interface{}
		expectedOutput string
	}{
		{
			name: "valid notebooks response",
			input: map[string]interface{}{
				"notebooks": []interface{}{
					map[string]interface{}{
						"name":            "My Notebook",
						"guid":            "abc123",
						"defaultNotebook": false,
					},
				},
			},
			expectedOutput: "Found 1 notebook(s):\n\n1. My Notebook\n   GUID: abc123\n\n",
		},
		{
			name:           "invalid data type",
			input:          "invalid",
			expectedOutput: `"invalid"`,
		},
		{
			name: "missing notebooks array",
			input: map[string]interface{}{
				"other": "data",
			},
			expectedOutput: `{
  "other": "data"
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{}
			var buf bytes.Buffer
			cmd.SetOut(&buf)

			err := formatNotebooks(cmd, tt.input)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			output := buf.String()
			if !strings.Contains(output, strings.TrimSpace(tt.expectedOutput)) {
				t.Errorf("Expected output to contain:\n%s\nGot:\n%s", tt.expectedOutput, output)
			}
		})
	}
}
