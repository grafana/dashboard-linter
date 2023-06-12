package lint

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTemplateOnTimeRangeReloadRule(t *testing.T) {
	linter := NewTemplateOnTimeRangeReloadRule()

	good := []Template{
		{
			Type:  "datasource",
			Query: "prometheus",
		},
		{
			Name:       "namespaces",
			Datasource: "$datasource",
			Query:      "label_values(up{job=~\"$job\"}, namespace)",
			Type:       "query",
			Label:      "job",
			Refresh:    2,
		},
	}
	for _, tc := range []struct {
		name      string
		result    Result
		dashboard Dashboard
		fixed     *Dashboard
	}{
		{
			name:   "OK",
			result: ResultSuccess,
			dashboard: Dashboard{
				Title: "test",
				Templating: struct {
					List []Template `json:"list"`
				}{
					List: good,
				},
			},
		},
		{
			name: "autofix",
			result: Result{
				Severity: Fixed,
				Message:  `Dashboard 'test' templated datasource variable named 'namespaces', should be set to be refreshed 'On Time Range Change (value 2)', is currently '1'`,
			},
			dashboard: Dashboard{
				Title: "test",
				Templating: struct {
					List []Template `json:"list"`
				}{
					List: ([]Template{
						{
							Type:  "datasource",
							Query: "prometheus",
						},
						{
							Name:       "namespaces",
							Datasource: "$datasource",
							Query:      "label_values(up{job=~\"$job\"}, namespace)",
							Type:       "query",
							Label:      "job",
							Refresh:    1,
						},
					}),
				},
			},
			fixed: &Dashboard{
				Title: "test",
				Templating: struct {
					List []Template `json:"list"`
				}{
					List: good,
				},
			},
		},
		{
			name: "error",
			result: Result{
				Severity: Error,
				Message:  `Dashboard 'test' templated datasource variable named 'namespaces', should be set to be refreshed 'On Time Range Change (value 2)', is currently '1'`,
			},
			dashboard: Dashboard{
				Title: "test",
				Templating: struct {
					List []Template `json:"list"`
				}{
					List: ([]Template{
						{
							Type:  "datasource",
							Query: "prometheus",
						},
						{
							Name:       "namespaces",
							Datasource: "$datasource",
							Query:      "label_values(up{job=~\"$job\"}, namespace)",
							Type:       "query",
							Label:      "job",
							Refresh:    1,
						},
					}),
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			autofix := tc.fixed != nil
			testRuleWithAutofix(t, linter, &tc.dashboard, []Result{tc.result}, autofix)
			if autofix {
				expected, _ := json.Marshal(tc.fixed)
				actual, _ := json.Marshal(tc.dashboard)
				require.Equal(t, string(expected), string(actual))
			}
		})
	}
}
