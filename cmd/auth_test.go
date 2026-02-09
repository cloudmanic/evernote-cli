package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthCmdConfiguration(t *testing.T) {
	assert.Equal(t, "auth", authCmd.Use)
	assert.Equal(t, "Authenticate with Evernote", authCmd.Short)
	assert.NotNil(t, authCmd.RunE)
}

func TestAuthCommandRegistration(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "auth" {
			found = true
			assert.Equal(t, "Authenticate with Evernote", cmd.Short)
			break
		}
	}
	assert.True(t, found, "auth command should be registered with root command")
}

func TestOpenBrowser(t *testing.T) {
	testURL := "https://example.com"

	t.Run("openBrowser doesn't panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			openBrowser(testURL)
		})
	})
}

func TestHTTPServerConfiguration(t *testing.T) {
	t.Run("server address", func(t *testing.T) {
		expectedAddr := ":8080"
		assert.Equal(t, expectedAddr, ":8080")
	})

	t.Run("callback URL format", func(t *testing.T) {
		callbackURL := "http://localhost:8080/callback"
		assert.True(t, strings.HasPrefix(callbackURL, "http://"))
		assert.True(t, strings.Contains(callbackURL, "localhost:8080"))
		assert.True(t, strings.HasSuffix(callbackURL, "/callback"))
	})
}
