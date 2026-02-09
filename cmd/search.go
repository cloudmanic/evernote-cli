package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dreampuf/evernote-sdk-golang/edam"
	"github.com/spf13/cobra"
)

// searchCmd searches notes by a query string using the Evernote search grammar.
var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search notes",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ns, token, err := getNoteStoreFunc()
		if err != nil {
			return err
		}

		query := strings.Join(args, " ")
		includeTitle := true
		includeCreated := true
		includeUpdated := true
		includeNotebookGuid := true

		filter := &edam.NoteFilter{
			Words: &query,
		}
		resultSpec := &edam.NotesMetadataResultSpec{
			IncludeTitle:        &includeTitle,
			IncludeCreated:      &includeCreated,
			IncludeUpdated:      &includeUpdated,
			IncludeNotebookGuid: &includeNotebookGuid,
		}

		results, err := ns.FindNotesMetadata(context.Background(), token, filter, 0, 100, resultSpec)
		if err != nil {
			return fmt.Errorf("failed to search notes: %w", formatAPIError(err))
		}

		if jsonFlag {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(results)
		}

		notes := results.GetNotes()
		if len(notes) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No notes found.")
			return nil
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Found %d note(s):\n\n", results.GetTotalNotes())
		for i, note := range notes {
			fmt.Fprintf(cmd.OutOrStdout(), "%d. %s\n", i+1, note.GetTitle())
			fmt.Fprintf(cmd.OutOrStdout(), "   GUID: %s\n", note.GetGUID())
			if note.GetCreated() != 0 {
				created := time.Unix(int64(note.GetCreated())/1000, 0)
				fmt.Fprintf(cmd.OutOrStdout(), "   Created: %s\n", created.Format("2006-01-02 15:04:05"))
			}
			if note.GetUpdated() != 0 {
				updated := time.Unix(int64(note.GetUpdated())/1000, 0)
				fmt.Fprintf(cmd.OutOrStdout(), "   Updated: %s\n", updated.Format("2006-01-02 15:04:05"))
			}
			fmt.Fprintln(cmd.OutOrStdout())
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
