// Copyright 2026. All rights reserved.
// Date: 2026-02-06
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd prints the current version of the CLI.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of evernote-cli",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintf(cmd.OutOrStdout(), "evernote-cli version %s\n", Version)
		return nil
	},
}

// init registers the version command with the root command.
func init() {
	rootCmd.AddCommand(versionCmd)
}
