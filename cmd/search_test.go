package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/dreampuf/evernote-sdk-golang/edam"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchCmdConfiguration(t *testing.T) {
	assert.Equal(t, "search [query]", searchCmd.Use)
	assert.Equal(t, "Search notes", searchCmd.Short)
	assert.NotNil(t, searchCmd.Args)
}

func TestSearchCommand(t *testing.T) {
	t.Run("successful search with results", func(t *testing.T) {
		title1 := "Meeting Notes"
		guid1 := edam.GUID("note-123")
		created1 := edam.Timestamp(1700000000000)
		updated1 := edam.Timestamp(1700100000000)
		title2 := "Project Plan"
		guid2 := edam.GUID("note-456")

		mock := &mockNoteStore{
			notes: &edam.NotesMetadataList{
				TotalNotes: 2,
				Notes: []*edam.NoteMetadata{
					{GUID: guid1, Title: &title1, Created: &created1, Updated: &updated1},
					{GUID: guid2, Title: &title2},
				},
			},
		}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		var buf bytes.Buffer
		searchCmd.SetOut(&buf)
		jsonFlag = false
		err := searchCmd.RunE(searchCmd, []string{"test query"})
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Found 2 note(s):")
		assert.Contains(t, output, "1. Meeting Notes")
		assert.Contains(t, output, "GUID: note-123")
		assert.Contains(t, output, "2. Project Plan")
	})

	t.Run("search with JSON output", func(t *testing.T) {
		title := "Test Note"
		guid := edam.GUID("note-789")

		mock := &mockNoteStore{
			notes: &edam.NotesMetadataList{
				TotalNotes: 1,
				Notes: []*edam.NoteMetadata{
					{GUID: guid, Title: &title},
				},
			},
		}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		var buf bytes.Buffer
		searchCmd.SetOut(&buf)
		jsonFlag = true
		err := searchCmd.RunE(searchCmd, []string{"test"})
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Test Note")
		assert.Contains(t, output, "note-789")
	})

	t.Run("search with no results", func(t *testing.T) {
		mock := &mockNoteStore{
			notes: &edam.NotesMetadataList{
				TotalNotes: 0,
				Notes:      []*edam.NoteMetadata{},
			},
		}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		var buf bytes.Buffer
		searchCmd.SetOut(&buf)
		jsonFlag = false
		err := searchCmd.RunE(searchCmd, []string{"nonexistent"})
		require.NoError(t, err)

		assert.Contains(t, buf.String(), "No notes found.")
	})

	t.Run("search API error", func(t *testing.T) {
		mock := &mockNoteStore{
			notes: &edam.NotesMetadataList{},
			err:   fmt.Errorf("service unavailable"),
		}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		jsonFlag = false
		err := searchCmd.RunE(searchCmd, []string{"test"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "service unavailable")
	})

	t.Run("auth error returns error", func(t *testing.T) {
		mock := &mockNoteStore{err: fmt.Errorf("not authenticated")}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		err := searchCmd.RunE(searchCmd, []string{"test"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not authenticated")
	})

	t.Run("multiple words joined as query", func(t *testing.T) {
		mock := &mockNoteStore{
			notes: &edam.NotesMetadataList{
				TotalNotes: 0,
				Notes:      []*edam.NoteMetadata{},
			},
		}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		var buf bytes.Buffer
		searchCmd.SetOut(&buf)
		jsonFlag = false
		err := searchCmd.RunE(searchCmd, []string{"hello", "world"})
		require.NoError(t, err)
	})
}

func TestSearchCmdRegistration(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "search" {
			found = true
			assert.Equal(t, "Search notes", cmd.Short)
			break
		}
	}
	assert.True(t, found, "search command should be registered with root command")
}
