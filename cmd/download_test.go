// Copyright 2026. All rights reserved.
// Date: 2026-02-08
package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/dreampuf/evernote-sdk-golang/edam"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownloadCommand(t *testing.T) {
	t.Run("successful download with original filename", func(t *testing.T) {
		tempDir := t.TempDir()
		fileName := "test-doc.pdf"
		mime := "application/pdf"
		guid := edam.GUID("res-123")
		fileData := []byte("fake pdf content")

		mock := &mockNoteStore{
			resource: &edam.Resource{
				GUID: &guid,
				Mime: &mime,
				Data: &edam.Data{Body: fileData},
				Attributes: &edam.ResourceAttributes{
					FileName: &fileName,
				},
			},
		}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		downloadOutput = filepath.Join(tempDir, "test-doc.pdf")
		var buf bytes.Buffer
		downloadCmd.SetOut(&buf)
		err := downloadCmd.RunE(downloadCmd, []string{"res-123"})
		require.NoError(t, err)

		assert.Contains(t, buf.String(), "Downloaded:")
		assert.Contains(t, buf.String(), "16 bytes")

		data, err := os.ReadFile(filepath.Join(tempDir, "test-doc.pdf"))
		require.NoError(t, err)
		assert.Equal(t, fileData, data)
	})

	t.Run("API error", func(t *testing.T) {
		mock := &mockNoteStore{
			resource: &edam.Resource{},
			err:      fmt.Errorf("resource not found"),
		}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		err := downloadCmd.RunE(downloadCmd, []string{"bad-guid"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "resource not found")
	})

	t.Run("auth error", func(t *testing.T) {
		mock := &mockNoteStore{err: fmt.Errorf("not authenticated")}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		err := downloadCmd.RunE(downloadCmd, []string{"any-guid"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not authenticated")
	})
}

func TestDownloadCmdConfiguration(t *testing.T) {
	assert.Equal(t, "download [resource-guid]", downloadCmd.Use)
	assert.Equal(t, "Download a note attachment by resource GUID", downloadCmd.Short)
	assert.NotNil(t, downloadCmd.RunE)
}

func TestDownloadCmdRegistration(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Name() == "download" {
			found = true
			assert.Equal(t, "Download a note attachment by resource GUID", c.Short)
			break
		}
	}
	assert.True(t, found, "download command should be registered")
}
