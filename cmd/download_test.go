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

	t.Run("path traversal attack prevented", func(t *testing.T) {
		tempDir := t.TempDir()
		// Malicious filename attempting path traversal
		maliciousFileName := "../../etc/passwd"
		mime := "text/plain"
		guid := edam.GUID("res-malicious")
		fileData := []byte("malicious content")

		mock := &mockNoteStore{
			resource: &edam.Resource{
				GUID: &guid,
				Mime: &mime,
				Data: &edam.Data{Body: fileData},
				Attributes: &edam.ResourceAttributes{
					FileName: &maliciousFileName,
				},
			},
		}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		// Change to temp directory to test
		oldWd, _ := os.Getwd()
		os.Chdir(tempDir)
		defer os.Chdir(oldWd)

		downloadOutput = "" // Let it use the filename from resource
		var buf bytes.Buffer
		downloadCmd.SetOut(&buf)
		err := downloadCmd.RunE(downloadCmd, []string{"res-malicious"})
		require.NoError(t, err)

		// File should be created in current directory with sanitized name (just "passwd")
		// not in ../../etc/passwd
		data, err := os.ReadFile("passwd")
		require.NoError(t, err)
		assert.Equal(t, fileData, data)

		// Verify file was NOT created in parent directories
		_, err = os.Stat("../../etc/passwd")
		assert.True(t, os.IsNotExist(err), "file should not exist in parent directory")
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
