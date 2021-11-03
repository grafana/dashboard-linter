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

		r := &ResultSet{}
		r.AddResult(ResultContext{
			Result:      linter.LintIntegration(integration),
			Integration: integration,
			Rule:        linter,
		})

		require.Len(t, r.results, 1)
		require.Equal(t, r.results[0].Result.Severity, Success)
	})
	t.Run("warning when supported_platforms empty but invisible", func(t *testing.T) {
		linter := NewMetaSupportedPlatformsRule()
		integration := &integrations.Integration{
			Meta: &integrations.Metadata{
				Visible: false,
			},
		}

		r := &ResultSet{}
		r.AddResult(ResultContext{
			Result:      linter.LintIntegration(integration),
			Integration: integration,
			Rule:        linter,
		})

		require.Len(t, r.results, 1)
		require.Equal(t, r.results[0].Result.Severity, Warning)
	})
	t.Run("error when supported_platforms empty and visible", func(t *testing.T) {
		linter := NewMetaSupportedPlatformsRule()
		integration := &integrations.Integration{
			Meta: &integrations.Metadata{
				Visible: true,
			},
		}

		r := &ResultSet{}
		r.AddResult(ResultContext{
			Result:      linter.LintIntegration(integration),
			Integration: integration,
			Rule:        linter,
		})

		require.Len(t, r.results, 1)
		require.Equal(t, r.results[0].Result.Severity, Error)
	})
}
