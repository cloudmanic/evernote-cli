package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// notebooksCmd lists all notebooks in the authenticated Evernote account.
var notebooksCmd = &cobra.Command{
	Use:   "notebooks",
	Short: "List all notebooks",
	RunE: func(cmd *cobra.Command, args []string) error {
		ns, token, err := getNoteStoreFunc()
		if err != nil {
			return err
		}

		notebooks, err := ns.ListNotebooks(context.Background(), token)
		if err != nil {
			return fmt.Errorf("failed to list notebooks: %w", formatAPIError(err))
		}

		if jsonFlag {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(notebooks)
		}

		if len(notebooks) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No notebooks found.")
			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Found %d notebook(s):\n\n", len(notebooks))
		for i, nb := range notebooks {
			fmt.Fprintf(cmd.OutOrStdout(), "%d. %s", i+1, nb.GetName())
			if nb.GetDefaultNotebook() {
				fmt.Fprint(cmd.OutOrStdout(), " (default)")
			}
			fmt.Fprintln(cmd.OutOrStdout())
			if nb.GetGUID() != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "   GUID: %s\n", nb.GetGUID())
			}
			fmt.Fprintln(cmd.OutOrStdout())
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(notebooksCmd)
}
