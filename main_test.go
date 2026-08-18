package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafana/dashboard-linter/lint"
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

func TestWriteMergesAndFixesEditable(t *testing.T) {
	original := `{"title":"t","editable":true,"panels":[{"id":1,"title":"p","type":"timeseries"}],"customField":{"nested":[1,2]}}`
	dir := t.TempDir()
	filename := filepath.Join(dir, "dashboard.json")
	require.NoError(t, os.WriteFile(filename, []byte(original), 0644))

	dashboard, err := lint.NewDashboard([]byte(original))
	require.NoError(t, err)
	dashboard.Editable = false

	require.NoError(t, write(dashboard, filename, []byte(original)))

	out, err := os.ReadFile(filename)
	require.NoError(t, err)
	content := string(out)

	// фикс применяется
	require.Contains(t, content, `"editable": false`)
	// мусорные null-поля не появляются
	require.NotContains(t, content, `"__inputs"`)
	require.NotContains(t, content, `"templating"`)
	require.NotContains(t, content, `"annotations"`)
	require.NotContains(t, content, `"hide"`)
	// неизвестные линтеру поля сохраняются
	require.Contains(t, content, `"customField"`)
	require.Contains(t, content, `"nested"`)
	// форматирование читаемо
	require.Contains(t, content, "\n  \"editable\"")
}

func TestWritePreservesFilePermissions(t *testing.T) {
	original := `{"title":"t","editable":true,"panels":[]}`
	dir := t.TempDir()
	filename := filepath.Join(dir, "dashboard.json")
	require.NoError(t, os.WriteFile(filename, []byte(original), 0755))

	dashboard, err := lint.NewDashboard([]byte(original))
	require.NoError(t, err)
	dashboard.Editable = false

	require.NoError(t, write(dashboard, filename, []byte(original)))

	info, err := os.Stat(filename)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0755), info.Mode().Perm())
}

func TestMergeMapsPreservesOldArray(t *testing.T) {
	oldMap := map[string]interface{}{
		"panels": []interface{}{map[string]interface{}{"id": float64(1)}},
	}
	newMap := map[string]interface{}{
		"panels": []interface{}{map[string]interface{}{"id": float64(2), "title": "p"}},
	}
	merged := mergeMaps(oldMap, newMap)
	panels := merged["panels"].([]interface{})
	require.Len(t, panels, 1)
	require.Equal(t, "p", panels[0].(map[string]interface{})["title"])
}
