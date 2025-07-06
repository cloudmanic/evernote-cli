package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
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

func TestAddHTTPRequest(t *testing.T) {
	tempDir := t.TempDir()
	originalConfigPath := configPath
	configPath = filepath.Join(tempDir, "auth.json")
	defer func() { configPath = originalConfigPath }()

	testConfig := &Config{
		ClientID:     "id",
		ClientSecret: "secret",
		Token: &oauth2.Token{
			AccessToken: "token",
			TokenType:   "Bearer",
		},
	}
	testConfig.Token.Expiry = time.Now().Add(time.Hour)
	require.NoError(t, saveConfig(testConfig))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/v1/notes", r.URL.Path)
		assert.Equal(t, "Bearer token", r.Header.Get("Authorization"))
		var body map[string]interface{}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "title text", body["title"])
		assert.Equal(t, "body text", body["content"])
		assert.Equal(t, "nb", body["notebook"])
		tags, ok := body["tags"].([]interface{})
		require.True(t, ok)
		assert.Len(t, tags, 2)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"ok": "true"})
	}))
	defer server.Close()

	cmd := &cobra.Command{
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := checkAuth()
			if err != nil {
				return err
			}
			payload := map[string]interface{}{
				"title":    addTitle,
				"content":  addBody,
				"notebook": addNotebook,
				"tags":     addTags,
			}
			data, _ := json.Marshal(payload)
			req, err := http.NewRequest("POST", server.URL+"/v1/notes", bytes.NewReader(data))
			if err != nil {
				return err
			}
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				return fmt.Errorf("unexpected status: %s", resp.Status)
			}
			var dataResp interface{}
			return json.NewDecoder(resp.Body).Decode(&dataResp)
		},
	}
	cmd.Flags().StringVar(&addBody, "body", "", "")
	cmd.Flags().StringVar(&addTitle, "title", "", "")
	cmd.Flags().StringVar(&addNotebook, "notebook", "", "")
	cmd.Flags().StringSliceVar(&addTags, "tags", nil, "")

	cmd.SetArgs([]string{"--body=body text", "--title=title text", "--notebook=nb", "--tags=tag1,tag2"})

	err := cmd.Execute()
	require.NoError(t, err)
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
