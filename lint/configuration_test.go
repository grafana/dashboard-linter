package lint

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigurationFileLoadFromYAML(t *testing.T) {
	yaml := `
exclusions:
  target-rate-interval-rule:
    reason: "интервал не критичен"
    entries:
      - dashboard: "Sample dashboard"
        panel: "Timeseries"
        targetIdx: "0"
warnings:
  panel-title-description-rule:
    entries:
      - dashboard: "Sample dashboard"
`
	dir := t.TempDir()
	configPath := filepath.Join(dir, ".lint")
	require.NoError(t, os.WriteFile(configPath, []byte(yaml), 0644))

	cf := NewConfigurationFile()
	require.NoError(t, cf.Load(configPath))

	require.Contains(t, cf.Exclusions, "target-rate-interval-rule")
	entries := cf.Exclusions["target-rate-interval-rule"].Entries
	require.Len(t, entries, 1)
	require.Equal(t, "Timeseries", entries[0].Panel)
	require.Equal(t, "0", entries[0].TargetIdx)
	require.Equal(t, "Sample dashboard", entries[0].Dashboard)

	require.Contains(t, cf.Warnings, "panel-title-description-rule")
}

func TestConfigurationFileLoadMissingFile(t *testing.T) {
	cf := NewConfigurationFile()
	require.NoError(t, cf.Load(filepath.Join(t.TempDir(), "does-not-exist")))
}

func TestConfigurationEntryIsMatchTargetIdx(t *testing.T) {
	ce := ConfigurationEntry{TargetIdx: "0"}

	matching := ResultContext{
		Dashboard: &Dashboard{Title: "dash"},
		Panel:     &Panel{Title: "panel"},
		Target:    &Target{Idx: 0},
	}
	require.True(t, ce.IsMatch(matching))

	nonMatching := ResultContext{
		Dashboard: &Dashboard{Title: "dash"},
		Panel:     &Panel{Title: "panel"},
		Target:    &Target{Idx: 1},
	}
	require.False(t, ce.IsMatch(nonMatching))
}
