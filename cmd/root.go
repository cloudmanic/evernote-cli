package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// jsonFlag is used by subcommands to output JSON.
var jsonFlag bool

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
	configPath := os.Getenv("HOME") + "/.config/evernote/auth.json"
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("could not read auth token: %w", err)
	}
	return string(data), nil
}
