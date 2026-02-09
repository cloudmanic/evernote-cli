package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/dreampuf/evernote-sdk-golang/edam"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotebooksCommand(t *testing.T) {
	t.Run("successful notebooks list with formatting", func(t *testing.T) {
		name1 := "Default Notebook"
		guid1 := edam.GUID("12345")
		def1 := true
		name2 := "Work Notes"
		guid2 := edam.GUID("67890")

		mock := &mockNoteStore{
			notebooks: []*edam.Notebook{
				{Name: &name1, GUID: &guid1, DefaultNotebook: &def1},
				{Name: &name2, GUID: &guid2},
			},
		}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		var buf bytes.Buffer
		notebooksCmd.SetOut(&buf)
		jsonFlag = false
		err := notebooksCmd.RunE(notebooksCmd, []string{})
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Found 2 notebook(s):")
		assert.Contains(t, output, "1. Default Notebook (default)")
		assert.Contains(t, output, "GUID: 12345")
		assert.Contains(t, output, "2. Work Notes")
		assert.Contains(t, output, "GUID: 67890")
	})

	t.Run("successful notebooks list with JSON output", func(t *testing.T) {
		name := "Default Notebook"
		guid := edam.GUID("12345")
		def := true

		mock := &mockNoteStore{
			notebooks: []*edam.Notebook{
				{Name: &name, GUID: &guid, DefaultNotebook: &def},
			},
		}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		var buf bytes.Buffer
		notebooksCmd.SetOut(&buf)
		jsonFlag = true
		err := notebooksCmd.RunE(notebooksCmd, []string{})
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Default Notebook")
		assert.Contains(t, output, "12345")
	})

	t.Run("empty notebooks list", func(t *testing.T) {
		mock := &mockNoteStore{
			notebooks: []*edam.Notebook{},
		}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		var buf bytes.Buffer
		notebooksCmd.SetOut(&buf)
		jsonFlag = false
		err := notebooksCmd.RunE(notebooksCmd, []string{})
		require.NoError(t, err)

		assert.Contains(t, buf.String(), "No notebooks found.")
	})

	t.Run("API error", func(t *testing.T) {
		mock := &mockNoteStore{
			notebooks: []*edam.Notebook{},
			err:       fmt.Errorf("auth expired"),
		}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		jsonFlag = false
		err := notebooksCmd.RunE(notebooksCmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "auth expired")
	})

	t.Run("auth error returns error", func(t *testing.T) {
		mock := &mockNoteStore{err: fmt.Errorf("not authenticated")}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		err := notebooksCmd.RunE(notebooksCmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not authenticated")
	})
}

func TestNotebooksCmdConfiguration(t *testing.T) {
	assert.Equal(t, "notebooks", notebooksCmd.Use)
	assert.Equal(t, "List all notebooks", notebooksCmd.Short)
	assert.NotNil(t, notebooksCmd.RunE)
}

func TestNotebooksCommandRegistration(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "notebooks" {
			found = true
			assert.Equal(t, "List all notebooks", cmd.Short)
			break
		}
	}
	assert.True(t, found, "notebooks command should be registered with root command")
}
