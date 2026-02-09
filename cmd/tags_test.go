package cmd

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/dreampuf/evernote-sdk-golang/edam"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTagsCommand(t *testing.T) {
	t.Run("successful tags list with formatting", func(t *testing.T) {
		name1 := "Work"
		guid1 := edam.GUID("tag-12345")
		name2 := "Personal"
		guid2 := edam.GUID("tag-67890")

		mock := &mockNoteStore{
			tags: []*edam.Tag{
				{Name: &name1, GUID: &guid1},
				{Name: &name2, GUID: &guid2},
			},
		}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		var buf bytes.Buffer
		tagsCmd.SetOut(&buf)
		jsonFlag = false
		err := tagsCmd.RunE(tagsCmd, []string{})
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Found 2 tag(s):")
		assert.Contains(t, output, "1. Work")
		assert.Contains(t, output, "GUID: tag-12345")
		assert.Contains(t, output, "2. Personal")
		assert.Contains(t, output, "GUID: tag-67890")
	})

	t.Run("successful tags list with JSON output", func(t *testing.T) {
		name := "Work"
		guid := edam.GUID("tag-12345")

		mock := &mockNoteStore{
			tags: []*edam.Tag{
				{Name: &name, GUID: &guid},
			},
		}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		var buf bytes.Buffer
		tagsCmd.SetOut(&buf)
		jsonFlag = true
		err := tagsCmd.RunE(tagsCmd, []string{})
		require.NoError(t, err)

		output := buf.String()
		assert.Contains(t, output, "Work")
		assert.Contains(t, output, "tag-12345")
	})

	t.Run("empty tags list", func(t *testing.T) {
		mock := &mockNoteStore{
			tags: []*edam.Tag{},
		}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		var buf bytes.Buffer
		tagsCmd.SetOut(&buf)
		jsonFlag = false
		err := tagsCmd.RunE(tagsCmd, []string{})
		require.NoError(t, err)

		assert.Contains(t, buf.String(), "No tags found.")
	})

	t.Run("API error", func(t *testing.T) {
		mock := &mockNoteStore{
			tags: []*edam.Tag{},
			err:  fmt.Errorf("unauthorized"),
		}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		jsonFlag = false
		err := tagsCmd.RunE(tagsCmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})

	t.Run("auth error returns error", func(t *testing.T) {
		mock := &mockNoteStore{err: fmt.Errorf("not authenticated")}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		err := tagsCmd.RunE(tagsCmd, []string{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not authenticated")
	})
}

func TestTagsCmdConfiguration(t *testing.T) {
	assert.Equal(t, "tags", tagsCmd.Use)
	assert.Equal(t, "List all tags", tagsCmd.Short)
	assert.NotNil(t, tagsCmd.RunE)
}

func TestTagsCommandRegistration(t *testing.T) {
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "tags" {
			found = true
			assert.Equal(t, "List all tags", cmd.Short)
			break
		}
	}
	assert.True(t, found, "tags command should be registered with root command")
}
