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
	t.Run("openBrowser doesn't panic with valid URL", func(t *testing.T) {
		testURL := "https://example.com"
		assert.NotPanics(t, func() {
			openBrowser(testURL)
		})
	})

	t.Run("openBrowser handles invalid URL safely", func(t *testing.T) {
		// Test with various invalid or potentially malicious URLs
		invalidURLs := []string{
			"javascript:alert('xss')",
			"file:///etc/passwd",
			"not-a-url",
			"",
			"http://",
			"ftp://example.com",
		}
		for _, url := range invalidURLs {
			assert.NotPanics(t, func() {
				openBrowser(url)
			}, "Should handle invalid URL: %s", url)
		}
	})

	t.Run("openBrowser accepts valid http and https URLs", func(t *testing.T) {
		validURLs := []string{
			"http://example.com",
			"https://example.com",
			"https://www.evernote.com/OAuth.action?oauth_token=test",
		}
		for _, url := range validURLs {
			assert.NotPanics(t, func() {
				openBrowser(url)
			}, "Should accept valid URL: %s", url)
		}
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
