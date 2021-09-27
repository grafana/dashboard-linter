package lint

import (
	"strconv"
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

func appendConfigExclude(t *testing.T, rule string, integration string, dashboard string, panel string, targetIdx string, config *Configuration) {
	t.Helper()
	var ce *ConfigurationEntry
	if dashboard != "" || panel != "" || targetIdx != "" {
		ce = &ConfigurationEntry{
			Dashboard: dashboard,
			Panel:     panel,
			TargetIdx: targetIdx,
		}
	}
	if _, ok := config.configs[integration]; !ok {
		config.configs[integration] = ConfigurationFile{
			Warnings: map[string]*ConfigurationRuleEntries{},
			Exclusions: map[string]*ConfigurationRuleEntries{
				rule: {},
			},
		}
	}
	if ce != nil {
		config.configs[integration].Exclusions[rule].AddEntry(*ce)
	}
}

func appendConfigWarning(t *testing.T, rule string, integration string, dashboard string, panel string, targetIdx string, config *Configuration) {
	t.Helper()
	var ce *ConfigurationEntry
	if dashboard != "" || panel != "" || targetIdx != "" {
		ce = &ConfigurationEntry{
			Dashboard: dashboard,
			Panel:     panel,
			TargetIdx: targetIdx,
		}
	}
	if _, ok := config.configs[integration]; !ok {
		config.configs[integration] = ConfigurationFile{
			Exclusions: map[string]*ConfigurationRuleEntries{},
			Warnings: map[string]*ConfigurationRuleEntries{
				rule: {},
			},
		}
	}
	if ce != nil {
		config.configs[integration].Warnings[rule].AddEntry(*ce)
	}
}

func newResultContext(t *testing.T, rule string, integration string, dashboard string, panel string, targetIdx string, result Severity) ResultContext {
	ret := ResultContext{
		Result: Result{Severity: result, Message: "foo"},
	}
	if rule != "" {
		ret.Rule = &TestRule{name: rule}
	}
	if integration != "" {
		ret.Integration = &integrations.Integration{Meta: &integrations.Metadata{Slug: integration}}
	}
	if dashboard != "" {
		ret.Dashboard = &Dashboard{Title: dashboard}
	}
	if panel != "" {
		ret.Panel = &Panel{Title: panel}
	}
	if targetIdx != "" {
		idx, err := strconv.Atoi(targetIdx)
		if err == nil {
			ret.Target = &Target{Idx: idx}
		}
	}
	return ret
}

func TestResultSet(t *testing.T) {
	t.Run("MaximumSeverity", func(t *testing.T) {
		r := ResultSet{
			results: []ResultContext{
				{Result: Result{Severity: Success}},
				{Result: Result{Severity: Warning}},
				{Result: Result{Severity: Error}},
			},
		}

		require.Equal(t, r.MaximumSeverity(), Error)
	})

	t.Run("ByRule", func(t *testing.T) {
		r := ResultSet{
			results: []ResultContext{
				newResultContext(t, "rule1", "Bintegration", "", "", "", Success),
				newResultContext(t, "rule1", "Aintegration", "", "", "", Success),
				newResultContext(t, "rule2", "Bintegration", "", "", "", Success),
				newResultContext(t, "rule2", "Aintegration", "", "", "", Success),
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
			results: []ResultContext{
				newResultContext(t, "", "Aintegration", "", "", "", Success),
				newResultContext(t, "", "Bintegration", "", "", "", Success),
			},
		}

		byIntegration := r.ByIntegration()

		require.Len(t, byIntegration, 2)
		require.Contains(t, byIntegration, "Aintegration")
		require.Contains(t, byIntegration, "Bintegration")
	})
}

func TestConfiguration(t *testing.T) {
	t.Run("Excludes Integration", func(t *testing.T) {
		c := NewConfiguration()
		appendConfigExclude(t, "rule1", "Aintegration", "", "", "", c)

		r1 := newResultContext(t, "rule1", "Aintegration", "", "", "", Error)
		r2 := newResultContext(t, "rule1", "Bintegration", "", "", "", Error)

		rc1 := c.Apply(r1)
		require.Equal(t, Exclude, rc1.Result.Severity)

		rc2 := c.Apply(r2)
		require.Equal(t, Error, rc2.Result.Severity)
	})

	t.Run("Warns Integration", func(t *testing.T) {
		c := NewConfiguration()
		appendConfigWarning(t, "rule1", "Aintegration", "", "", "", c)

		r1 := newResultContext(t, "rule1", "Aintegration", "", "", "", Error)
		r2 := newResultContext(t, "rule1", "Bintegration", "", "", "", Error)

		rc1 := c.Apply(r1)
		require.Equal(t, Warning, rc1.Result.Severity)

		rc2 := c.Apply(r2)
		require.Equal(t, Error, rc2.Result.Severity)
	})

	t.Run("Excludes More Specific Config", func(t *testing.T) {
		c := NewConfiguration()
		appendConfigExclude(t, "rule1", "Aintegration", "", "", "", c)
		appendConfigExclude(t, "rule1", "Aintegration", "dash1", "", "", c)

		r1 := newResultContext(t, "rule1", "Aintegration", "dash1", "", "", Error)
		r2 := newResultContext(t, "rule1", "Aintegration", "dash2", "", "", Error)

		rc1 := c.Apply(r1)
		require.Equal(t, Exclude, rc1.Result.Severity)

		rc2 := c.Apply(r2)
		require.Equal(t, Error, rc2.Result.Severity)
	})

	t.Run("Excludes all when rule defined but entries empty", func(t *testing.T) {
		c := NewConfiguration()
		appendConfigExclude(t, "rule1", "Aintegration", "", "", "", c)

		r1 := newResultContext(t, "rule1", "Aintegration", "dash1", "panel1", "0", Error)
		r2 := newResultContext(t, "rule1", "Aintegration", "dash1", "panel1", "1", Error)
		r3 := newResultContext(t, "rule1", "Aintegration", "dash2", "panel2", "0", Error)
		r4 := newResultContext(t, "rule1", "Aintegration", "dash2", "panel2", "1", Error)

		rs := []ResultContext{r1, r2, r3, r4}
		for _, r := range rs {
			rc := c.Apply(r)
			require.Equal(t, Exclude, rc.Result.Severity)
		}
	})

	// Dashboards
	t.Run("Excludes Dashboard", func(t *testing.T) {
		c := NewConfiguration()
		appendConfigExclude(t, "rule1", "Aintegration", "dash1", "", "", c)

		r1 := newResultContext(t, "rule1", "Aintegration", "dash1", "", "", Error)
		r2 := newResultContext(t, "rule1", "Aintegration", "dash2", "", "", Error)

		rc1 := c.Apply(r1)
		require.Equal(t, Exclude, rc1.Result.Severity)

		rc2 := c.Apply(r2)
		require.Equal(t, Error, rc2.Result.Severity)
	})

	t.Run("Warns Dashboard", func(t *testing.T) {
		c := NewConfiguration()
		appendConfigWarning(t, "rule1", "Aintegration", "dash1", "", "", c)

		r1 := newResultContext(t, "rule1", "Aintegration", "dash1", "", "", Error)
		r2 := newResultContext(t, "rule1", "Aintegration", "dash2", "", "", Error)

		rc1 := c.Apply(r1)
		require.Equal(t, Warning, rc1.Result.Severity)

		rc2 := c.Apply(r2)
		require.Equal(t, Error, rc2.Result.Severity)
	})

	// Panels
	t.Run("Excludes Panels", func(t *testing.T) {
		c := NewConfiguration()
		appendConfigExclude(t, "rule1", "Aintegration", "dash1", "panel1", "", c)

		r1 := newResultContext(t, "rule1", "Aintegration", "dash1", "panel1", "", Error)
		r2 := newResultContext(t, "rule1", "Aintegration", "dash1", "panel2", "", Error)

		rc1 := c.Apply(r1)
		require.Equal(t, Exclude, rc1.Result.Severity)

		rc2 := c.Apply(r2)
		require.Equal(t, Error, rc2.Result.Severity)
	})

	t.Run("Warns Panels", func(t *testing.T) {
		c := NewConfiguration()
		appendConfigWarning(t, "rule1", "Aintegration", "dash1", "panel1", "", c)

		r1 := newResultContext(t, "rule1", "Aintegration", "dash1", "panel1", "", Error)
		r2 := newResultContext(t, "rule1", "Aintegration", "dash1", "panel2", "", Error)

		rc1 := c.Apply(r1)
		require.Equal(t, Warning, rc1.Result.Severity)

		rc2 := c.Apply(r2)
		require.Equal(t, Error, rc2.Result.Severity)
	})

	// Targets
	t.Run("Excludes Targets", func(t *testing.T) {
		c := NewConfiguration()
		appendConfigExclude(t, "rule1", "Aintegration", "dash1", "panel1", "0", c)

		r1 := newResultContext(t, "rule1", "Aintegration", "dash1", "panel1", "0", Error)
		r2 := newResultContext(t, "rule1", "Aintegration", "dash1", "panel1", "1", Error)

		rc1 := c.Apply(r1)
		require.Equal(t, Exclude, rc1.Result.Severity)

		rc2 := c.Apply(r2)
		require.Equal(t, Error, rc2.Result.Severity)
	})

	t.Run("Warns Targets", func(t *testing.T) {
		c := NewConfiguration()
		appendConfigWarning(t, "rule1", "Aintegration", "dash1", "panel1", "0", c)

		r1 := newResultContext(t, "rule1", "Aintegration", "dash1", "panel1", "0", Error)
		r2 := newResultContext(t, "rule1", "Aintegration", "dash1", "panel1", "1", Error)

		rc1 := c.Apply(r1)
		require.Equal(t, Warning, rc1.Result.Severity)

		rc2 := c.Apply(r2)
		require.Equal(t, Error, rc2.Result.Severity)
	})
}
