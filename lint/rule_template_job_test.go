package lint

import (
	"testing"
)

func TestJobTemplate(t *testing.T) {
	linter := NewTemplateJobRule()

	for _, tc := range []struct {
		name      string
		result    []Result
		dashboard Dashboard
	}{
		{
			name:   "Non-promtheus dashboards shouldn't fail.",
			result: []Result{ResultSuccess},
			dashboard: Dashboard{
				Title: "test",
			},
		},
		{
			name: "Missing job template.",
			result: []Result{{
				Severity: Error,
				Message:  "Dashboard 'test' is missing the job template",
			}},
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
					},
				},
			},
		},
		{
			name: "Wrong datasource.",
			result: []Result{
				{Severity: Error, Message: "Dashboard 'test' job template should use datasource '$datasource', is currently 'foo'"},
				{Severity: Error, Message: "Dashboard 'test' job template should be a Prometheus query, is currently ''"},
				{Severity: Warning, Message: "Dashboard 'test' job template should be a labeled 'Job', is currently ''"},
				{Severity: Error, Message: "Dashboard 'test' job template should be a multi select"},
				{Severity: Error, Message: "Dashboard 'test' job template allValue should be '.+', is currently ''"}},
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
							Datasource: "foo",
						},
					},
				},
			},
		},
		{
			name: "Wrong type.",
			result: []Result{
				{Severity: Error, Message: "Dashboard 'test' job template should be a Prometheus query, is currently 'bar'"},
				{Severity: Warning, Message: "Dashboard 'test' job template should be a labeled 'Job', is currently ''"},
				{Severity: Error, Message: "Dashboard 'test' job template should be a multi select"},
				{Severity: Error, Message: "Dashboard 'test' job template allValue should be '.+', is currently ''"}},
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
							Type:       "bar",
						},
					},
				},
			},
		},
		{
			name: "Wrong job label.",
			result: []Result{
				{Severity: Warning, Message: "Dashboard 'test' job template should be a labeled 'Job', is currently 'bar'"},
				{Severity: Error, Message: "Dashboard 'test' job template should be a multi select"},
				{Severity: Error, Message: "Dashboard 'test' job template allValue should be '.+', is currently ''"}},
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
							Label:      "bar",
						},
					},
				},
			},
		},
		{
			name:   "OK",
			result: []Result{ResultSuccess},
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
		t.Run(tc.name, func(t *testing.T) {
			testMultiResultRule(t, linter, tc.dashboard, tc.result)
		})
	}
}
