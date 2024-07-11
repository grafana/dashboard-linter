package lint

import (
	"testing"
)

func TestTemplateLabelPromQLRule(t *testing.T) {
	linter := NewTemplateLabelPromQLRule()

	for _, tc := range []struct {
		name      string
		result    Result
		dashboard Dashboard
	}{
		{
			name:   "Don't fail on non prometheus template.",
			result: ResultSuccess,
			dashboard: Dashboard{
				Title: "test",
				Templating: struct {
					List []Template `json:"list"`
				}{
					List: []Template{
						{
							Type:  "datasource",
							Query: "foo",
						},
					},
				},
			},
		},
		{
			name:   "OK",
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
							Name:       "namespaces",
							Datasource: "$datasource",
							Query:      "label_values(up{job=~\"$job\"}, namespace)",
							Type:       "query",
							Label:      "job",
						},
					},
				},
			},
		},
		{
			name: "Error",
			result: Result{
				Severity: Error,
				Message:  `Dashboard 'test' template 'namespaces' invalid templated label 'label_values(up{, namespace)': 1:4: parse error: unexpected "," in label matching, expected identifier or "}"`,
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
						},
					},
				},
			},
		},
		{
			name: "Invalid function.",
			result: Result{
				Severity: Error,
				Message:  `Dashboard 'test' template 'namespaces' invalid templated label 'foo(up, namespace)': invalid 'function': foo`,
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
							Query:      "foo(up, namespace)",
							Type:       "query",
							Label:      "job",
						},
					},
				},
			},
		},
		{
			name: "Invalid query expression.",
			result: Result{
				Severity: Error,
				Message:  `Dashboard 'test' template 'namespaces' invalid templated label 'foo': invalid 'query': foo`,
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
							Query:      "foo",
							Type:       "query",
							Label:      "job",
						},
					},
				},
			},
		},
		// Support main grafana variables.
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
							Name:       "namespaces",
							Datasource: "$datasource",
							Query:      "query_result(max by(namespaces) (max_over_time(memory{}[$__range])))",
							Type:       "query",
							Label:      "job",
						},
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			testRule(t, linter, tc.dashboard, tc.result)
		})
	}
}
