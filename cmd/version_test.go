// Copyright 2026. All rights reserved.
// Date: 2026-02-06
package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVersionCommand(t *testing.T) {
	t.Run("prints default dev version", func(t *testing.T) {
		original := Version
		Version = "dev"
		defer func() { Version = original }()

		var buf bytes.Buffer
		versionCmd.SetOut(&buf)
		err := versionCmd.RunE(versionCmd, []string{})
		require.NoError(t, err)

		assert.Equal(t, "evernote-cli version dev\n", buf.String())
	})

	t.Run("prints injected version", func(t *testing.T) {
		original := Version
		Version = "20260209130350-5179b6b"
		defer func() { Version = original }()

		var buf bytes.Buffer
		versionCmd.SetOut(&buf)
		err := versionCmd.RunE(versionCmd, []string{})
		require.NoError(t, err)

		assert.Equal(t, "evernote-cli version 20260209130350-5179b6b\n", buf.String())
	})
}

func TestVersionCmdConfiguration(t *testing.T) {
	assert.Equal(t, "version", versionCmd.Use)
	assert.Equal(t, "Print the version of evernote-cli", versionCmd.Short)
	assert.NotNil(t, versionCmd.RunE)
}

func TestVersionCmdRegistration(t *testing.T) {
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Name() == "version" {
			found = true
			assert.Equal(t, "Print the version of evernote-cli", c.Short)
			break
		}
	}
	assert.True(t, found, "version command should be registered")
}
