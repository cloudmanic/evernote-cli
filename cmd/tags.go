package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// tagsCmd lists all tags in the authenticated Evernote account.
var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "List all tags",
	RunE: func(cmd *cobra.Command, args []string) error {
		ns, token, err := getNoteStoreFunc()
		if err != nil {
			return err
		}

		tags, err := ns.ListTags(context.Background(), token)
		if err != nil {
			return fmt.Errorf("failed to list tags: %w", formatAPIError(err))
		}

		if jsonFlag {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(tags)
		}

		if len(tags) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No tags found.")
			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Found %d tag(s):\n\n", len(tags))
		for i, tag := range tags {
			fmt.Fprintf(cmd.OutOrStdout(), "%d. %s", i+1, tag.GetName())
			fmt.Fprintln(cmd.OutOrStdout())
			if tag.GetGUID() != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "   GUID: %s\n", tag.GetGUID())
			}
			fmt.Fprintln(cmd.OutOrStdout())
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(tagsCmd)
}
