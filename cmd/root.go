package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/dghubble/oauth1"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

// jsonFlag is used by subcommands to output JSON.
var jsonFlag bool
var configPath = filepath.Join(os.Getenv("HOME"), ".config", "evernote", "auth.json")

type Config struct {
	ClientID          string        `json:"client_id"`
	ClientSecret      string        `json:"client_secret"`
	Token             *oauth2.Token `json:"token,omitempty"` // Keep for backwards compatibility
	OAuth1Token       string        `json:"oauth1_token,omitempty"`
	OAuth1TokenSecret string        `json:"oauth1_token_secret,omitempty"`
}

func loadConfig() (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

func saveConfig(c *Config) error {
	os.MkdirAll(filepath.Dir(configPath), 0700)
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0600)
}

// rootCmd is the main command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "evernote-cli",
	Short: "A CLI tool to interact with Evernote",
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "output in JSON format")
}

// getOAuth1Config returns an OAuth1 config with the stored credentials
func getOAuth1Config() (*oauth1.Config, *oauth1.Token, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("could not read config: %w", err)
	}

	if cfg.OAuth1Token == "" {
		return nil, nil, fmt.Errorf("no valid OAuth 1.0a token found, run 'evernote-cli auth'")
	}

	config := oauth1.NewConfig(cfg.ClientID, cfg.ClientSecret)
	config.Endpoint = oauth1.Endpoint{
		RequestTokenURL: requestTokenURL,
		AuthorizeURL:    authorizeURL,
		AccessTokenURL:  accessTokenURL,
	}

	token := oauth1.NewToken(cfg.OAuth1Token, cfg.OAuth1TokenSecret)

	return config, token, nil
}

// Legacy function for backwards compatibility
func checkAuth() (string, error) {
	cfg, err := loadConfig()
	if err != nil {
		return "", fmt.Errorf("could not read config: %w", err)
	}

	// Check for OAuth 1.0a token first
	if cfg.OAuth1Token != "" {
		return cfg.OAuth1Token, nil
	}

	// Fall back to OAuth2 token for backwards compatibility
	if cfg.Token != nil && cfg.Token.Valid() {
		return cfg.Token.AccessToken, nil
	}

	return "", fmt.Errorf("no valid token found, run 'evernote-cli auth'")
}

// getOAuth1Client returns an HTTP client configured for OAuth 1.0a signed requests
func getOAuth1Client() (*http.Client, error) {
	config, token, err := getOAuth1Config()
	if err != nil {
		return nil, err
	}
	
	// Create context and return client
	return config.Client(context.Background(), token), nil
}
