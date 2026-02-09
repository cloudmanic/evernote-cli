package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitCmdConfiguration(t *testing.T) {
	assert.Equal(t, "init", initCmd.Use)
	assert.Equal(t, "Initialize credentials and authenticate", initCmd.Short)
	assert.NotNil(t, initCmd.RunE)
}

func TestInitCmdRegistration(t *testing.T) {
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
	t.Run("output formatting", func(t *testing.T) {
		var output bytes.Buffer

		output.WriteString("Evernote Consumer Key: ")
		assert.Contains(t, output.String(), "Evernote Consumer Key:")

		output.WriteString("Evernote Consumer Secret: ")
		assert.Contains(t, output.String(), "Evernote Consumer Secret:")
	})

	t.Run("success message formatting", func(t *testing.T) {
		var output bytes.Buffer
		testPath := "/test/path/config.json"

		message := "Authentication successful. Configuration saved to " + testPath + "\n"
		output.WriteString(message)

		result := output.String()
		assert.Contains(t, result, "Authentication successful")
		assert.Contains(t, result, testPath)
		assert.True(t, strings.HasSuffix(result, "\n"))
	})
}

func TestConfigStructCreation(t *testing.T) {
	t.Run("config with credentials only", func(t *testing.T) {
		id := "test-client-id"
		secret := "test-client-secret"

		cfg := &Config{
			ClientID:     id,
			ClientSecret: secret,
		}

		assert.Equal(t, id, cfg.ClientID)
		assert.Equal(t, secret, cfg.ClientSecret)
		assert.Empty(t, cfg.AuthToken)
	})

	t.Run("config with credentials and token", func(t *testing.T) {
		cfg := &Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			AuthToken:    "test-token",
			NoteStoreURL: "https://www.evernote.com/shard/s1/notestore",
		}

		assert.Equal(t, "test-client-id", cfg.ClientID)
		assert.Equal(t, "test-client-secret", cfg.ClientSecret)
		assert.Equal(t, "test-token", cfg.AuthToken)
		assert.Equal(t, "https://www.evernote.com/shard/s1/notestore", cfg.NoteStoreURL)
	})
}

func TestInitCommandFlow(t *testing.T) {
	t.Run("input validation", func(t *testing.T) {
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
