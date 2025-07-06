package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
)

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "List all tags",
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := checkAuth()
		if err != nil {
			return err
		}

		req, err := http.NewRequest("GET", "https://api.evernote.com/v1/tags", nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
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

		// Format tags in a nice way
		return formatTags(cmd, data)
	},
}

// formatTags formats the tags data in a user-friendly way
func formatTags(cmd *cobra.Command, data interface{}) error {
	// Try to parse the data as a tags response
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

	// Look for tags array in the response
	tags, ok := dataMap["tags"].([]interface{})
	if !ok {
		// If no tags array found, fall back to JSON formatting
		bytes, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "%s\n", string(bytes))
		return nil
	}

	if len(tags) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No tags found.")
		return nil
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Found %d tag(s):\n\n", len(tags))

	for i, tag := range tags {
		tagMap, ok := tag.(map[string]interface{})
		if !ok {
			continue
		}

		name, _ := tagMap["name"].(string)
		guid, _ := tagMap["guid"].(string)

		fmt.Fprintf(cmd.OutOrStdout(), "%d. %s", i+1, name)
		fmt.Fprintln(cmd.OutOrStdout())
		if guid != "" {
			fmt.Fprintf(cmd.OutOrStdout(), "   GUID: %s\n", guid)
		}
		fmt.Fprintln(cmd.OutOrStdout())
	}

	return nil
}

func init() {
	rootCmd.AddCommand(tagsCmd)
}