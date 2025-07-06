package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

// jsonFlag is used by subcommands to output JSON.
var jsonFlag bool
var configPath = filepath.Join(os.Getenv("HOME"), ".config", "evernote", "auth.json")

type Config struct {
	ClientID     string        `json:"client_id"`
	ClientSecret string        `json:"client_secret"`
	Token        *oauth2.Token `json:"token,omitempty"`
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

func checkAuth() (string, error) {
	cfg, err := loadConfig()
	if err != nil {
		return "", fmt.Errorf("could not read config: %w", err)
	}
	if cfg.Token == nil || !cfg.Token.Valid() {
		return "", fmt.Errorf("no valid token found, run 'evernote-cli init'")
	}
	return cfg.Token.AccessToken, nil
}
