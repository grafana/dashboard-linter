package lint

import "testing"

func TestInstanceTemplate(t *testing.T) {
	linter := NewTemplateInstanceRule()

	for _, tc := range []struct {
		result    Result
		dashboard Dashboard
	}{
		// Non-promtheus dashboards shouldn't fail.
		{
			result: ResultSuccess,
			dashboard: Dashboard{
				Title: "test",
			},
		},
		// Missing instance templates.
		{
			result: Result{
				Severity: Error,
				Message:  "Dashboard 'test' is missing the instance template",
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
							Name:       "job",
							Datasource: "$datasource",
							Type:       "query",
							Label:      "Job",
							Multi:      true,
							AllValue:   ".+",
						},
					},
				},
			},
		},
		// What success looks like.
		{
			result: ResultSuccess,
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
							Name:       "job",
							Datasource: "$datasource",
							Type:       "query",
							Label:      "Job",
							Multi:      true,
							AllValue:   ".+",
						},
						{
							Name:       "instance",
							Datasource: "${datasource}",
							Type:       "query",
							Label:      "Instance",
							Multi:      true,
							AllValue:   ".+",
						},
					},
				},
			},
		},
	} {
		testRule(t, linter, tc.dashboard, tc.result)
	}
}
