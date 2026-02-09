package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dreampuf/evernote-sdk-golang/client"
	"github.com/dreampuf/evernote-sdk-golang/edam"
	"github.com/spf13/cobra"
)

// jsonFlag is used by subcommands to output JSON.
var jsonFlag bool
var configPath = filepath.Join(os.Getenv("HOME"), ".config", "evernote", "auth.json")

// Config holds the Evernote API credentials and auth token.
type Config struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	AuthToken    string `json:"auth_token"`
	NoteStoreURL string `json:"note_store_url"`
}

// noteStoreClient defines the interface for Evernote NoteStore operations.
type noteStoreClient interface {
	ListNotebooks(ctx context.Context, authenticationToken string) ([]*edam.Notebook, error)
	ListTags(ctx context.Context, authenticationToken string) ([]*edam.Tag, error)
	FindNotesMetadata(ctx context.Context, authenticationToken string, filter *edam.NoteFilter, offset int32, maxNotes int32, resultSpec *edam.NotesMetadataResultSpec) (*edam.NotesMetadataList, error)
	CreateNote(ctx context.Context, authenticationToken string, note *edam.Note) (*edam.Note, error)
	GetNote(ctx context.Context, authenticationToken string, guid edam.GUID, withContent bool, withResourcesData bool, withResourcesRecognition bool, withResourcesAlternateData bool) (*edam.Note, error)
	GetResource(ctx context.Context, authenticationToken string, guid edam.GUID, withData bool, withRecognition bool, withAttributes bool, withAlternateData bool) (*edam.Resource, error)
	UpdateNote(ctx context.Context, authenticationToken string, note *edam.Note) (*edam.Note, error)
}

// getNoteStoreFunc returns a NoteStore client and auth token. Can be overridden in tests.
var getNoteStoreFunc = getDefaultNoteStore

// getDefaultNoteStore loads config and creates a NoteStore client using the Evernote SDK.
func getDefaultNoteStore() (noteStoreClient, string, error) {
	cfg, err := loadConfig()
	if err != nil {
		return nil, "", fmt.Errorf("could not read config: %w", err)
	}
	if cfg.AuthToken == "" {
		return nil, "", fmt.Errorf("not authenticated, run 'evernote-cli init' or 'evernote-cli auth'")
	}

	c := client.NewClient(cfg.ClientID, cfg.ClientSecret, client.PRODUCTION)

	var ns *edam.NoteStoreClient
	if cfg.NoteStoreURL != "" {
		ns, err = c.GetNoteStoreWithURL(cfg.NoteStoreURL)
	} else {
		ns, err = c.GetNoteStore(context.Background(), cfg.AuthToken)
	}
	if err != nil {
		return nil, "", fmt.Errorf("failed to connect to Evernote: %w", err)
	}

	return ns, cfg.AuthToken, nil
}

// loadConfig reads the config file from disk.
func loadConfig() (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}
	return &c, nil
}

// saveConfig writes the config file to disk with secure permissions.
func saveConfig(c *Config) error {
	os.MkdirAll(filepath.Dir(configPath), 0700)
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0600)
}

// formatAPIError converts Evernote SDK exceptions into human-readable error messages.
func formatAPIError(err error) error {
	var sysErr *edam.EDAMSystemException
	if errors.As(err, &sysErr) {
		if sysErr.GetErrorCode() == edam.EDAMErrorCode_RATE_LIMIT_REACHED {
			msg := sysErr.GetMessage()
			duration := sysErr.GetRateLimitDuration()
			if strings.Contains(msg, "RTE room has already been open") {
				return fmt.Errorf("note is currently open in Evernote, close it there first then retry")
			}
			if duration > 0 {
				return fmt.Errorf("rate limited by Evernote, try again in %d seconds", duration)
			}
			return fmt.Errorf("rate limited by Evernote: %s", msg)
		}
		msg := sysErr.GetMessage()
		if msg != "" {
			return fmt.Errorf("Evernote system error (%s): %s", sysErr.GetErrorCode(), msg)
		}
		return fmt.Errorf("Evernote system error: %s", sysErr.GetErrorCode())
	}

	var userErr *edam.EDAMUserException
	if errors.As(err, &userErr) {
		param := userErr.GetParameter()
		if param != "" {
			return fmt.Errorf("invalid request (%s): %s", userErr.GetErrorCode(), param)
		}
		return fmt.Errorf("invalid request: %s", userErr.GetErrorCode())
	}

	var notFound *edam.EDAMNotFoundException
	if errors.As(err, &notFound) {
		id := notFound.GetIdentifier()
		key := notFound.GetKey()
		if id != "" && key != "" {
			return fmt.Errorf("not found: %s = %s", id, key)
		}
		if id != "" {
			return fmt.Errorf("not found: %s", id)
		}
		return fmt.Errorf("not found")
	}

	return err
}

// rootCmd is the main command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "evernote-cli",
	Short: "A CLI tool to interact with Evernote",
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "output in JSON format")
}
