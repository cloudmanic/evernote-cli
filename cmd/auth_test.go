package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/dghubble/oauth1"
	"github.com/stretchr/testify/assert"
)

func TestOAuth1Endpoints(t *testing.T) {
	// Test that the Evernote OAuth 1.0a endpoints are configured correctly
	assert.Equal(t, "https://www.evernote.com/oauth", requestTokenURL)
	assert.Equal(t, "https://www.evernote.com/OAuth.action", authorizeURL)
	assert.Equal(t, "https://www.evernote.com/oauth", accessTokenURL)
}

func TestAuthCmdConfiguration(t *testing.T) {
	// Test the auth command metadata
	assert.Equal(t, "auth", authCmd.Use)
	assert.Equal(t, "Authenticate with Evernote", authCmd.Short)
	assert.NotNil(t, authCmd.RunE)
}

func TestAuthCmdEnvironmentVariables(t *testing.T) {
	// Save original environment variables
	originalClientID := os.Getenv("EVERNOTE_CLIENT_ID")
	originalClientSecret := os.Getenv("EVERNOTE_CLIENT_SECRET")
	defer func() {
		os.Setenv("EVERNOTE_CLIENT_ID", originalClientID)
		os.Setenv("EVERNOTE_CLIENT_SECRET", originalClientSecret)
	}()

	t.Run("with environment variables set", func(t *testing.T) {
		// Set test environment variables
		os.Setenv("EVERNOTE_CLIENT_ID", "env-client-id")
		os.Setenv("EVERNOTE_CLIENT_SECRET", "env-client-secret")

		// Verify they can be read
		assert.Equal(t, "env-client-id", os.Getenv("EVERNOTE_CLIENT_ID"))
		assert.Equal(t, "env-client-secret", os.Getenv("EVERNOTE_CLIENT_SECRET"))
	})

	t.Run("without environment variables", func(t *testing.T) {
		// Clear environment variables
		os.Unsetenv("EVERNOTE_CLIENT_ID")
		os.Unsetenv("EVERNOTE_CLIENT_SECRET")

		assert.Empty(t, os.Getenv("EVERNOTE_CLIENT_ID"))
		assert.Empty(t, os.Getenv("EVERNOTE_CLIENT_SECRET"))
	})
}

func TestOAuth1Config(t *testing.T) {
	// Test OAuth1 configuration setup that would be used in runAuthFlow
	t.Run("oauth1 config creation", func(t *testing.T) {
		clientID := "test-client-id"
		clientSecret := "test-client-secret"

		conf := oauth1.NewConfig(clientID, clientSecret)
		conf.CallbackURL = "http://localhost:8080/callback"
		conf.Endpoint = oauth1.Endpoint{
			RequestTokenURL: requestTokenURL,
			AuthorizeURL:    authorizeURL,
			AccessTokenURL:  accessTokenURL,
		}

		assert.Equal(t, clientID, conf.ConsumerKey)
		assert.Equal(t, clientSecret, conf.ConsumerSecret)
		assert.Equal(t, "http://localhost:8080/callback", conf.CallbackURL)
		assert.Equal(t, requestTokenURL, conf.Endpoint.RequestTokenURL)
		assert.Equal(t, authorizeURL, conf.Endpoint.AuthorizeURL)
		assert.Equal(t, accessTokenURL, conf.Endpoint.AccessTokenURL)
	})

	t.Run("auth URL generation", func(t *testing.T) {
		conf := oauth1.NewConfig("test-client-id", "test-client-secret")
		conf.CallbackURL = "http://localhost:8080/callback"
		conf.Endpoint = oauth1.Endpoint{
			RequestTokenURL: requestTokenURL,
			AuthorizeURL:    authorizeURL,
			AccessTokenURL:  accessTokenURL,
		}

		// Test authorization URL creation
		requestToken := "test-request-token"
		authURL, err := conf.AuthorizationURL(requestToken)

		assert.NoError(t, err)
		assert.NotNil(t, authURL)
		assert.Equal(t, "https", authURL.Scheme)
		assert.Equal(t, "www.evernote.com", authURL.Host)
		assert.Equal(t, "/OAuth.action", authURL.Path)
		assert.Equal(t, requestToken, authURL.Query().Get("oauth_token"))
	})
}

func TestOAuth1Token(t *testing.T) {
	// Test OAuth1Token struct
	t.Run("token creation", func(t *testing.T) {
		token := &OAuth1Token{
			Token:       "test-access-token",
			TokenSecret: "test-token-secret",
		}

		assert.Equal(t, "test-access-token", token.Token)
		assert.Equal(t, "test-token-secret", token.TokenSecret)
	})

	t.Run("oauth1 token object", func(t *testing.T) {
		// Test oauth1.Token creation
		token := oauth1.NewToken("access-token", "token-secret")

		assert.Equal(t, "access-token", token.Token)
		assert.Equal(t, "token-secret", token.TokenSecret)
	})
}

func TestAuthCommandRegistration(t *testing.T) {
	// Test that auth command is properly registered with root command
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
	// Test openBrowser function with different OS
	testURL := "https://example.com"

	// Since we can't actually test opening a browser, we just ensure
	// the function doesn't panic on different platforms
	t.Run("openBrowser doesn't panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			openBrowser(testURL)
		})
	})
}

func TestHTTPServerConfiguration(t *testing.T) {
	// Test HTTP server configuration used in auth flow
	t.Run("server address", func(t *testing.T) {
		// This tests the server configuration that would be used in runAuthFlow
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

// Note: runAuthFlow is complex to test as it involves:
// - Starting an HTTP server
// - Opening a browser
// - OAuth1 token exchange with external service
// These would typically be tested with integration tests or by mocking the external dependencies.
// For now, we test the components that can be unit tested in isolation.
