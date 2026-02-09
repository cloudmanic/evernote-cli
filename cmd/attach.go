// Copyright 2026. All rights reserved.
// Date: 2026-02-08
package cmd

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"html"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/dreampuf/evernote-sdk-golang/edam"
	"github.com/spf13/cobra"
)

// buildResource reads a file from disk and returns an Evernote Resource with its MD5 hash.
func buildResource(filePath string) (*edam.Resource, []byte, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	hash := md5.Sum(data)
	hashSlice := hash[:]
	size := int32(len(data))
	fileName := filepath.Base(filePath)
	isAttachment := true

	// Detect MIME type from file extension, stripping any parameters (e.g. charset)
	mimeType, _, _ := mime.ParseMediaType(mime.TypeByExtension(filepath.Ext(filePath)))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	resource := &edam.Resource{
		Data: &edam.Data{
			Body:     data,
			Size:     &size,
			BodyHash: hashSlice,
		},
		Mime: &mimeType,
		Attributes: &edam.ResourceAttributes{
			FileName:   &fileName,
			Attachment: &isAttachment,
		},
	}

	return resource, hashSlice, nil
}

// buildMediaTag returns an ENML <en-media> tag for a resource with the given hash and MIME type.
func buildMediaTag(hash []byte, mimeType string) string {
	return fmt.Sprintf(`<en-media type="%s" hash="%x"/>`, html.EscapeString(mimeType), hash)
}

// attachCmd attaches one or more files to an existing note.
var attachCmd = &cobra.Command{
	Use:   "attach [note-guid] [file...]",
	Short: "Attach files to an existing note",
	Long: `Attach one or more files to an existing note by GUID.

Examples:
  evernote-cli attach <guid> document.pdf
  evernote-cli attach <guid> photo.jpg report.pdf data.csv`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ns, token, err := getNoteStoreFunc()
		if err != nil {
			return err
		}

		guid := edam.GUID(args[0])

		// Fetch the existing note with content so we can append media tags
		existing, err := ns.GetNote(context.Background(), token, guid, true, true, false, false)
		if err != nil {
			return fmt.Errorf("failed to get note: %w", formatAPIError(err))
		}

		// Build resources from file arguments
		var resources []*edam.Resource
		var mediaTags []string
		for _, filePath := range args[1:] {
			res, hash, err := buildResource(filePath)
			if err != nil {
				return err
			}
			resources = append(resources, res)
			mediaTags = append(mediaTags, buildMediaTag(hash, res.GetMime()))
		}

		// Preserve existing resources by including them (metadata only, no body needed for existing)
		for _, existingRes := range existing.GetResources() {
			resources = append(resources, existingRes)
		}

		// Build updated content with media tags appended before </en-note>
		content := existing.GetContent()
		mediaBlock := strings.Join(mediaTags, "")
		content = strings.Replace(content, "</en-note>", mediaBlock+"</en-note>", 1)

		// Build update note
		title := existing.GetTitle()
		note := &edam.Note{
			GUID:      &guid,
			Title:     &title,
			Content:   &content,
			Resources: resources,
		}

		updated, err := ns.UpdateNote(context.Background(), token, note)
		if err != nil {
			return fmt.Errorf("failed to attach files: %w", formatAPIError(err))
		}

		if jsonFlag {
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetIndent("", "  ")
			return enc.Encode(updated)
		}

		fmt.Fprintf(cmd.OutOrStdout(), "Attached %d file(s) to note: %s\n", len(args[1:]), updated.GetTitle())
		fmt.Fprintf(cmd.OutOrStdout(), "GUID: %s\n", updated.GetGUID())
		return nil
	},
}

func init() {
	rootCmd.AddCommand(attachCmd)
}
