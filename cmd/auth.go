package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/dghubble/oauth1"
	"github.com/spf13/cobra"
)

// Evernote's OAuth 1.0a endpoints
const (
	requestTokenURL = "https://www.evernote.com/oauth"
	authorizeURL    = "https://www.evernote.com/OAuth.action"
	accessTokenURL  = "https://www.evernote.com/oauth"
)

// OAuth1Token represents the OAuth 1.0a token
type OAuth1Token struct {
	Token       string `json:"token"`
	TokenSecret string `json:"token_secret"`
}

// Custom HTTP client for debugging
type loggingTransport struct {
	transport http.RoundTripper
}

func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	fmt.Printf("Debug: Request: %s %s\n", req.Method, req.URL)
	fmt.Printf("Debug: Headers: %v\n", req.Header)
	
	resp, err := t.transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Debug: Response Status: %s\n", resp.Status)
	fmt.Printf("Debug: Response Body: %s\n", string(body))
	
	// Replace the body so it can be read again
	resp.Body = io.NopCloser(bytes.NewBuffer(body))
	
	return resp, nil
}

func createLoggingClient() *http.Client {
	return &http.Client{
		Transport: &loggingTransport{http.DefaultTransport},
	}
}

// Custom access token method for Evernote's OAuth implementation
func evernoteAccessToken(config *oauth1.Config, requestToken, requestSecret, verifier string) (*OAuth1Token, error) {
	// Create OAuth1 client with request token
	requestTokenObj := oauth1.NewToken(requestToken, requestSecret)
	client := config.Client(context.Background(), requestTokenObj)
	
	// Make the access token request manually
	values := url.Values{}
	values.Set("oauth_verifier", verifier)
	
	req, err := http.NewRequest("POST", accessTokenURL, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	// Add OAuth signature manually using the oauth1 transport
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	fmt.Printf("Debug: Manual access token response: %s\n", string(body))
	
	// Parse the response
	responseValues, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, err
	}
	
	accessToken := responseValues.Get("oauth_token")
	if accessToken == "" {
		return nil, fmt.Errorf("response missing oauth_token: %s", string(body))
	}
	
	// For Evernote, the token secret might be empty, so we'll use empty string
	accessSecret := responseValues.Get("oauth_token_secret")
	
	return &OAuth1Token{
		Token:       accessToken,
		TokenSecret: accessSecret, // This might be empty for Evernote
	}, nil
}

// runAuthFlow performs the OAuth 1.0a flow and returns the resulting token.
func runAuthFlow(clientID, clientSecret string) (*OAuth1Token, error) {
	config := oauth1.NewConfig(clientID, clientSecret)
	config.CallbackURL = "http://localhost:8080/callback"
	config.Endpoint = oauth1.Endpoint{
		RequestTokenURL: requestTokenURL,
		AuthorizeURL:    authorizeURL,
		AccessTokenURL:  accessTokenURL,
	}
	
	// Use custom HTTP client for debugging
	config.HTTPClient = createLoggingClient()

	// Get request token
	requestToken, requestSecret, err := config.RequestToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get request token: %w", err)
	}

	// Set up local server for callback
	srv := &http.Server{Addr: ":8080"}
	verifierCh := make(chan string)
	tokenCh := make(chan string)

	http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		verifier := r.URL.Query().Get("oauth_verifier")
		token := r.URL.Query().Get("oauth_token")

		if verifier == "" || token == "" {
			fmt.Fprintf(w, "Authentication failed. Missing verifier or token.")
			return
		}

		fmt.Fprintf(w, "Authentication complete. You can close this window.")
		verifierCh <- verifier
		tokenCh <- token
	})

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	// Build authorization URL
	authURL, err := config.AuthorizationURL(requestToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create authorization URL: %w", err)
	}

	// Open browser
	openBrowser(authURL.String())
	fmt.Printf("If the browser did not open, visit: %s\n", authURL.String())

	// Wait for callback
	var verifier, callbackToken string
	select {
	case verifier = <-verifierCh:
		callbackToken = <-tokenCh
	case <-time.After(5 * time.Minute):
		srv.Shutdown(context.Background())
		return nil, fmt.Errorf("authentication timeout")
	}

	// Shutdown server
	srv.Shutdown(context.Background())

	// Verify the callback token matches
	if callbackToken != requestToken {
		return nil, fmt.Errorf("oauth token mismatch")
	}

	// Exchange for access token
	fmt.Printf("Debug: Request token: %s\n", requestToken)
	fmt.Printf("Debug: Request secret: %s\n", requestSecret)
	fmt.Printf("Debug: Verifier: %s\n", verifier)
	fmt.Printf("Debug: Callback token: %s\n", callbackToken)
	
	// Use custom Evernote access token method
	token, err := evernoteAccessToken(config, requestToken, requestSecret, verifier)
	if err != nil {
		fmt.Printf("Debug: Access token exchange failed: %v\n", err)
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	return token, nil
}

func openBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		// Browser couldn't be opened, user will need to copy/paste URL
	}
}

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

		token, err := runAuthFlow(clientID, clientSecret)
		if err != nil {
			return err
		}

		if cfg == nil {
			cfg = &Config{}
		}
		cfg.ClientID = clientID
		cfg.ClientSecret = clientSecret
		cfg.OAuth1Token = token.Token
		cfg.OAuth1TokenSecret = token.TokenSecret

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
