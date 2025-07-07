package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search notes",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get OAuth 1.0a client
		client, err := getOAuth1Client()
		if err != nil {
			return err
		}
		
		q := url.QueryEscape(args[0])
		req, err := http.NewRequest("GET", "https://api.evernote.com/v1/search?query="+q, nil)
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

		// naive formatting
		bytes, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", string(bytes))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
