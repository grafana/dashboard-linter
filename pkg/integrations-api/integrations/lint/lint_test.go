package lint

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafana/cloud-onboarding/pkg/integrations-api/integrations"
)

type TestRule struct {
	name string
}

func (r *TestRule) Description() string {
	return "Test Rule"
}

func (r *TestRule) Name() string {
	return r.name
}

func (r *TestRule) Lint(i *integrations.Integration) []Result {
	return make([]Result, 0)
}

func TestResultSet(t *testing.T) {
	t.Run("MaximumSeverity", func(t *testing.T) {
		r := ResultSet{
			results: []Result{
				{Severity: Success},
				{Severity: Warning},
				{Severity: Error},
			},
		}

		require.Equal(t, r.MaximumSeverity(), Error)
	})

	t.Run("ByRule", func(t *testing.T) {
		r := ResultSet{
			results: []Result{
				{
					Severity:    Success,
					Rule:        &TestRule{name: "rule1"},
					Integration: &integrations.Integration{Meta: &integrations.Metadata{Slug: "Bintegration"}},
				},
				{
					Severity:    Success,
					Rule:        &TestRule{name: "rule1"},
					Integration: &integrations.Integration{Meta: &integrations.Metadata{Slug: "Aintegration"}},
				},
				{
					Severity:    Success,
					Rule:        &TestRule{name: "rule2"},
					Integration: &integrations.Integration{Meta: &integrations.Metadata{Slug: "Bintegration"}},
				},
				{
					Severity:    Success,
					Rule:        &TestRule{name: "rule2"},
					Integration: &integrations.Integration{Meta: &integrations.Metadata{Slug: "Aintegration"}},
				},
			},
		}

		byRule := r.ByRule()

		require.Len(t, byRule, 2)
		require.Contains(t, byRule, "rule1")
		require.Contains(t, byRule, "rule2")
		require.Len(t, byRule["rule1"], 2)
		require.Len(t, byRule["rule2"], 2)
		require.Equal(t, "Aintegration", byRule["rule1"][0].Integration.Meta.Slug)
	})

	t.Run("ByIntegration", func(t *testing.T) {
		r := ResultSet{
			results: []Result{
				{
					Severity:    Success,
					Integration: &integrations.Integration{Meta: &integrations.Metadata{Slug: "Aintegration"}},
				},
				{
					Severity:    Success,
					Integration: &integrations.Integration{Meta: &integrations.Metadata{Slug: "Bintegration"}},
				},
			},
		}

		byIntegration := r.ByIntegration()

		require.Len(t, byIntegration, 2)
		require.Contains(t, byIntegration, "Aintegration")
		require.Contains(t, byIntegration, "Bintegration")
	})
}
