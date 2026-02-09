package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"html"

	"github.com/dreampuf/evernote-sdk-golang/edam"
	"github.com/spf13/cobra"
)

var (
	addTitle    string
	addBody     string
	addNotebook string
	addTags     []string
)

// wrapENML wraps plain text content in the required Evernote ENML format.
func wrapENML(body string) string {
	escaped := html.EscapeString(body)
	return `<?xml version="1.0" encoding="UTF-8"?>` +
		`<!DOCTYPE en-note SYSTEM "http://xml.evernote.com/pub/enml2.dtd">` +
		`<en-note>` + escaped + `</en-note>`
}

// addCmd creates a new note in the authenticated Evernote account.
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new note",
	RunE: func(cmd *cobra.Command, args []string) error {
		if addTitle == "" {
			return fmt.Errorf("--title is required")
		}

		ns, token, err := getNoteStoreFunc()
		if err != nil {
			return err
		}

		content := wrapENML(addBody)
		note := &edam.Note{
			Title:   &addTitle,
			Content: &content,
		}

		if addNotebook != "" {
			note.NotebookGuid = &addNotebook
		}
		if len(addTags) > 0 {
			note.TagNames = addTags
		}

		created, err := ns.CreateNote(context.Background(), token, note)
		if err != nil {
			return fmt.Errorf("failed to create note: %w", formatAPIError(err))
		}

		if jsonFlag {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(created)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Note created: %s\n", created.GetTitle())
		fmt.Fprintf(cmd.OutOrStdout(), "GUID: %s\n", created.GetGUID())
		return nil
	},
}

func init() {
	addCmd.Flags().StringVar(&addTitle, "title", "", "title of the note (required)")
	addCmd.Flags().StringVar(&addBody, "body", "", "body of the note")
	addCmd.Flags().StringVar(&addNotebook, "notebook", "", "notebook GUID")
	addCmd.Flags().StringSliceVar(&addTags, "tags", nil, "comma separated list of tag names")
	rootCmd.AddCommand(addCmd)
}
