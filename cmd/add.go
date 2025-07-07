package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

var (
	addTitle    string
	addBody     string
	addNotebook string
	addTags     []string
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new note",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get OAuth 1.0a client
		client, err := getOAuth1Client()
		if err != nil {
			return err
		}

		payload := map[string]interface{}{
			"title":   addTitle,
			"content": addBody,
		}
		if addNotebook != "" {
			payload["notebook"] = addNotebook
		}
		if len(addTags) > 0 {
			payload["tags"] = addTags
		}
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		req, err := http.NewRequest("POST", "https://api.evernote.com/v1/notes", bytes.NewReader(data))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			return fmt.Errorf("unexpected status: %s", resp.Status)
		}

		var respData interface{}
		if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
			return err
		}

		if jsonFlag {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(respData)
		}

		bytes, err := json.MarshalIndent(respData, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", string(bytes))
		return nil
	},
}

func init() {
	addCmd.Flags().StringVar(&addBody, "body", "", "body of the note")
	addCmd.Flags().StringVar(&addTitle, "title", "", "title of the note")
	addCmd.Flags().StringVar(&addNotebook, "notebook", "", "notebook GUID")
	addCmd.Flags().StringSliceVar(&addTags, "tags", nil, "comma separated list of tags")
	rootCmd.AddCommand(addCmd)
}
