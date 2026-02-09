// Copyright 2026. All rights reserved.
// Date: 2026-02-08
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dreampuf/evernote-sdk-golang/edam"
	"github.com/spf13/cobra"
)

// stripENML removes ENML/XML tags and returns plain text content.
func stripENML(content string) string {
	// Remove XML declaration and DOCTYPE
	content = regexp.MustCompile(`<\?xml[^?]*\?>`).ReplaceAllString(content, "")
	content = regexp.MustCompile(`<!DOCTYPE[^>]*>`).ReplaceAllString(content, "")

	// Replace <br/> and <div> tags with newlines
	content = regexp.MustCompile(`<br\s*/?>|<div>|</div>`).ReplaceAllString(content, "\n")

	// Remove all remaining HTML/ENML tags
	content = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(content, "")

	// Clean up whitespace
	content = strings.TrimSpace(content)

	return content
}

// getCmd retrieves a single note by its GUID.
var getCmd = &cobra.Command{
	Use:   "get [guid]",
	Short: "Get a note by GUID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ns, token, err := getNoteStoreFunc()
		if err != nil {
			return err
		}

		guid := edam.GUID(args[0])
		note, err := ns.GetNote(context.Background(), token, guid, true, false, false, false)
		if err != nil {
			return fmt.Errorf("failed to get note: %w", formatAPIError(err))
		}

		if jsonFlag {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(note)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Title: %s\n", note.GetTitle())
		fmt.Fprintf(cmd.OutOrStdout(), "GUID:  %s\n", note.GetGUID())

		if note.GetNotebookGuid() != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "Notebook: %s\n", note.GetNotebookGuid())
		}

		if len(note.GetTagNames()) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "Tags: %s\n", strings.Join(note.GetTagNames(), ", "))
		} else if len(note.GetTagGuids()) > 0 {
			tagGUIDs := make([]string, len(note.GetTagGuids()))
			for i, g := range note.GetTagGuids() {
				tagGUIDs[i] = string(g)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Tag GUIDs: %s\n", strings.Join(tagGUIDs, ", "))
		}

		if note.GetCreated() != 0 {
			created := time.Unix(int64(note.GetCreated())/1000, 0)
			fmt.Fprintf(cmd.OutOrStdout(), "Created: %s\n", created.Format("2006-01-02 15:04:05"))
		}
		if note.GetUpdated() != 0 {
			updated := time.Unix(int64(note.GetUpdated())/1000, 0)
			fmt.Fprintf(cmd.OutOrStdout(), "Updated: %s\n", updated.Format("2006-01-02 15:04:05"))
		}

		if len(note.GetResources()) > 0 {
			fmt.Fprintf(cmd.OutOrStdout(), "\nAttachments (%d):\n", len(note.GetResources()))
			for i, res := range note.GetResources() {
				fileName := "unnamed"
				if res.GetAttributes() != nil && res.GetAttributes().GetFileName() != "" {
					fileName = res.GetAttributes().GetFileName()
				}
				fmt.Fprintf(cmd.OutOrStdout(), "  %d. %s (GUID: %s, %s)\n", i+1, fileName, res.GetGUID(), res.GetMime())
			}
			fmt.Fprintf(cmd.OutOrStdout(), "\nUse: evernote-cli download <resource-guid>\n")
		}

		if note.GetContent() != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "\n%s\n", stripENML(note.GetContent()))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
