package lint

import (
	"testing"
)

func TestTemplateOnTimeRangeReloadRule(t *testing.T) {
	linter := NewTemplateOnTimeRangeReloadRule()

	for _, tc := range []struct {
		result    Result
		dashboard Dashboard
	}{
		// What success looks like.
		{
			result: Result{
				Severity: Success,
				Message:  "OK",
			},
			dashboard: Dashboard{
				Title: "test",
				Templating: struct {
					List []Template `json:"list"`
				}{
					List: []Template{
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
					},
				},
			},
		},
		// What failure looks like.
		{
			result: Result{
				Severity: Error,
				Message:  `Dashboard 'test' templated datasource variable named 'namespaces', should be set to be refreshed 'On Time Range Change (value 2)', is currently '1'`,
			},
			dashboard: Dashboard{
				Title: "test",
				Templating: struct {
					List []Template `json:"list"`
				}{
					List: []Template{
						{
							Type:  "datasource",
							Query: "prometheus",
						},
						{
							Name:       "namespaces",
							Datasource: "$datasource",
							Query:      "label_values(up{, namespace)",
							Type:       "query",
							Label:      "job",
							Refresh:    1,
						},
					},
				},
			},
		},
	} {
		testRule(t, linter, tc.dashboard, tc.result)
	}
}
