package lint

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A minimal kubernetes "v2" dashboard exercising the parts the adapter maps:
// a panel (with type, unit, a prometheus query and datasource), a datasource
// variable, a query variable, and an annotation query.
const v2Dashboard = `{
	"apiVersion": "dashboard.grafana.app/v2",
	"kind": "Dashboard",
	"spec": {
		"title": "V2 Test",
		"editable": true,
		"variables": [
			{
				"kind": "DatasourceVariable",
				"spec": { "name": "datasource", "label": "Data source", "pluginId": "prometheus" }
			},
			{
				"kind": "QueryVariable",
				"spec": {
					"name": "cluster", "label": "cluster", "multi": true,
					"refresh": "onTimeRangeChanged", "allValue": "all-clusters",
					"query": {
						"kind": "DataQuery", "group": "prometheus", "version": "v0",
						"datasource": { "name": "${datasource}" },
						"spec": { "query": "label_values(up, cluster)" }
					}
				}
			}
		],
		"annotations": [
			{ "kind": "AnnotationQuery", "spec": {
				"name": "Annotations & Alerts",
				"query": { "kind": "DataQuery", "group": "grafana", "version": "v0", "datasource": { "name": "-- Grafana --" }, "spec": {} }
			}}
		],
		"elements": {
			"panel-1": {
				"kind": "Panel",
				"spec": {
					"id": 1,
					"title": "CPU",
					"description": "cpu usage",
					"data": { "kind": "QueryGroup", "spec": { "queries": [
						{ "kind": "PanelQuery", "spec": {
							"refId": "A", "hidden": false,
							"query": {
								"kind": "DataQuery", "group": "prometheus", "version": "v0",
								"datasource": { "name": "$datasource" },
								"spec": { "expr": "sum(rate(node_cpu_seconds_total{cluster=\"$cluster\"}[5m]))" }
							}
						}}
					] } },
					"vizConfig": {
						"kind": "VizConfig", "group": "timeseries", "version": "1.0",
						"spec": { "options": {}, "fieldConfig": { "defaults": { "unit": "percent" }, "overrides": [] } }
					}
				}
			}
		}
	}
}`

func TestParseV2Dashboard(t *testing.T) {
	d, err := NewDashboard([]byte(v2Dashboard))
	require.NoError(t, err)

	assert.Equal(t, "V2 Test", d.Title)
	assert.Equal(t, "dashboard.grafana.app/v2", d.APIVersion)
	assert.True(t, d.Editable)

	t.Run("panel", func(t *testing.T) {
		panels := d.GetPanels()
		require.Len(t, panels, 1)
		p := panels[0]
		assert.Equal(t, 1, p.Id)
		assert.Equal(t, "CPU", p.Title)
		assert.Equal(t, "cpu usage", p.Description)
		assert.Equal(t, "timeseries", p.Type)
		require.NotNil(t, p.FieldConfig)
		assert.Equal(t, "percent", p.FieldConfig.Defaults.Unit)

		// panel datasource is derived from the first query and resolves to the
		// templated reference the panel-datasource rule expects.
		src, err := p.GetDataSource()
		require.NoError(t, err)
		assert.Equal(t, "$datasource", src.UID)
	})

	t.Run("target", func(t *testing.T) {
		p := d.GetPanels()[0]
		require.Len(t, p.Targets, 1)
		tg := p.Targets[0]
		assert.Equal(t, "A", tg.RefId)
		assert.False(t, tg.Hide)
		assert.Equal(t, `sum(rate(node_cpu_seconds_total{cluster="$cluster"}[5m]))`, tg.Expr)
		src, err := tg.GetDataSource()
		require.NoError(t, err)
		assert.Equal(t, "$datasource", src.UID)
		assert.Equal(t, "prometheus", src.Type)
	})

	t.Run("variables", func(t *testing.T) {
		dsVars := d.GetTemplateByType("datasource")
		require.Len(t, dsVars, 1)
		assert.Equal(t, "datasource", dsVars[0].Name)
		// PluginId becomes Query so prometheus/loki detection works.
		assert.Equal(t, Prometheus, dsVars[0].Query)

		queryVars := d.GetTemplateByType("query")
		require.Len(t, queryVars, 1)
		assert.Equal(t, "cluster", queryVars[0].Name)
		assert.True(t, queryVars[0].Multi)
		assert.Equal(t, "label_values(up, cluster)", queryVars[0].Query)
		// "onTimeRangeChanged" maps to the classic refresh value 2.
		assert.Equal(t, 2, queryVars[0].Refresh)
		// AllValue is copied through verbatim (distinctive value catches a
		// wrong-field mapping, e.g. Regex -> AllValue).
		assert.Equal(t, "all-clusters", queryVars[0].AllValue)
	})

	t.Run("annotations", func(t *testing.T) {
		require.Len(t, d.Annotations.List, 1)
		assert.Equal(t, "Annotations & Alerts", d.Annotations.List[0].Name)
	})
}

// ruleHasError reports whether the named rule produced any Error-severity result.
func ruleHasError(results *ResultSet, rule string) bool {
	for _, rc := range results.ByRule()[rule] {
		for _, r := range rc.Result.Results {
			if r.Severity == Error {
				return true
			}
		}
	}
	return false
}

// TestLintV2Dashboard exercises the full rule set against a v2 dashboard to verify
// the adapter feeds each mapped field into the rules end-to-end.
func TestLintV2Dashboard(t *testing.T) {
	d, err := NewDashboard([]byte(v2Dashboard))
	require.NoError(t, err)
	rs := NewRuleSet()
	results, err := rs.Lint([]Dashboard{d})
	require.NoError(t, err)

	// Each rule here passes only if a specific adapter mapping is correct, so a
	// regression turns one into a false-positive lint error.
	t.Run("adapter mappings produce no false positives", func(t *testing.T) {
		mustPass := map[string]string{
			"template-on-time-change-reload-rule": "Refresh enum -> int",
			"template-datasource-rule":            "datasource variable (pluginId -> Query)",
			"panel-datasource-rule":               "per-query datasource name -> panel uid",
			"panel-units-rule":                    "vizConfig fieldConfig unit round-trip",
			"panel-title-description-rule":        "panel title/description",
			"panel-no-targets-rule":               "panel queries -> targets",
			"target-promql-rule":                  "query expr extraction (valid, non-empty)",
		}
		for rule, mapped := range mustPass {
			assert.Falsef(t, ruleHasError(results, rule),
				"%s should pass (validates: %s)", rule, mapped)
		}
	})

	// Behavioral round-trip: the mapped Refresh value must actually be consumed by
	// the rule. Flip the fixture's query var to onDashboardLoad (-> 1) and the
	// on-time-range rule must now fire, proving the mapping distinguishes pass/fail.
	t.Run("mapped Refresh is consumed by the on-time-range rule", func(t *testing.T) {
		load := strings.Replace(v2Dashboard, `"onTimeRangeChanged"`, `"onDashboardLoad"`, 1)
		d, err := NewDashboard([]byte(load))
		require.NoError(t, err)
		rs := NewRuleSet()
		results, err := rs.Lint([]Dashboard{d})
		require.NoError(t, err)
		assert.True(t, ruleHasError(results, "template-on-time-change-reload-rule"),
			"on-time-range rule should fire for an onDashboardLoad query variable")
	})
}
