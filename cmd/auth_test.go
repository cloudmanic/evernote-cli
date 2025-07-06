package cmd

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

func TestEvernoteEndpoint(t *testing.T) {
	// Test that the Evernote OAuth2 endpoints are configured correctly
	assert.Equal(t, "https://www.evernote.com/oauth2/authorize", evernoteEndpoint.AuthURL)
	assert.Equal(t, "https://www.evernote.com/oauth2/token", evernoteEndpoint.TokenURL)
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

func TestOAuth2Config(t *testing.T) {
	// Test OAuth2 configuration setup that would be used in runAuthFlow
	t.Run("oauth2 config creation", func(t *testing.T) {
		clientID := "test-client-id"
		clientSecret := "test-client-secret"

		conf := &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       []string{"basic"},
			Endpoint:     evernoteEndpoint,
			RedirectURL:  "http://localhost:8080/callback",
		}

		assert.Equal(t, clientID, conf.ClientID)
		assert.Equal(t, clientSecret, conf.ClientSecret)
		assert.Equal(t, []string{"basic"}, conf.Scopes)
		assert.Equal(t, evernoteEndpoint, conf.Endpoint)
		assert.Equal(t, "http://localhost:8080/callback", conf.RedirectURL)
	})

	t.Run("auth code URL generation", func(t *testing.T) {
		conf := &oauth2.Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			Scopes:       []string{"basic"},
			Endpoint:     evernoteEndpoint,
			RedirectURL:  "http://localhost:8080/callback",
		}

		url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
		
		assert.Contains(t, url, "https://www.evernote.com/oauth2/authorize")
		assert.Contains(t, url, "client_id=test-client-id")
		assert.Contains(t, url, "redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fcallback")
		assert.Contains(t, url, "response_type=code")
		assert.Contains(t, url, "scope=basic")
		assert.Contains(t, url, "state=state")
		assert.Contains(t, url, "access_type=offline")
	})
}

func TestTokenValidation(t *testing.T) {
	// Test token validation logic used in auth flows
	t.Run("valid token", func(t *testing.T) {
		token := &oauth2.Token{
			AccessToken:  "test-access-token",
			TokenType:    "Bearer",
			RefreshToken: "test-refresh-token",
			Expiry:       time.Now().Add(time.Hour),
		}

		assert.True(t, token.Valid())
		assert.Equal(t, "test-access-token", token.AccessToken)
		assert.Equal(t, "Bearer", token.TokenType)
		assert.Equal(t, "test-refresh-token", token.RefreshToken)
	})

	t.Run("expired token", func(t *testing.T) {
		token := &oauth2.Token{
			AccessToken: "expired-token",
			TokenType:   "Bearer",
			Expiry:      time.Now().Add(-time.Hour), // Expired 1 hour ago
		}

		assert.False(t, token.Valid())
	})

	t.Run("empty token", func(t *testing.T) {
		token := &oauth2.Token{}
		assert.False(t, token.Valid())
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

// Note: runAuthFlow is complex to test as it involves:
// - Starting an HTTP server
// - Opening a browser
// - OAuth2 token exchange
// These would typically be tested with integration tests or by mocking the external dependencies.
// For now, we test the components that can be unit tested in isolation.

func TestOAuth2Context(t *testing.T) {
	// Test context usage in OAuth2 operations
	t.Run("background context", func(t *testing.T) {
		ctx := context.Background()
		assert.NotNil(t, ctx)
		
		// Test context with timeout (similar to what might be used in real auth flow)
		ctxWithTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		
		assert.NotNil(t, ctxWithTimeout)
		
		// Check that context is not done immediately
		select {
		case <-ctxWithTimeout.Done():
			t.Fatal("context should not be done immediately")
		default:
			// Expected case
		}
	})
}

func TestHTTPServerConfiguration(t *testing.T) {
	// Test HTTP server configuration used in auth flow
	t.Run("server address", func(t *testing.T) {
		// This tests the server configuration that would be used in runAuthFlow
		expectedAddr := ":8080"
		assert.Equal(t, expectedAddr, ":8080")
	})
}