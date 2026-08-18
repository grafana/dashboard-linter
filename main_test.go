package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLintCommandWithoutArgs(t *testing.T) {
	rootCmd.SetArgs([]string{"lint"})
	err := rootCmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "no dashboard file specified")
}

func TestLintCommandWithNonExistentFile(t *testing.T) {
	rootCmd.SetArgs([]string{"lint", "/path/that/does/not/exist.json"})
	err := rootCmd.Execute()
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "failed to read file"))
}
