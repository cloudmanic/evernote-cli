package cmd

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// initCmd prompts for API credentials and runs the OAuth flow.
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize credentials and authenticate",
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(cmd.InOrStdin())
		fmt.Fprint(cmd.OutOrStdout(), "Evernote Consumer Key: ")
		id, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		fmt.Fprint(cmd.OutOrStdout(), "Evernote Consumer Secret: ")
		secret, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		id = strings.TrimSpace(id)
		secret = strings.TrimSpace(secret)

		token, noteStoreURL, err := runAuthFlow(id, secret)
		if err != nil {
			return err
		}

		cfg := &Config{
			ClientID:     id,
			ClientSecret: secret,
			AuthToken:    token,
			NoteStoreURL: noteStoreURL,
		}
		if err := saveConfig(cfg); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Authentication successful. Configuration saved to", configPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
