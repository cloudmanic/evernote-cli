package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/dreampuf/evernote-sdk-golang/client"
	"github.com/spf13/cobra"
)

// runAuthFlow performs the OAuth 1.0a flow using the Evernote SDK and returns
// the auth token and NoteStore URL.
func runAuthFlow(clientID, clientSecret string) (string, string, error) {
	c := client.NewClient(clientID, clientSecret, client.PRODUCTION)

	// Get request token and authorization URL
	requestToken, authURL, err := c.GetRequestToken("http://localhost:8080/callback")
	if err != nil {
		return "", "", fmt.Errorf("failed to get request token: %w", err)
	}

	// Set up local server for callback using a custom mux to avoid conflicts
	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	verifierCh := make(chan string, 1)

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		verifier := r.URL.Query().Get("oauth_verifier")
		if verifier == "" {
			fmt.Fprintf(w, "Authentication failed. Missing verifier.")
			return
		}
		fmt.Fprintf(w, "Authentication complete. You can close this window.")
		verifierCh <- verifier
	})

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	// Open browser to authorization URL
	openBrowser(authURL)
	fmt.Printf("If the browser did not open, visit: %s\n", authURL)

	// Wait for callback with timeout
	var verifier string
	select {
	case verifier = <-verifierCh:
	case <-time.After(5 * time.Minute):
		srv.Shutdown(context.Background())
		return "", "", fmt.Errorf("authentication timeout")
	}

	srv.Shutdown(context.Background())

	// Exchange verifier for access token
	accessToken, err := c.GetAuthorizedToken(requestToken, verifier)
	if err != nil {
		return "", "", fmt.Errorf("failed to get access token: %w", err)
	}

	// Extract the NoteStore URL from the access token response
	noteStoreURL := accessToken.AdditionalData["edam_noteStoreUrl"]

	return accessToken.Token, noteStoreURL, nil
}

// openBrowser opens the given URL in the user's default browser.
// It validates the URL to prevent command injection attacks.
func openBrowser(urlStr string) {
	// Validate URL to prevent command injection
	parsedURL, err := url.Parse(urlStr)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		// Invalid URL, don't attempt to open
		return
	}
	
	// Only allow http and https schemes
	if !strings.EqualFold(parsedURL.Scheme, "http") && !strings.EqualFold(parsedURL.Scheme, "https") {
		return
	}

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", urlStr).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", urlStr).Start()
	case "darwin":
		err = exec.Command("open", urlStr).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		// Browser couldn't be opened, user will need to copy/paste URL
	}
}

// authCmd authenticates with Evernote using OAuth 1.0a.
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with Evernote",
	RunE: func(cmd *cobra.Command, args []string) error {
		clientID := os.Getenv("EVERNOTE_CLIENT_ID")
		clientSecret := os.Getenv("EVERNOTE_CLIENT_SECRET")
		cfg, _ := loadConfig()
		if clientID == "" && cfg != nil {
			clientID = cfg.ClientID
		}
		if clientSecret == "" && cfg != nil {
			clientSecret = cfg.ClientSecret
		}
		if clientID == "" || clientSecret == "" {
			return fmt.Errorf("client ID and secret must be provided (run 'evernote-cli init')")
		}

		token, noteStoreURL, err := runAuthFlow(clientID, clientSecret)
		if err != nil {
			return err
		}

		if cfg == nil {
			cfg = &Config{}
		}
		cfg.ClientID = clientID
		cfg.ClientSecret = clientSecret
		cfg.AuthToken = token
		cfg.NoteStoreURL = noteStoreURL

		if err := saveConfig(cfg); err != nil {
			return err
		}
		fmt.Println("Authentication successful. Token saved.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
}
