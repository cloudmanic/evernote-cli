// Copyright 2026. All rights reserved.
// Date: 2026-02-08
package cmd

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/dreampuf/evernote-sdk-golang/edam"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildResource(t *testing.T) {
	t.Run("valid file", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "test.pdf")
		content := []byte("fake pdf data")
		require.NoError(t, os.WriteFile(filePath, content, 0644))

		res, hash, err := buildResource(filePath)
		require.NoError(t, err)

		expectedHash := md5.Sum(content)
		assert.Equal(t, expectedHash[:], hash)
		assert.Equal(t, content, res.Data.Body)
		assert.Equal(t, int32(len(content)), res.Data.GetSize())
		assert.Equal(t, "application/pdf", res.GetMime())
		assert.Equal(t, "test.pdf", res.GetAttributes().GetFileName())
		assert.True(t, res.GetAttributes().GetAttachment())
	})

	t.Run("unknown extension defaults to octet-stream", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "data.xyz123")
		require.NoError(t, os.WriteFile(filePath, []byte("data"), 0644))

		res, _, err := buildResource(filePath)
		require.NoError(t, err)
		assert.Equal(t, "application/octet-stream", res.GetMime())
	})

	t.Run("file not found", func(t *testing.T) {
		_, _, err := buildResource("/nonexistent/file.pdf")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read file")
	})
}

func TestBuildMediaTag(t *testing.T) {
	hash := md5.Sum([]byte("test"))
	tag := buildMediaTag(hash[:], "application/pdf")
	assert.Contains(t, tag, "en-media")
	assert.Contains(t, tag, "application/pdf")
	assert.Contains(t, tag, fmt.Sprintf("%x", hash[:]))
}

func TestAttachCommand(t *testing.T) {
	originalJSON := jsonFlag
	defer func() { jsonFlag = originalJSON }()
	jsonFlag = false

	t.Run("attach single file", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "doc.pdf")
		require.NoError(t, os.WriteFile(filePath, []byte("pdf content"), 0644))

		guid := edam.GUID("note-123")
		title := "My Note"
		content := wrapENML("existing text")
		mock := &mockNoteStore{
			gotNote: &edam.Note{
				GUID:    &guid,
				Title:   &title,
				Content: &content,
			},
		}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		var buf bytes.Buffer
		attachCmd.SetOut(&buf)
		err := attachCmd.RunE(attachCmd, []string{"note-123", filePath})
		require.NoError(t, err)

		assert.Contains(t, buf.String(), "Attached 1 file(s)")
		assert.Contains(t, buf.String(), "My Note")
	})

	t.Run("attach multiple files", func(t *testing.T) {
		tempDir := t.TempDir()
		file1 := filepath.Join(tempDir, "photo.jpg")
		file2 := filepath.Join(tempDir, "report.pdf")
		require.NoError(t, os.WriteFile(file1, []byte("jpg data"), 0644))
		require.NoError(t, os.WriteFile(file2, []byte("pdf data"), 0644))

		guid := edam.GUID("note-456")
		title := "Multi Attach"
		content := wrapENML("body")
		mock := &mockNoteStore{
			gotNote: &edam.Note{
				GUID:    &guid,
				Title:   &title,
				Content: &content,
			},
		}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		var buf bytes.Buffer
		attachCmd.SetOut(&buf)
		err := attachCmd.RunE(attachCmd, []string{"note-456", file1, file2})
		require.NoError(t, err)

		assert.Contains(t, buf.String(), "Attached 2 file(s)")
	})

	t.Run("file not found", func(t *testing.T) {
		guid := edam.GUID("note-123")
		title := "My Note"
		content := wrapENML("text")
		mock := &mockNoteStore{
			gotNote: &edam.Note{
				GUID:    &guid,
				Title:   &title,
				Content: &content,
			},
		}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		err := attachCmd.RunE(attachCmd, []string{"note-123", "/nonexistent/file.pdf"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read file")
	})

	t.Run("auth error", func(t *testing.T) {
		mock := &mockNoteStore{err: fmt.Errorf("not authenticated")}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		err := attachCmd.RunE(attachCmd, []string{"note-123", "file.pdf"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not authenticated")
	})

	t.Run("API error on get note", func(t *testing.T) {
		mock := &mockNoteStore{
			gotNote: &edam.Note{},
			err:     fmt.Errorf("note not found"),
		}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		err := attachCmd.RunE(attachCmd, []string{"bad-guid", "file.pdf"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "note not found")
	})
}

func TestAttachCmdConfiguration(t *testing.T) {
	assert.Equal(t, "attach [note-guid] [file...]", attachCmd.Use)
	assert.Equal(t, "Attach files to an existing note", attachCmd.Short)
	assert.NotNil(t, attachCmd.RunE)
}

func TestAttachCmdRegistration(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Name() == "attach" {
			found = true
			assert.Equal(t, "Attach files to an existing note", c.Short)
			break
		}
	}
	assert.True(t, found, "attach command should be registered")
}
