package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/dreampuf/evernote-sdk-golang/edam"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddCmdConfiguration(t *testing.T) {
	assert.Equal(t, "add", addCmd.Use)
	assert.Equal(t, "Add a new note", addCmd.Short)
	assert.NotNil(t, addCmd.RunE)
}

func TestAddCmdFlagParsing(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().StringVar(&addBody, "body", "", "")
	cmd.Flags().StringVar(&addTitle, "title", "", "")
	cmd.Flags().StringVar(&addNotebook, "notebook", "", "")
	cmd.Flags().StringSliceVar(&addTags, "tags", nil, "")

	err := cmd.Flags().Parse([]string{"--body=test body", "--title=test title", "--notebook=nb1", "--tags=tag1,tag2"})
	require.NoError(t, err)

	assert.Equal(t, "test body", addBody)
	assert.Equal(t, "test title", addTitle)
	assert.Equal(t, "nb1", addNotebook)
	assert.Equal(t, []string{"tag1", "tag2"}, addTags)
}

func TestAddCommand(t *testing.T) {
	t.Run("successful note creation", func(t *testing.T) {
		createdTitle := "Created Note"
		createdGuid := edam.GUID("new-note-123")

		mock := &mockNoteStore{
			createdNote: &edam.Note{
				Title: &createdTitle,
				GUID:  &createdGuid,
			},
		}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		addTitle = "Test Note"
		addBody = "Test body content"
		addNotebook = ""
		addTags = nil

		var buf bytes.Buffer
		addCmd.SetOut(&buf)
		jsonFlag = false
		err := addCmd.RunE(addCmd, []string{})
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Note created: Created Note")
		assert.Contains(t, output, "GUID: new-note-123")
	})

	t.Run("note creation with JSON output", func(t *testing.T) {
		createdTitle := "JSON Note"
		createdGuid := edam.GUID("json-note-123")

		mock := &mockNoteStore{
			createdNote: &edam.Note{
				Title: &createdTitle,
				GUID:  &createdGuid,
			},
		}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		addTitle = "JSON Note"
		addBody = "Test body"
		addNotebook = ""
		addTags = nil

		var buf bytes.Buffer
		addCmd.SetOut(&buf)
		jsonFlag = true
		err := addCmd.RunE(addCmd, []string{})
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "JSON Note")
		assert.Contains(t, output, "json-note-123")
	})

	t.Run("title required", func(t *testing.T) {
		addTitle = ""
		addBody = "body"

		err := addCmd.RunE(addCmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "--title is required")
	})

	t.Run("API error", func(t *testing.T) {
		mock := &mockNoteStore{
			createdNote: &edam.Note{},
			err:         fmt.Errorf("quota exceeded"),
		}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		addTitle = "Test"
		addBody = "body"

		err := addCmd.RunE(addCmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "quota exceeded")
	})

	t.Run("auth error returns error", func(t *testing.T) {
		mock := &mockNoteStore{err: fmt.Errorf("not authenticated")}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		addTitle = "Test"
		err := addCmd.RunE(addCmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not authenticated")
	})
}

func TestWrapENML(t *testing.T) {
	t.Run("plain text", func(t *testing.T) {
		result := wrapENML("Hello world")
		assert.Contains(t, result, "<en-note>Hello world</en-note>")
		assert.Contains(t, result, "<?xml version")
		assert.Contains(t, result, "<!DOCTYPE en-note")
	})

	t.Run("text with special characters", func(t *testing.T) {
		result := wrapENML("Hello <world> & \"friends\"")
		assert.Contains(t, result, "Hello &lt;world&gt; &amp; &#34;friends&#34;")
		assert.NotContains(t, result, "<world>")
	})

	t.Run("empty body", func(t *testing.T) {
		result := wrapENML("")
		assert.Contains(t, result, "<en-note></en-note>")
	})
}

func TestAddCmdRegistration(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Name() == "add" {
			found = true
			assert.Equal(t, "Add a new note", c.Short)
			break
		}
	}
	assert.True(t, found, "add command should be registered")
}
