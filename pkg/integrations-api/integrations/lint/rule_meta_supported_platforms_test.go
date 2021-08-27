package lint

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafana/cloud-onboarding/pkg/integrations-api/integrations"
)

func TestMetaSupportedPlatforms(t *testing.T) {
	t.Run("success when supported_platforms exists and has values", func(t *testing.T) {
		linter := NewMetaSupportedPlatformsRule()
		integration := &integrations.Integration{
			Meta: &integrations.Metadata{
				Visible:            true,
				SupportedPlatforms: []string{"atleast one"},
			},
		}

		r := linter.Lint(integration)

		require.Len(t, r, 1)
		require.Equal(t, r[0].Severity, Success)
	})
	t.Run("warning when supported_platforms empty but invisible", func(t *testing.T) {
		linter := NewMetaSupportedPlatformsRule()
		integration := &integrations.Integration{
			Meta: &integrations.Metadata{
				Visible: false,
			},
		}

		r := linter.Lint(integration)

		require.Len(t, r, 1)
		require.Equal(t, r[0].Severity, Warning)
	})
	t.Run("error when supported_platforms empty and visible", func(t *testing.T) {
		linter := NewMetaSupportedPlatformsRule()
		integration := &integrations.Integration{
			Meta: &integrations.Metadata{
				Visible: true,
			},
		}

		r := linter.Lint(integration)

		require.Len(t, r, 1)
		require.Equal(t, r[0].Severity, Error)
	})
}
