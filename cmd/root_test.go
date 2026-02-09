package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/dreampuf/evernote-sdk-golang/edam"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockNoteStore implements the noteStoreClient interface for testing.
type mockNoteStore struct {
	notebooks   []*edam.Notebook
	tags        []*edam.Tag
	notes       *edam.NotesMetadataList
	createdNote *edam.Note
	gotNote     *edam.Note
	updatedNote *edam.Note
	resource    *edam.Resource
	err         error
}

// ListNotebooks returns the mock notebooks.
func (m *mockNoteStore) ListNotebooks(ctx context.Context, authenticationToken string) ([]*edam.Notebook, error) {
	return m.notebooks, m.err
}

// ListTags returns the mock tags.
func (m *mockNoteStore) ListTags(ctx context.Context, authenticationToken string) ([]*edam.Tag, error) {
	return m.tags, m.err
}

// FindNotesMetadata returns the mock note metadata.
func (m *mockNoteStore) FindNotesMetadata(ctx context.Context, authenticationToken string, filter *edam.NoteFilter, offset int32, maxNotes int32, resultSpec *edam.NotesMetadataResultSpec) (*edam.NotesMetadataList, error) {
	return m.notes, m.err
}

// CreateNote returns the mock created note.
func (m *mockNoteStore) CreateNote(ctx context.Context, authenticationToken string, note *edam.Note) (*edam.Note, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.createdNote != nil {
		return m.createdNote, nil
	}
	return note, nil
}

// GetNote returns the mock note.
func (m *mockNoteStore) GetNote(ctx context.Context, authenticationToken string, guid edam.GUID, withContent bool, withResourcesData bool, withResourcesRecognition bool, withResourcesAlternateData bool) (*edam.Note, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.gotNote, nil
}

// GetResource returns the mock resource.
func (m *mockNoteStore) GetResource(ctx context.Context, authenticationToken string, guid edam.GUID, withData bool, withRecognition bool, withAttributes bool, withAlternateData bool) (*edam.Resource, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.resource, nil
}

// UpdateNote returns the mock updated note.
func (m *mockNoteStore) UpdateNote(ctx context.Context, authenticationToken string, note *edam.Note) (*edam.Note, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.updatedNote != nil {
		return m.updatedNote, nil
	}
	return note, nil
}

// setMockNoteStore overrides getNoteStoreFunc for testing and returns a cleanup function.
func setMockNoteStore(mock *mockNoteStore) func() {
	original := getNoteStoreFunc
	getNoteStoreFunc = func() (noteStoreClient, string, error) {
		if mock.err != nil && mock.notebooks == nil && mock.tags == nil && mock.notes == nil && mock.createdNote == nil && mock.gotNote == nil && mock.updatedNote == nil && mock.resource == nil {
			return nil, "", mock.err
		}
		return mock, "test-token", nil
	}
	return func() { getNoteStoreFunc = original }
}

func TestLoadConfig(t *testing.T) {
	tempDir := t.TempDir()
	originalConfigPath := configPath
	defer func() { configPath = originalConfigPath }()

	t.Run("successful load", func(t *testing.T) {
		configPath = filepath.Join(tempDir, "test-config.json")

		testConfig := Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			AuthToken:    "test-auth-token",
			NoteStoreURL: "https://www.evernote.com/shard/s1/notestore",
		}

		data, err := json.MarshalIndent(testConfig, "", "  ")
		require.NoError(t, err)
		err = os.WriteFile(configPath, data, 0600)
		require.NoError(t, err)

		config, err := loadConfig()
		require.NoError(t, err)
		assert.Equal(t, "test-client-id", config.ClientID)
		assert.Equal(t, "test-client-secret", config.ClientSecret)
		assert.Equal(t, "test-auth-token", config.AuthToken)
		assert.Equal(t, "https://www.evernote.com/shard/s1/notestore", config.NoteStoreURL)
	})

	t.Run("file does not exist", func(t *testing.T) {
		configPath = filepath.Join(tempDir, "nonexistent-config.json")

		config, err := loadConfig()
		assert.Error(t, err)
		assert.Nil(t, config)
		assert.Contains(t, err.Error(), "no such file or directory")
	})

	t.Run("invalid JSON", func(t *testing.T) {
		configPath = filepath.Join(tempDir, "invalid-config.json")

		err := os.WriteFile(configPath, []byte("invalid json content"), 0600)
		require.NoError(t, err)

		config, err := loadConfig()
		assert.Error(t, err)
		assert.Nil(t, config)
	})

	t.Run("empty file", func(t *testing.T) {
		configPath = filepath.Join(tempDir, "empty-config.json")

		err := os.WriteFile(configPath, []byte(""), 0600)
		require.NoError(t, err)

		config, err := loadConfig()
		assert.Error(t, err)
		assert.Nil(t, config)
	})
}

func TestSaveConfig(t *testing.T) {
	tempDir := t.TempDir()
	originalConfigPath := configPath
	defer func() { configPath = originalConfigPath }()

	t.Run("successful save", func(t *testing.T) {
		configPath = filepath.Join(tempDir, "subdir", "test-config.json")

		testConfig := &Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			AuthToken:    "test-auth-token",
			NoteStoreURL: "https://www.evernote.com/shard/s1/notestore",
		}

		err := saveConfig(testConfig)
		require.NoError(t, err)

		assert.FileExists(t, configPath)

		data, err := os.ReadFile(configPath)
		require.NoError(t, err)

		var savedConfig Config
		err = json.Unmarshal(data, &savedConfig)
		require.NoError(t, err)

		assert.Equal(t, testConfig.ClientID, savedConfig.ClientID)
		assert.Equal(t, testConfig.ClientSecret, savedConfig.ClientSecret)
		assert.Equal(t, testConfig.AuthToken, savedConfig.AuthToken)
		assert.Equal(t, testConfig.NoteStoreURL, savedConfig.NoteStoreURL)
	})

	t.Run("save config without token", func(t *testing.T) {
		configPath = filepath.Join(tempDir, "no-token-config.json")

		testConfig := &Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
		}

		err := saveConfig(testConfig)
		require.NoError(t, err)

		data, err := os.ReadFile(configPath)
		require.NoError(t, err)

		var savedConfig Config
		err = json.Unmarshal(data, &savedConfig)
		require.NoError(t, err)

		assert.Equal(t, testConfig.ClientID, savedConfig.ClientID)
		assert.Equal(t, testConfig.ClientSecret, savedConfig.ClientSecret)
		assert.Empty(t, savedConfig.AuthToken)
	})

	t.Run("save to protected directory", func(t *testing.T) {
		if os.Geteuid() == 0 {
			t.Skip("skipping permissions test when running as root")
		}
		configPath = "/root/protected/test-config.json"

		testConfig := &Config{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
		}

		err := saveConfig(testConfig)
		assert.Error(t, err)
	})
}

func TestGetNoteStoreFunc_NoConfig(t *testing.T) {
	tempDir := t.TempDir()
	originalConfigPath := configPath
	defer func() { configPath = originalConfigPath }()

	configPath = filepath.Join(tempDir, "nonexistent.json")

	ns, token, err := getDefaultNoteStore()
	assert.Error(t, err)
	assert.Nil(t, ns)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "could not read config")
}

func TestGetNoteStoreFunc_NoToken(t *testing.T) {
	tempDir := t.TempDir()
	originalConfigPath := configPath
	defer func() { configPath = originalConfigPath }()

	configPath = filepath.Join(tempDir, "no-token.json")
	cfg := &Config{ClientID: "id", ClientSecret: "secret"}
	require.NoError(t, saveConfig(cfg))

	ns, token, err := getDefaultNoteStore()
	assert.Error(t, err)
	assert.Nil(t, ns)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "not authenticated")
}

func TestConfig_JSONRoundTrip(t *testing.T) {
	original := Config{
		ClientID:     "test-id",
		ClientSecret: "test-secret",
		AuthToken:    "S=s1:U=abc:E=123:C=456:P=1:A=test:V=2:H=abc123",
		NoteStoreURL: "https://www.evernote.com/shard/s1/notestore",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var unmarshaled Config
	err = json.Unmarshal(data, &unmarshaled)
	require.NoError(t, err)

	assert.Equal(t, original.ClientID, unmarshaled.ClientID)
	assert.Equal(t, original.ClientSecret, unmarshaled.ClientSecret)
	assert.Equal(t, original.AuthToken, unmarshaled.AuthToken)
	assert.Equal(t, original.NoteStoreURL, unmarshaled.NoteStoreURL)
}

func TestConfig_EmptyFieldsOmission(t *testing.T) {
	config := Config{
		ClientID:     "test-id",
		ClientSecret: "test-secret",
	}

	data, err := json.Marshal(config)
	require.NoError(t, err)

	jsonStr := string(data)
	assert.Contains(t, jsonStr, "client_id")
	assert.Contains(t, jsonStr, "client_secret")
}

func TestMockNoteStore(t *testing.T) {
	t.Run("mock returns configured error", func(t *testing.T) {
		mock := &mockNoteStore{err: fmt.Errorf("connection failed")}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		_, _, err := getNoteStoreFunc()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connection failed")
	})

	t.Run("mock returns configured notebooks", func(t *testing.T) {
		name := "Test Notebook"
		guid := edam.GUID("nb-123")
		mock := &mockNoteStore{
			notebooks: []*edam.Notebook{{Name: &name, GUID: &guid}},
		}
		cleanup := setMockNoteStore(mock)
		defer cleanup()

		ns, token, err := getNoteStoreFunc()
		require.NoError(t, err)
		assert.Equal(t, "test-token", token)

		notebooks, err := ns.ListNotebooks(context.Background(), token)
		require.NoError(t, err)
		assert.Len(t, notebooks, 1)
		assert.Equal(t, "Test Notebook", notebooks[0].GetName())
	})
}

func TestFormatAPIError(t *testing.T) {
	t.Run("rate limit error", func(t *testing.T) {
		duration := int32(60)
		err := &edam.EDAMSystemException{
			ErrorCode:         edam.EDAMErrorCode_RATE_LIMIT_REACHED,
			RateLimitDuration: &duration,
		}
		result := formatAPIError(err)
		assert.Contains(t, result.Error(), "rate limited by Evernote")
		assert.Contains(t, result.Error(), "60 seconds")
	})

	t.Run("system error with message", func(t *testing.T) {
		msg := "internal failure"
		err := &edam.EDAMSystemException{
			ErrorCode: edam.EDAMErrorCode_INTERNAL_ERROR,
			Message:   &msg,
		}
		result := formatAPIError(err)
		assert.Contains(t, result.Error(), "internal failure")
		assert.Contains(t, result.Error(), "INTERNAL_ERROR")
	})

	t.Run("system error without message", func(t *testing.T) {
		err := &edam.EDAMSystemException{
			ErrorCode: edam.EDAMErrorCode_INTERNAL_ERROR,
		}
		result := formatAPIError(err)
		assert.Contains(t, result.Error(), "INTERNAL_ERROR")
	})

	t.Run("user error with parameter", func(t *testing.T) {
		param := "Note.title"
		err := &edam.EDAMUserException{
			ErrorCode: edam.EDAMErrorCode_BAD_DATA_FORMAT,
			Parameter: &param,
		}
		result := formatAPIError(err)
		assert.Contains(t, result.Error(), "Note.title")
		assert.Contains(t, result.Error(), "BAD_DATA_FORMAT")
	})

	t.Run("user error without parameter", func(t *testing.T) {
		err := &edam.EDAMUserException{
			ErrorCode: edam.EDAMErrorCode_PERMISSION_DENIED,
		}
		result := formatAPIError(err)
		assert.Contains(t, result.Error(), "PERMISSION_DENIED")
	})

	t.Run("not found error with identifier and key", func(t *testing.T) {
		id := "Note.guid"
		key := "abc-123"
		err := &edam.EDAMNotFoundException{
			Identifier: &id,
			Key:        &key,
		}
		result := formatAPIError(err)
		assert.Contains(t, result.Error(), "Note.guid")
		assert.Contains(t, result.Error(), "abc-123")
	})

	t.Run("not found error with identifier only", func(t *testing.T) {
		id := "Note.guid"
		err := &edam.EDAMNotFoundException{
			Identifier: &id,
		}
		result := formatAPIError(err)
		assert.Contains(t, result.Error(), "Note.guid")
	})

	t.Run("not found error bare", func(t *testing.T) {
		err := &edam.EDAMNotFoundException{}
		result := formatAPIError(err)
		assert.Contains(t, result.Error(), "not found")
	})

	t.Run("regular error passes through", func(t *testing.T) {
		err := fmt.Errorf("some other error")
		result := formatAPIError(err)
		assert.Equal(t, "some other error", result.Error())
	})
}
