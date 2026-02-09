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

func TestGetCommand(t *testing.T) {
	t.Run("successful get with content", func(t *testing.T) {
		title := "My Note"
		guid := edam.GUID("note-abc-123")
		nbGuid := "nb-123"
		content := `<?xml version="1.0" encoding="UTF-8"?><!DOCTYPE en-note SYSTEM "http://xml.evernote.com/note/note.dtd"><en-note>Hello world</en-note>`
		created := edam.Timestamp(1700000000000)
		updated := edam.Timestamp(1700100000000)

		mock := &mockNoteStore{
			gotNote: &edam.Note{
				Title:        &title,
				GUID:         &guid,
				NotebookGuid: &nbGuid,
				Content:      &content,
				Created:      &created,
				Updated:      &updated,
			},
		}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		var buf bytes.Buffer
		getCmd.SetOut(&buf)
		jsonFlag = false
		err := getCmd.RunE(getCmd, []string{"note-abc-123"})
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Title: My Note")
		assert.Contains(t, output, "GUID:  note-abc-123")
		assert.Contains(t, output, "Notebook: nb-123")
		assert.Contains(t, output, "Hello world")
		assert.Contains(t, output, "Created:")
		assert.Contains(t, output, "Updated:")
	})

	t.Run("get with JSON output", func(t *testing.T) {
		title := "JSON Note"
		guid := edam.GUID("note-json-1")

		mock := &mockNoteStore{
			gotNote: &edam.Note{
				Title: &title,
				GUID:  &guid,
			},
		}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		var buf bytes.Buffer
		getCmd.SetOut(&buf)
		jsonFlag = true
		err := getCmd.RunE(getCmd, []string{"note-json-1"})
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "JSON Note")
		assert.Contains(t, output, "note-json-1")
	})

	t.Run("get with tags", func(t *testing.T) {
		title := "Tagged Note"
		guid := edam.GUID("note-tagged")
		tagNames := []string{"Work", "Important"}

		mock := &mockNoteStore{
			gotNote: &edam.Note{
				Title:    &title,
				GUID:     &guid,
				TagNames: tagNames,
			},
		}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		var buf bytes.Buffer
		getCmd.SetOut(&buf)
		jsonFlag = false
		err := getCmd.RunE(getCmd, []string{"note-tagged"})
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Tags: Work, Important")
	})

	t.Run("API error", func(t *testing.T) {
		mock := &mockNoteStore{
			gotNote: &edam.Note{},
			err:     fmt.Errorf("note not found"),
		}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		err := getCmd.RunE(getCmd, []string{"bad-guid"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "note not found")
	})

	t.Run("auth error", func(t *testing.T) {
		mock := &mockNoteStore{err: fmt.Errorf("not authenticated")}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		err := getCmd.RunE(getCmd, []string{"any-guid"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not authenticated")
	})
}

func TestStripENML(t *testing.T) {
	t.Run("full ENML document", func(t *testing.T) {
		input := `<?xml version="1.0" encoding="UTF-8"?><!DOCTYPE en-note SYSTEM "http://xml.evernote.com/note/note.dtd"><en-note>Hello world</en-note>`
		result := stripENML(input)
		assert.Equal(t, "Hello world", result)
	})

	t.Run("ENML with line breaks", func(t *testing.T) {
		input := `<en-note>Line 1<br/>Line 2</en-note>`
		result := stripENML(input)
		assert.Contains(t, result, "Line 1")
		assert.Contains(t, result, "Line 2")
	})

	t.Run("empty content", func(t *testing.T) {
		result := stripENML("")
		assert.Equal(t, "", result)
	})

	t.Run("plain text passthrough", func(t *testing.T) {
		result := stripENML("Just plain text")
		assert.Equal(t, "Just plain text", result)
	})
}

func TestGetCmdConfiguration(t *testing.T) {
	assert.Equal(t, "get [guid]", getCmd.Use)
	assert.Equal(t, "Get a note by GUID", getCmd.Short)
	assert.NotNil(t, getCmd.RunE)
}

func TestGetCmdRegistration(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Name() == "get" {
			found = true
			assert.Equal(t, "Get a note by GUID", c.Short)
			break
		}
	}
	assert.True(t, found, "get command should be registered")
}
