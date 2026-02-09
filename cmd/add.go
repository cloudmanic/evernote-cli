package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"strings"

	"github.com/dreampuf/evernote-sdk-golang/edam"
	"github.com/spf13/cobra"
)

var (
	addTitle    string
	addBody     string
	addHTML     string
	addNotebook string
	addTags     []string
	addAttach   []string
)

// wrapENML wraps plain text content in the required Evernote ENML format.
func wrapENML(body string) string {
	escaped := html.EscapeString(body)
	return `<?xml version="1.0" encoding="UTF-8"?>` +
		`<!DOCTYPE en-note SYSTEM "http://xml.evernote.com/pub/enml2.dtd">` +
		`<en-note>` + escaped + `</en-note>`
}

// wrapHTMLInENML wraps raw HTML content in the required Evernote ENML envelope
// without escaping, allowing rich formatting tags to pass through.
func wrapHTMLInENML(body string) string {
	return `<?xml version="1.0" encoding="UTF-8"?>` +
		`<!DOCTYPE en-note SYSTEM "http://xml.evernote.com/pub/enml2.dtd">` +
		`<en-note>` + body + `</en-note>`
}

// addCmd creates a new note in the authenticated Evernote account.
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new note",
	RunE: func(cmd *cobra.Command, args []string) error {
		if addTitle == "" {
			return fmt.Errorf("--title is required")
		}
		if addBody != "" && addHTML != "" {
			return fmt.Errorf("--body and --html cannot be used together")
		}

		ns, token, err := getNoteStoreFunc()
		if err != nil {
			return err
		}

		// Build resources from attached files
		var resources []*edam.Resource
		var mediaTags []string
		for _, filePath := range addAttach {
			res, hash, err := buildResource(filePath)
			if err != nil {
				return err
			}
			resources = append(resources, res)
			mediaTags = append(mediaTags, buildMediaTag(hash, res.GetMime()))
		}

		// Build ENML content with optional media tags for attachments
		var content string
		if addHTML != "" {
			content = wrapHTMLInENML(addHTML)
		} else {
			content = wrapENML(addBody)
		}
		if len(mediaTags) > 0 {
			mediaBlock := strings.Join(mediaTags, "")
			content = strings.Replace(content, "</en-note>", mediaBlock+"</en-note>", 1)
		}

		note := &edam.Note{
			Title:     &addTitle,
			Content:   &content,
			Resources: resources,
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
	addCmd.Flags().StringVar(&addHTML, "html", "", "body of the note as raw HTML (not escaped)")
	addCmd.Flags().StringVar(&addNotebook, "notebook", "", "notebook GUID")
	addCmd.Flags().StringSliceVar(&addTags, "tags", nil, "comma separated list of tag names")
	addCmd.Flags().StringSliceVar(&addAttach, "attach", nil, "file paths to attach to the note")
	rootCmd.AddCommand(addCmd)
}
