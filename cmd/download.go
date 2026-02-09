// Copyright 2026. All rights reserved.
// Date: 2026-02-08
package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dreampuf/evernote-sdk-golang/edam"
	"github.com/spf13/cobra"
)

var downloadOutput string

// downloadCmd downloads a resource (attachment) by its GUID.
var downloadCmd = &cobra.Command{
	Use:   "download [resource-guid]",
	Short: "Download a note attachment by resource GUID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ns, token, err := getNoteStoreFunc()
		if err != nil {
			return err
		}

		guid := edam.GUID(args[0])

		// Get resource metadata to determine filename
		resource, err := ns.GetResource(context.Background(), token, guid, true, false, true, false)
		if err != nil {
			return fmt.Errorf("failed to get resource: %w", formatAPIError(err))
		}

		// Determine output filename
		outputPath := downloadOutput
		if outputPath == "" {
			if resource.GetAttributes() != nil && resource.GetAttributes().GetFileName() != "" {
				// Sanitize filename to prevent path traversal attacks
				outputPath = filepath.Base(resource.GetAttributes().GetFileName())
			} else {
				outputPath = string(guid)
			}
		}

		// Get the file data from the resource
		if resource.GetData() == nil || len(resource.GetData().GetBody()) == 0 {
			return fmt.Errorf("resource has no data")
		}

		data := resource.GetData().GetBody()

		// Ensure output directory exists
		dir := filepath.Dir(outputPath)
		if dir != "." && dir != "" {
			os.MkdirAll(dir, 0755)
		}

		if err := os.WriteFile(outputPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Downloaded: %s (%d bytes, %s)\n", outputPath, len(data), resource.GetMime())
		return nil
	},
}

func init() {
	downloadCmd.Flags().StringVarP(&downloadOutput, "output", "o", "", "output file path (defaults to original filename)")
	rootCmd.AddCommand(downloadCmd)
}
