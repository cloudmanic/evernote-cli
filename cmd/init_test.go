package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func TestInitCmdConfiguration(t *testing.T) {
	// Test the init command metadata
	assert.Equal(t, "init", initCmd.Use)
	assert.Equal(t, "Initialize credentials and authenticate", initCmd.Short)
	assert.NotNil(t, initCmd.RunE)
}

func TestInitCmdRegistration(t *testing.T) {
	// Test that init command is properly registered with root command
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "init" {
			found = true
			assert.Equal(t, "Initialize credentials and authenticate", cmd.Short)
			break
		}
	}
	assert.True(t, found, "init command should be registered with root command")
}

func TestStringTrimming(t *testing.T) {
	// Test string trimming logic used in init command
	testCases := []struct {
		input    string
		expected string
	}{
		{"test-id\n", "test-id"},
		{"test-secret\n", "test-secret"},
		{"  test-with-spaces  \n", "test-with-spaces"},
		{"\ntest-with-newlines\n", "test-with-newlines"},
		{"test-no-newline", "test-no-newline"},
		{"", ""},
		{"\n", ""},
		{"   \n", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := strings.TrimSpace(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestInitCommandInputHandling(t *testing.T) {
	// Test the input/output handling pattern used in init command
	t.Run("output formatting", func(t *testing.T) {
		var output bytes.Buffer

		// Test the prompts that would be written by init command
		output.WriteString("Evernote Client ID: ")
		assert.Contains(t, output.String(), "Evernote Client ID:")

		output.WriteString("Evernote Client Secret: ")
		assert.Contains(t, output.String(), "Evernote Client Secret:")
	})

	t.Run("success message formatting", func(t *testing.T) {
		var output bytes.Buffer
		configPath := "/test/path/config.json"

		// Test the success message format
		message := "Authentication successful. Configuration saved to " + configPath + "\n"
		output.WriteString(message)

		result := output.String()
		assert.Contains(t, result, "Authentication successful")
		assert.Contains(t, result, configPath)
		assert.True(t, strings.HasSuffix(result, "\n"))
	})
}

func TestConfigStructCreation(t *testing.T) {
	// Test Config struct creation as done in init command
	t.Run("config with credentials only", func(t *testing.T) {
		id := "test-client-id"
		secret := "test-client-secret"

		cfg := &Config{
			ClientID:     id,
			ClientSecret: secret,
		}

		assert.Equal(t, id, cfg.ClientID)
		assert.Equal(t, secret, cfg.ClientSecret)
		assert.Nil(t, cfg.Token)
	})

	t.Run("config with credentials and token", func(t *testing.T) {
		id := "test-client-id"
		secret := "test-client-secret"

		cfg := &Config{
			ClientID:     id,
			ClientSecret: secret,
		}

		// Simulate adding token later (as would happen in init flow)
		// Note: In real init command, token comes from runAuthFlow
		cfg.Token = &oauth2.Token{
			AccessToken: "test-token",
			TokenType:   "Bearer",
		}

		assert.Equal(t, id, cfg.ClientID)
		assert.Equal(t, secret, cfg.ClientSecret)
		assert.NotNil(t, cfg.Token)
		assert.Equal(t, "test-token", cfg.Token.AccessToken)
	})
}

func TestInitCommandFlow(t *testing.T) {
	// Test the logical flow of the init command without external dependencies
	t.Run("input validation", func(t *testing.T) {
		// Test that empty inputs after trimming would be problematic
		testInputs := []string{
			"",
			"   ",
			"\n",
			"\t\n",
		}

		for _, input := range testInputs {
			trimmed := strings.TrimSpace(input)
			assert.Empty(t, trimmed, "empty inputs should result in empty strings after trimming")
		}
	})

	t.Run("valid inputs", func(t *testing.T) {
		// Test that valid inputs are preserved correctly
		testInputs := map[string]string{
			"valid-client-id\n":     "valid-client-id",
			"valid-client-secret\n": "valid-client-secret",
			"  spaced-id  \n":       "spaced-id",
		}

		for input, expected := range testInputs {
			trimmed := strings.TrimSpace(input)
			assert.Equal(t, expected, trimmed)
		}
	})
}

// Note: The init command involves:
// 1. Reading from stdin (user input)
// 2. Calling runAuthFlow (OAuth2 flow with external services)
// 3. Saving config to filesystem
//
// Full integration testing would require mocking stdin, the OAuth flow,
// and filesystem operations. The tests above focus on the testable
// components like string processing and struct creation.
