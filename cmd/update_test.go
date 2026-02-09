// Copyright 2026. All rights reserved.
// Date: 2026-02-08
package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/dreampuf/evernote-sdk-golang/edam"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateCommand(t *testing.T) {
	t.Run("update title only", func(t *testing.T) {
		guid := edam.GUID("note-123")
		existingTitle := "Old Title"
		existingContent := wrapENML("existing body")
		updatedTitle := "New Title"

		mock := &mockNoteStore{
			gotNote: &edam.Note{
				GUID:    &guid,
				Title:   &existingTitle,
				Content: &existingContent,
			},
			updatedNote: &edam.Note{
				GUID:  &guid,
				Title: &updatedTitle,
			},
		}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		updateTitle = "New Title"
		updateBody = ""
		updateAppend = ""
		updateTags = nil
		defer func() { updateTitle = "" }()

		var buf bytes.Buffer
		updateCmd.SetOut(&buf)
		err := updateCmd.RunE(updateCmd, []string{"note-123"})
		require.NoError(t, err)

		assert.Contains(t, buf.String(), "Note updated:")
		assert.Contains(t, buf.String(), "New Title")
	})

	t.Run("replace body", func(t *testing.T) {
		guid := edam.GUID("note-123")
		existingTitle := "My Note"
		existingContent := wrapENML("old content")

		mock := &mockNoteStore{
			gotNote: &edam.Note{
				GUID:    &guid,
				Title:   &existingTitle,
				Content: &existingContent,
			},
		}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		updateTitle = ""
		updateBody = "brand new content"
		updateAppend = ""
		updateTags = nil
		defer func() { updateBody = "" }()

		var buf bytes.Buffer
		updateCmd.SetOut(&buf)
		err := updateCmd.RunE(updateCmd, []string{"note-123"})
		require.NoError(t, err)

		assert.Contains(t, buf.String(), "Note updated:")
	})

	t.Run("append to existing content", func(t *testing.T) {
		guid := edam.GUID("note-123")
		existingTitle := "My Note"
		existingContent := wrapENML("original text")

		mock := &mockNoteStore{
			gotNote: &edam.Note{
				GUID:    &guid,
				Title:   &existingTitle,
				Content: &existingContent,
			},
		}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		updateTitle = ""
		updateBody = ""
		updateAppend = "appended text"
		updateTags = nil
		defer func() { updateAppend = "" }()

		var buf bytes.Buffer
		updateCmd.SetOut(&buf)
		err := updateCmd.RunE(updateCmd, []string{"note-123"})
		require.NoError(t, err)

		assert.Contains(t, buf.String(), "Note updated:")
	})

	t.Run("append to empty note", func(t *testing.T) {
		guid := edam.GUID("note-123")
		existingTitle := "Empty Note"
		existingContent := wrapENML("")

		mock := &mockNoteStore{
			gotNote: &edam.Note{
				GUID:    &guid,
				Title:   &existingTitle,
				Content: &existingContent,
			},
		}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		updateTitle = ""
		updateBody = ""
		updateAppend = "first content"
		updateTags = nil
		defer func() { updateAppend = "" }()

		var buf bytes.Buffer
		updateCmd.SetOut(&buf)
		err := updateCmd.RunE(updateCmd, []string{"note-123"})
		require.NoError(t, err)

		assert.Contains(t, buf.String(), "Note updated:")
	})

	t.Run("no flags provided", func(t *testing.T) {
		updateTitle = ""
		updateBody = ""
		updateAppend = ""
		updateTags = nil

		err := updateCmd.RunE(updateCmd, []string{"note-123"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least one of --title, --body, --append, or --tags is required")
	})

	t.Run("body and append conflict", func(t *testing.T) {
		updateTitle = ""
		updateBody = "new body"
		updateAppend = "append text"
		updateTags = nil
		defer func() {
			updateBody = ""
			updateAppend = ""
		}()

		err := updateCmd.RunE(updateCmd, []string{"note-123"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "--body and --append cannot be used together")
	})

	t.Run("API error on get", func(t *testing.T) {
		mock := &mockNoteStore{
			gotNote: &edam.Note{},
			err:     fmt.Errorf("note not found"),
		}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		updateTitle = "New Title"
		updateBody = ""
		updateAppend = ""
		updateTags = nil
		defer func() { updateTitle = "" }()

		err := updateCmd.RunE(updateCmd, []string{"bad-guid"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "note not found")
	})

	t.Run("auth error", func(t *testing.T) {
		mock := &mockNoteStore{err: fmt.Errorf("not authenticated")}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		updateTitle = "New Title"
		updateBody = ""
		updateAppend = ""
		updateTags = nil
		defer func() { updateTitle = "" }()

		err := updateCmd.RunE(updateCmd, []string{"any-guid"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not authenticated")
	})

	t.Run("update tags", func(t *testing.T) {
		guid := edam.GUID("note-123")
		existingTitle := "My Note"
		existingContent := wrapENML("some content")

		mock := &mockNoteStore{
			gotNote: &edam.Note{
				GUID:    &guid,
				Title:   &existingTitle,
				Content: &existingContent,
			},
		}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		updateTitle = ""
		updateBody = ""
		updateAppend = ""
		updateTags = []string{"tag1", "tag2"}
		defer func() { updateTags = nil }()

		var buf bytes.Buffer
		updateCmd.SetOut(&buf)
		err := updateCmd.RunE(updateCmd, []string{"note-123"})
		require.NoError(t, err)

		assert.Contains(t, buf.String(), "Note updated:")
	})
}

func TestUpdateCmdConfiguration(t *testing.T) {
	assert.Equal(t, "update [guid]", updateCmd.Use)
	assert.Equal(t, "Update an existing note", updateCmd.Short)
	assert.NotNil(t, updateCmd.RunE)
}

func TestUpdateCmdRegistration(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Name() == "update" {
			found = true
			assert.Equal(t, "Update an existing note", c.Short)
			break
		}
	}
	assert.True(t, found, "update command should be registered")
}
