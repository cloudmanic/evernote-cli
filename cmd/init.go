package cmd

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize credentials and authenticate",
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(cmd.InOrStdin())
		fmt.Fprint(cmd.OutOrStdout(), "Evernote Client ID: ")
		id, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		fmt.Fprint(cmd.OutOrStdout(), "Evernote Client Secret: ")
		secret, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		id = strings.TrimSpace(id)
		secret = strings.TrimSpace(secret)

		cfg := &Config{ClientID: id, ClientSecret: secret}
		token, err := runAuthFlow(id, secret)
		if err != nil {
			return err
		}
		cfg.Token = token
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
