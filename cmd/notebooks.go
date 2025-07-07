package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

var notebooksCmd = &cobra.Command{
	Use:   "notebooks",
	Short: "List all notebooks",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get OAuth 1.0a client
		client, err := getOAuth1Client()
		if err != nil {
			return err
		}

		req, err := http.NewRequest("GET", "https://api.evernote.com/v1/notebooks", nil)
		if err != nil {
			return err
		}

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status: %s", resp.Status)
		}

		var data interface{}
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
			return err
		}

		if jsonFlag {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(data)
		}

		// Format notebooks in a nice way
		return formatNotebooks(cmd, data)
	},
}

// formatNotebooks formats the notebooks data in a user-friendly way
func formatNotebooks(cmd *cobra.Command, data interface{}) error {
	// Try to parse the data as a notebooks response
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		// If we can't parse it properly, fall back to JSON formatting
		bytes, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", string(bytes))
		return nil
	}

	// Look for notebooks array in the response
	notebooks, ok := dataMap["notebooks"].([]interface{})
	if !ok {
		// If no notebooks array found, fall back to JSON formatting
		bytes, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", string(bytes))
		return nil
	}

	if len(notebooks) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No notebooks found.")
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Found %d notebook(s):\n\n", len(notebooks))

	for i, notebook := range notebooks {
		notebookMap, ok := notebook.(map[string]interface{})
		if !ok {
			continue
		}

		name, _ := notebookMap["name"].(string)
		guid, _ := notebookMap["guid"].(string)
		defaultNotebook, _ := notebookMap["defaultNotebook"].(bool)

		fmt.Fprintf(cmd.OutOrStdout(), "%d. %s", i+1, name)
		if defaultNotebook {
			fmt.Fprint(cmd.OutOrStdout(), " (default)")
		}
		fmt.Fprintln(cmd.OutOrStdout())
		if guid != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "   GUID: %s\n", guid)
		}
		fmt.Fprintln(cmd.OutOrStdout())
	}

	return nil
}

func init() {
	rootCmd.AddCommand(notebooksCmd)
}
