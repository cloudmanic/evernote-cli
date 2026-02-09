// Copyright 2026. All rights reserved.
// Date: 2026-02-08
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dreampuf/evernote-sdk-golang/edam"
	"github.com/spf13/cobra"
)

var (
	updateTitle  string
	updateBody   string
	updateAppend string
	updateTags   []string
)

// updateCmd updates an existing note by its GUID.
var updateCmd = &cobra.Command{
	Use:   "update [guid]",
	Short: "Update an existing note",
	Long: `Update an existing note by GUID. You can change the title, replace the body,
or append text to the existing content.

Examples:
  evernote-cli update <guid> --title "New Title"
  evernote-cli update <guid> --body "Replace body with this"
  evernote-cli update <guid> --append "Add this to the end"
  evernote-cli update <guid> --title "New Title" --append "And add this"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if updateTitle == "" && updateBody == "" && updateAppend == "" && len(updateTags) == 0 {
			return fmt.Errorf("at least one of --title, --body, --append, or --tags is required")
		}
		if updateBody != "" && updateAppend != "" {
			return fmt.Errorf("--body and --append cannot be used together")
		}

		ns, token, err := getNoteStoreFunc()
		if err != nil {
			return err
		}

		guid := edam.GUID(args[0])

		// Fetch the existing note to get its current title and content
		existing, err := ns.GetNote(context.Background(), token, guid, true, false, false, false)
		if err != nil {
			return fmt.Errorf("failed to get note: %w", err)
		}

		// Build the note update - GUID and title are always required by the API
		note := &edam.Note{
			GUID: &guid,
		}

		// Set the title (use existing if not changing)
		if updateTitle != "" {
			note.Title = &updateTitle
		} else {
			title := existing.GetTitle()
			note.Title = &title
		}

		// Handle content changes
		if updateBody != "" {
			// Replace body entirely
			content := wrapENML(updateBody)
			note.Content = &content
		} else if updateAppend != "" {
			// Append to existing content
			existingContent := existing.GetContent()
			plainText := stripENML(existingContent)

			// Combine existing text with appended text
			var combined string
			if strings.TrimSpace(plainText) != "" {
				combined = plainText + "\n\n" + updateAppend
			} else {
				combined = updateAppend
			}
			content := wrapENML(combined)
			note.Content = &content
		}

		// Handle tag changes
		if len(updateTags) > 0 {
			note.TagNames = updateTags
		}

		// Send the update to Evernote
		updated, err := ns.UpdateNote(context.Background(), token, note)
		if err != nil {
			return fmt.Errorf("failed to update note: %w", err)
		}

		if jsonFlag {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(updated)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Note updated: %s\n", updated.GetTitle())
		fmt.Fprintf(cmd.OutOrStdout(), "GUID: %s\n", updated.GetGUID())
		return nil
	},
}

func init() {
	updateCmd.Flags().StringVar(&updateTitle, "title", "", "new title for the note")
	updateCmd.Flags().StringVar(&updateBody, "body", "", "replace the note body entirely")
	updateCmd.Flags().StringVar(&updateAppend, "append", "", "append text to the existing note content")
	updateCmd.Flags().StringSliceVar(&updateTags, "tags", nil, "comma separated list of tag names")
	rootCmd.AddCommand(updateCmd)
}
